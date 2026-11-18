package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type emailJob struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CompanyID *uuid.UUID
	Kind      string
	Sequence  int16
}

func (s *Service) CancelProviderReminders(ctx context.Context, companyID uuid.UUID) error {
	_, err := s.DB.Exec(ctx, `
		UPDATE email_jobs
		SET status = 'cancelled'
		WHERE company_id = $1 AND kind = 'provider_onboarding' AND status = 'pending'`, companyID)
	if err != nil {
		return fmt.Errorf("cancel provider onboarding reminders: %w", err)
	}
	return nil
}

func (s *Service) RunEmailQueue(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		for {
			job, err := s.claimEmailJob(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				s.Log.Error("could not claim scheduled email", "error", err)
				break
			}
			if job == nil {
				break
			}
			s.deliverEmailJob(ctx, *job)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) claimEmailJob(ctx context.Context) (*emailJob, error) {
	var job *emailJob
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var found emailJob
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, company_id, kind, sequence
			FROM email_jobs
			WHERE (status = 'pending' AND scheduled_at <= now())
			   OR (status = 'sending' AND claimed_at < now() - interval '5 minutes')
			ORDER BY scheduled_at, created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED`).Scan(&found.ID, &found.UserID, &found.CompanyID, &found.Kind, &found.Sequence)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE email_jobs SET status = 'sending', claimed_at = now() WHERE id = $1`, found.ID); err != nil {
			return err
		}
		job = &found
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("claim email job: %w", err)
	}
	return job, nil
}

func (s *Service) deliverEmailJob(ctx context.Context, job emailJob) {
	var email, fullName, companyName string
	err := s.DB.QueryRow(ctx, `
		SELECT u.email, u.full_name, COALESCE(c.name, '')
		FROM users u
		LEFT JOIN companies c ON c.id = $2
		WHERE u.id = $1 AND u.deleted_at IS NULL`, job.UserID, job.CompanyID).Scan(&email, &fullName, &companyName)
	if errors.Is(err, pgx.ErrNoRows) {
		s.cancelEmailJob(ctx, job.ID)
		return
	}
	if err != nil {
		s.retryEmailJob(ctx, job, err)
		return
	}

	var sendErr error
	switch job.Kind {
	case "welcome_customer":
		sendErr = s.Mailer.SendWelcome(ctx, email, fullName, false)
	case "welcome_provider":
		sendErr = s.Mailer.SendWelcome(ctx, email, fullName, true)
	case "provider_onboarding":
		var incomplete bool
		err = s.DB.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM companies
				WHERE id = $1 AND deleted_at IS NULL
				  AND status NOT IN ('under_review', 'verified')
				  AND onboarding_status = 'incomplete'
			)`, job.CompanyID).Scan(&incomplete)
		if err != nil {
			s.retryEmailJob(ctx, job, err)
			return
		}
		if !incomplete {
			s.cancelEmailJob(ctx, job.ID)
			return
		}
		sendErr = s.Mailer.SendOnboardingReminder(ctx, email, fullName, companyName, int(job.Sequence))
	default:
		sendErr = fmt.Errorf("unsupported scheduled email kind %q", job.Kind)
	}

	if sendErr != nil {
		s.retryEmailJob(ctx, job, sendErr)
		return
	}
	s.finishEmailJob(ctx, job.ID)
}

func (s *Service) finishEmailJob(ctx context.Context, id uuid.UUID) {
	if _, err := s.DB.Exec(ctx, `
		UPDATE email_jobs SET status = 'sent', sent_at = now(), claimed_at = NULL
		WHERE id = $1`, id); err != nil {
		s.Log.Error("could not complete scheduled email", "job_id", id, "error", err)
	}
}

func (s *Service) cancelEmailJob(ctx context.Context, id uuid.UUID) {
	if _, err := s.DB.Exec(ctx, `
		UPDATE email_jobs SET status = 'cancelled', claimed_at = NULL
		WHERE id = $1`, id); err != nil {
		s.Log.Error("could not cancel scheduled email", "job_id", id, "error", err)
	}
}

func (s *Service) retryEmailJob(ctx context.Context, job emailJob, cause error) {
	if _, err := s.DB.Exec(ctx, `
		UPDATE email_jobs
		SET attempts = attempts + 1,
		    status = CASE WHEN attempts >= 7 THEN 'failed' ELSE 'pending' END,
		    scheduled_at = now() + make_interval(mins => LEAST(60, power(2, attempts)::int)),
		    claimed_at = NULL,
		    last_error = left($2, 1000)
		WHERE id = $1`, job.ID, cause.Error()); err != nil {
		s.Log.Error("could not reschedule failed email", "job_id", job.ID, "error", err)
		return
	}
	s.Log.Warn("scheduled email delivery failed", "job_id", job.ID, "error", cause)
}
