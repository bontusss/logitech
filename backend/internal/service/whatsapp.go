package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bol-platform/backend/internal/auth"
)

var (
	ErrWhatsAppUnavailable = errors.New("service: WhatsApp verification is not configured")
	ErrPhoneInvalid        = errors.New("service: phone number must use international E.164 format")
	ErrPhoneCodeCooldown   = errors.New("service: wait before requesting another WhatsApp code")
	ErrPhoneRateLimited    = errors.New("service: too many WhatsApp verification codes requested")
	ErrPhoneCodeInvalid    = errors.New("service: WhatsApp verification code is invalid")
	ErrPhoneCodeExpired    = errors.New("service: WhatsApp verification code expired")
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

type WhatsAppSender interface {
	SendVerificationCode(ctx context.Context, phone, code string) error
}

func normalizePhone(phone string) string {
	phone = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(phone))
	if !e164Pattern.MatchString(phone) {
		return ""
	}
	return phone
}

func (s *Service) IsWhatsAppVerified(ctx context.Context, userID uuid.UUID) (bool, error) {
	var verified bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_phone_verifications v
			JOIN users u ON u.id = v.user_id AND u.phone = v.phone
			WHERE v.user_id = $1 AND v.verified_at IS NOT NULL
		)`, userID).Scan(&verified)
	if err != nil {
		return false, fmt.Errorf("check verified WhatsApp number: %w", err)
	}
	return verified, nil
}

func (s *Service) SendWhatsAppCode(ctx context.Context, userID uuid.UUID, rawPhone string) (bool, error) {
	if s.WhatsApp == nil {
		return false, ErrWhatsAppUnavailable
	}
	phone := normalizePhone(rawPhone)
	if phone == "" {
		return false, ErrPhoneInvalid
	}

	codeNumber, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return false, fmt.Errorf("generate WhatsApp verification code: %w", err)
	}
	code := fmt.Sprintf("%06d", codeNumber.Int64())
	hash := auth.HashToken(code)
	now := time.Now()
	alreadyVerified := false

	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), 0)`, userID.String()); err != nil {
			return err
		}
		var currentPhone, userPhone string
		var lastSentAt, windowStartedAt time.Time
		var sends int16
		var verifiedAt *time.Time
		err := tx.QueryRow(ctx, `
			SELECT v.phone, COALESCE(u.phone, ''), v.last_sent_at, v.window_started_at,
			       v.sends_in_window, v.verified_at
			FROM user_phone_verifications v
			JOIN users u ON u.id = v.user_id
			WHERE v.user_id = $1 FOR UPDATE OF v`, userID,
		).Scan(&currentPhone, &userPhone, &lastSentAt, &windowStartedAt, &sends, &verifiedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = tx.Exec(ctx, `
				INSERT INTO user_phone_verifications
					(user_id, phone, token_hash, expires_at, last_sent_at, window_started_at)
				VALUES ($1, $2, $3, $4, $5, $5)`,
				userID, phone, hash, now.Add(10*time.Minute), now)
			return err
		}
		if err != nil {
			return err
		}
		if currentPhone == phone && userPhone == phone && verifiedAt != nil {
			alreadyVerified = true
			return nil
		}
		if now.Sub(lastSentAt) < time.Minute {
			return ErrPhoneCodeCooldown
		}
		if now.Sub(windowStartedAt) < time.Hour && sends >= 5 {
			return ErrPhoneRateLimited
		}

		windowReset := now.Sub(windowStartedAt) >= time.Hour
		if windowReset {
			windowStartedAt = now
			sends = 0
		}
		_, err = tx.Exec(ctx, `
			UPDATE user_phone_verifications
			SET phone = $2, token_hash = $3, expires_at = $4, verified_at = NULL,
			    last_sent_at = $5, window_started_at = $6, sends_in_window = $7, attempts = 0
			WHERE user_id = $1`,
			userID, phone, hash, now.Add(10*time.Minute), now, windowStartedAt, sends+1)
		return err
	})
	if err != nil {
		return false, err
	}
	if alreadyVerified {
		return true, nil
	}

	if err := s.WhatsApp.SendVerificationCode(ctx, phone, code); err != nil {
		_, cleanupErr := s.DB.Exec(ctx, `
			UPDATE user_phone_verifications SET token_hash = '', expires_at = now()
			WHERE user_id = $1 AND token_hash = $2`, userID, hash)
		if cleanupErr != nil {
			s.Log.Error("could not invalidate unsent WhatsApp code", "user_id", userID, "error", cleanupErr)
		}
		return false, fmt.Errorf("send WhatsApp verification code: %w", err)
	}
	return false, nil
}

func (s *Service) VerifyWhatsAppCode(ctx context.Context, userID uuid.UUID, code string) error {
	if len(code) != 6 {
		return ErrPhoneCodeInvalid
	}

	var verificationErr error
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var phone, tokenHash string
		var expiresAt time.Time
		var attempts int16
		err := tx.QueryRow(ctx, `
			SELECT phone, token_hash, expires_at, attempts
			FROM user_phone_verifications WHERE user_id = $1 FOR UPDATE`, userID,
		).Scan(&phone, &tokenHash, &expiresAt, &attempts)
		if errors.Is(err, pgx.ErrNoRows) {
			verificationErr = ErrPhoneCodeInvalid
			return nil
		}
		if err != nil {
			return err
		}
		if !time.Now().Before(expiresAt) || tokenHash == "" {
			verificationErr = ErrPhoneCodeExpired
			return nil
		}
		if attempts >= 5 {
			verificationErr = ErrPhoneCodeInvalid
			return nil
		}
		if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(auth.HashToken(code))) != 1 {
			_, updateErr := tx.Exec(ctx, `
				UPDATE user_phone_verifications
				SET attempts = attempts + 1,
				    token_hash = CASE WHEN attempts >= 4 THEN '' ELSE token_hash END
				WHERE user_id = $1`, userID)
			if updateErr != nil {
				return updateErr
			}
			verificationErr = ErrPhoneCodeInvalid
			return nil
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET phone = $2, updated_at = now() WHERE id = $1`, userID, phone); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			UPDATE user_phone_verifications
			SET verified_at = now(), token_hash = '', attempts = 0
			WHERE user_id = $1`, userID)
		return err
	})
	if err != nil {
		return fmt.Errorf("verify WhatsApp code: %w", err)
	}
	if verificationErr != nil {
		return verificationErr
	}
	return nil
}
