package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/mailer"
)

var (
	ErrEmailTaken              = errors.New("service: email already registered")
	ErrInvalidCredentials      = errors.New("service: invalid email or password")
	ErrInvalidRefresh          = errors.New("service: invalid refresh token")
	ErrInvalidToken            = errors.New("service: invalid or expired token")
	ErrUserSuspended           = errors.New("service: user account is suspended")
	ErrInvalidInvitation       = errors.New("service: invalid or expired invitation")
	ErrInvitationNeedsLogin    = errors.New("service: invitation requires the invited account to sign in")
	ErrInvitationEmailMismatch = errors.New("service: signed-in account does not match the invitation")
	ErrInvalidInvitationBranch = errors.New("service: invitation contains a branch outside the company")
	ErrInvalidPlatformRole     = errors.New("service: platform role must be admin or support")
)

type Service struct {
	DB       *pgxpool.Pool
	Q        *db.Queries
	Auth     *auth.Manager
	Redis    *redis.Client
	Cfg      *config.Config
	Mailer   mailer.Mailer
	WhatsApp WhatsAppSender
	Log      *slog.Logger
}

type Session struct {
	User                 db.User
	Companies            []db.ListUserMembershipsRow
	AccessToken          string
	AccessExpiresAt      time.Time
	RefreshToken         string
	RefreshExpiresAt     time.Time
	DevVerificationToken string
}

type CompanyRegistration struct {
	CompanyName   string
	BusinessEmail string
	Phone         string
	BusinessType  string
	Country       string
	State         string
	City          string
	Address       string
	ContactName   string
	Password      string
}

type AuditEntry struct {
	ActorID    *uuid.UUID
	CompanyID  *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	Metadata   map[string]any
	IP         string
	UserAgent  string
}

func New(pool *pgxpool.Pool, authManager *auth.Manager, redisClient *redis.Client, cfg *config.Config, m mailer.Mailer, log *slog.Logger) *Service {
	return &Service{
		DB:     pool,
		Q:      db.New(pool),
		Auth:   authManager,
		Redis:  redisClient,
		Cfg:    cfg,
		Mailer: m,
		Log:    log,
	}
}

// ---------------------------------------------------------------------------
// Customer registration and login
// ---------------------------------------------------------------------------

func (s *Service) RegisterCustomer(ctx context.Context, fullName, email, password, phone, ip, ua string) (Session, error) {
	email = normalizeEmail(email)
	hash, err := auth.HashPassword(password)
	if err != nil {
		return Session{}, err
	}

	var user db.User
	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var createErr error
		user, createErr = s.Q.WithTx(tx).CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			PasswordHash: hash, FullName: strings.TrimSpace(fullName),
			Column4:     strings.TrimSpace(phone),
			AccountType: "customer",
		})
		if createErr != nil {
			return createErr
		}
		_, createErr = tx.Exec(ctx, `
			INSERT INTO email_jobs (user_id, kind, scheduled_at)
			VALUES ($1, 'welcome_customer', now() + interval '30 minutes')`, user.ID)
		return createErr
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Session{}, ErrEmailTaken
		}
		return Session{}, fmt.Errorf("register customer: %w", err)
	}

	s.Audit(ctx, AuditEntry{
		ActorID:    &user.ID,
		Action:     "auth.customer_registered",
		EntityType: "user",
		EntityID:   &user.ID,
		IP:         ip,
		UserAgent:  ua,
	})

	session, err := s.startSession(ctx, user, uuid.Nil, ip, ua)
	if err != nil {
		return Session{}, err
	}

	token, err := s.createEmailVerification(ctx, user)
	if err != nil {
		s.Log.Warn("could not create email verification", "error", err, "user_id", user.ID)
		return session, nil
	}
	if err := s.Mailer.SendVerification(ctx, user.Email, token); err != nil {
		s.Log.Warn("could not send verification email", "error", err, "user_id", user.ID)
	}
	if s.Cfg.ExposeDevKeys {
		session.DevVerificationToken = token
	}
	return session, nil
}

func (s *Service) Login(ctx context.Context, email, password, ip, ua string) (Session, error) {
	user, err := s.Q.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, fmt.Errorf("login lookup: %w", err)
	}
	if user.Status != "active" || user.DeletedAt != nil {
		return Session{}, ErrUserSuspended
	}

	valid, err := auth.VerifyPassword(user.PasswordHash, password)
	if err != nil || !valid {
		return Session{}, ErrInvalidCredentials
	}

	_ = s.Q.TouchLastLogin(ctx, user.ID)
	s.Audit(ctx, AuditEntry{
		ActorID:    &user.ID,
		Action:     "auth.login",
		EntityType: "user",
		EntityID:   &user.ID,
		IP:         ip,
		UserAgent:  ua,
	})

	return s.startSession(ctx, user, uuid.Nil, ip, ua)
}

// ---------------------------------------------------------------------------
// Company registration
// ---------------------------------------------------------------------------

func (s *Service) RegisterCompany(ctx context.Context, input CompanyRegistration, ip, ua string) (Session, error) {
	email := normalizeEmail(input.BusinessEmail)
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return Session{}, err
	}

	var user db.User
	var company db.Company

	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		q := s.Q.WithTx(tx)

		var txErr error
		user, txErr = q.CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			PasswordHash: hash,
			FullName:     strings.TrimSpace(input.ContactName),
			Column4:      strings.TrimSpace(input.Phone),
			AccountType:  "company",
		})
		if txErr != nil {
			return txErr
		}

		company, txErr = q.CreateCompany(ctx, db.CreateCompanyParams{
			Name:          strings.TrimSpace(input.CompanyName),
			Slug:          newSlug(input.CompanyName),
			BusinessEmail: email,
			Phone:         strings.TrimSpace(input.Phone),
			BusinessType:  strings.TrimSpace(input.BusinessType),
			Country:       defaultString(strings.TrimSpace(input.Country), "NG"),
			Column7:       strings.TrimSpace(input.State),
			Column8:       strings.TrimSpace(input.City),
			Column9:       strings.TrimSpace(input.Address),
			OwnerUserID:   user.ID,
		})
		if txErr != nil {
			return txErr
		}

		now := time.Now()
		member, txErr := q.CreateCompanyMember(ctx, db.CreateCompanyMemberParams{
			CompanyID: company.ID,
			UserID:    user.ID,
			Status:    "active",
			Column4:   "Owner",
			JoinedAt:  &now,
		})
		if txErr != nil {
			return txErr
		}

		ownerRole, txErr := q.GetSystemRoleByKey(ctx, "owner")
		if txErr != nil {
			return fmt.Errorf("owner role: %w", txErr)
		}
		if txErr = q.AddMemberRole(ctx, db.AddMemberRoleParams{
			MemberID: member.ID,
			RoleID:   ownerRole.ID,
		}); txErr != nil {
			return txErr
		}

		if _, txErr = q.CreateCompanySettings(ctx, company.ID); txErr != nil {
			return txErr
		}
		if _, txErr = q.CreateCompanyBranding(ctx, company.ID); txErr != nil {
			return txErr
		}
		_, txErr = tx.Exec(ctx, `
			INSERT INTO email_jobs (user_id, company_id, kind, sequence, scheduled_at)
			VALUES
				($1, $2, 'welcome_provider', 0, now() + interval '30 minutes'),
				($1, $2, 'provider_onboarding', 1, now() + interval '24 hours'),
				($1, $2, 'provider_onboarding', 2, now() + interval '72 hours')`,
			user.ID, company.ID)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Session{}, ErrEmailTaken
		}
		return Session{}, fmt.Errorf("register company: %w", err)
	}
	s.Audit(ctx, AuditEntry{
		ActorID:    &user.ID,
		CompanyID:  &company.ID,
		Action:     "company.registered",
		EntityType: "company",
		EntityID:   &company.ID,
		Metadata:   map[string]any{"status": company.Status, "onboarding_status": company.OnboardingStatus},
		IP:         ip,
		UserAgent:  ua,
	})

	session, err := s.startSession(ctx, user, uuid.Nil, ip, ua)
	if err != nil {
		return Session{}, err
	}

	token, err := s.createEmailVerification(ctx, user)
	if err != nil {
		s.Log.Warn("could not create email verification", "error", err, "user_id", user.ID)
		return session, nil
	}
	if err := s.Mailer.SendVerification(ctx, user.Email, token); err != nil {
		s.Log.Warn("could not send verification email", "error", err, "user_id", user.ID)
	}
	if s.Cfg.ExposeDevKeys {
		session.DevVerificationToken = token
	}
	return session, nil
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func (s *Service) startSession(ctx context.Context, user db.User, familyID uuid.UUID, ip, ua string) (Session, error) {
	access, _, accessExpiry, err := s.Auth.Issue(user.ID)
	if err != nil {
		return Session{}, err
	}

	rawRefresh, refreshHash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return Session{}, err
	}
	if familyID == uuid.Nil {
		familyID = uuid.New()
	}
	refreshExpiry := time.Now().Add(s.Cfg.RefreshTTL)

	if _, err := s.Q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshHash,
		FamilyID:  familyID,
		Column4:   ua,
		Column5:   ip,
		ExpiresAt: refreshExpiry,
	}); err != nil {
		return Session{}, fmt.Errorf("store refresh token: %w", err)
	}

	companies, err := s.Q.ListUserMemberships(ctx, user.ID)
	if err != nil {
		return Session{}, fmt.Errorf("load memberships: %w", err)
	}

	return Session{
		User:             user,
		Companies:        companies,
		AccessToken:      access,
		AccessExpiresAt:  accessExpiry,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

// Refresh rotates a refresh token. Reusing an already rotated token revokes
// the whole token family, which ends the session everywhere it was replayed.
func (s *Service) Refresh(ctx context.Context, rawToken, ip, ua string) (Session, error) {
	hash := auth.HashToken(rawToken)

	var user db.User
	var accessToken string
	var accessExpiry time.Time
	var newRefresh string
	var refreshExpiry time.Time
	reused := false

	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		q := s.Q.WithTx(tx)

		stored, txErr := q.GetRefreshTokenByHash(ctx, hash)
		if txErr != nil {
			if errors.Is(txErr, pgx.ErrNoRows) {
				return ErrInvalidRefresh
			}
			return txErr
		}

		if stored.RevokedAt != nil {
			if stored.ReplacedBy != nil {
				// Commit the family revocation even though the request fails.
				// Returning an error here would roll the revocation back.
				if txErr = q.RevokeRefreshTokenFamily(ctx, stored.FamilyID); txErr != nil {
					return txErr
				}
				s.Audit(ctx, AuditEntry{
					ActorID:    &stored.UserID,
					Action:     "auth.refresh_reuse_detected",
					EntityType: "refresh_token",
					EntityID:   &stored.ID,
					IP:         ip,
					UserAgent:  ua,
				})
				reused = true
				return nil
			}
			return ErrInvalidRefresh
		}
		if time.Now().After(stored.ExpiresAt) {
			return ErrInvalidRefresh
		}

		user, txErr = q.GetUserByID(ctx, stored.UserID)
		if txErr != nil || user.Status != "active" {
			return ErrInvalidRefresh
		}

		var tokenErr error
		var tokenHash string
		newRefresh, tokenHash, tokenErr = auth.GenerateOpaqueToken()
		if tokenErr != nil {
			return tokenErr
		}
		refreshExpiry = time.Now().Add(s.Cfg.RefreshTTL)

		created, txErr := q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			UserID:    stored.UserID,
			TokenHash: tokenHash,
			FamilyID:  stored.FamilyID,
			Column4:   ua,
			Column5:   ip,
			ExpiresAt: refreshExpiry,
		})
		if txErr != nil {
			return txErr
		}
		if txErr = q.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
			ID:         stored.ID,
			ReplacedBy: &created.ID,
		}); txErr != nil {
			return txErr
		}

		accessToken, _, accessExpiry, txErr = s.Auth.Issue(user.ID)
		return txErr
	})
	if err != nil {
		return Session{}, err
	}
	if reused {
		return Session{}, ErrInvalidRefresh
	}

	companies, err := s.Q.ListUserMemberships(ctx, user.ID)
	if err != nil {
		return Session{}, fmt.Errorf("load memberships: %w", err)
	}

	return Session{
		User:             user,
		Companies:        companies,
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiry,
		RefreshToken:     newRefresh,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh, ip, ua string) {
	if rawRefresh == "" {
		return
	}
	stored, err := s.Q.GetRefreshTokenByHash(ctx, auth.HashToken(rawRefresh))
	if err != nil {
		return
	}
	_ = s.Q.RevokeRefreshTokenFamily(ctx, stored.FamilyID)
	s.Audit(ctx, AuditEntry{
		ActorID:    &stored.UserID,
		Action:     "auth.logout",
		EntityType: "refresh_token",
		EntityID:   &stored.ID,
		IP:         ip,
		UserAgent:  ua,
	})
}

// DenyAccessToken adds a JTI to Redis so the still-valid access token cannot
// be used again after logout.
func (s *Service) DenyAccessToken(ctx context.Context, jti string, until time.Time) {
	if s.Redis == nil || jti == "" {
		return
	}
	ttl := time.Until(until)
	if ttl <= 0 {
		return
	}
	_ = s.Redis.Set(ctx, "revoked:access:"+jti, "1", ttl).Err()
}

func (s *Service) IsAccessTokenRevoked(ctx context.Context, jti string) bool {
	if s.Redis == nil || jti == "" {
		return false
	}
	n, err := s.Redis.Exists(ctx, "revoked:access:"+jti).Result()
	return err == nil && n > 0
}

// ---------------------------------------------------------------------------
// Email verification and password reset
// ---------------------------------------------------------------------------

func (s *Service) createEmailVerification(ctx context.Context, user db.User) (string, error) {
	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return "", err
	}
	if _, err := s.Q.CreateEmailVerification(ctx, db.CreateEmailVerificationParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *Service) ResendEmailVerification(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.Q.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("load email verification recipient: %w", err)
	}
	if user.EmailVerifiedAt != nil {
		return true, nil
	}
	token, err := s.createEmailVerification(ctx, user)
	if err != nil {
		return false, fmt.Errorf("create email verification: %w", err)
	}
	if err := s.Mailer.SendVerification(ctx, user.Email, token); err != nil {
		return false, fmt.Errorf("send email verification: %w", err)
	}
	return false, nil
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	record, err := s.Q.GetEmailVerificationByHash(ctx, auth.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidToken
		}
		return err
	}
	if err := s.Q.MarkEmailVerified(ctx, record.UserID); err != nil {
		return err
	}
	_ = s.Q.ConsumeEmailVerification(ctx, record.ID)
	s.Audit(ctx, AuditEntry{
		ActorID:    &record.UserID,
		Action:     "auth.email_verified",
		EntityType: "user",
		EntityID:   &record.UserID,
	})
	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, email, ip string) (string, error) {
	user, err := s.Q.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // do not reveal whether the account exists
		}
		return "", err
	}

	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return "", err
	}
	if _, err := s.Q.CreatePasswordReset(ctx, db.CreatePasswordResetParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		return "", err
	}
	if err := s.Mailer.SendPasswordReset(ctx, user.Email, raw); err != nil {
		s.Log.Warn("could not send password reset email", "error", err, "user_id", user.ID)
	}
	s.Audit(ctx, AuditEntry{
		ActorID:    &user.ID,
		Action:     "auth.password_reset_requested",
		EntityType: "user",
		EntityID:   &user.ID,
		IP:         ip,
	})
	if s.Cfg.ExposeDevKeys {
		return raw, nil
	}
	return "", nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword, ip string) error {
	record, err := s.Q.GetPasswordResetByHash(ctx, auth.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidToken
		}
		return err
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.Q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           record.UserID,
		PasswordHash: hash,
	}); err != nil {
		return err
	}
	_ = s.Q.ConsumePasswordReset(ctx, record.ID)
	_ = s.Q.RevokeUserRefreshTokens(ctx, record.UserID)
	s.Audit(ctx, AuditEntry{
		ActorID:    &record.UserID,
		Action:     "auth.password_reset_completed",
		EntityType: "user",
		EntityID:   &record.UserID,
		IP:         ip,
	})
	return nil
}

// ---------------------------------------------------------------------------
// Staff invitations
// ---------------------------------------------------------------------------

type InvitationPreview struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	CompanyName   string
	CompanySlug   string
	Email         string
	RoleKey       string
	RoleName      string
	ExpiresAt     time.Time
	BranchIDs     []uuid.UUID
	AccountExists bool
}

func (s *Service) InvitationPreview(ctx context.Context, rawToken string) (InvitationPreview, error) {
	details, err := s.Q.GetInvitationDetails(ctx, auth.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InvitationPreview{}, ErrInvalidInvitation
		}
		return InvitationPreview{}, err
	}
	if details.AcceptedAt != nil || details.RevokedAt != nil || time.Now().After(details.ExpiresAt) {
		return InvitationPreview{}, ErrInvalidInvitation
	}

	branchIDs, err := s.Q.ListInvitationBranchIDs(ctx, details.ID)
	if err != nil {
		return InvitationPreview{}, err
	}

	accountExists := false
	if _, err := s.Q.GetUserByEmail(ctx, details.Email); err == nil {
		accountExists = true
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return InvitationPreview{}, err
	}

	return InvitationPreview{
		ID:            details.ID,
		CompanyID:     details.CompanyID,
		CompanyName:   details.CompanyName,
		CompanySlug:   details.CompanySlug,
		Email:         details.Email,
		RoleKey:       details.RoleKey,
		RoleName:      details.RoleName,
		ExpiresAt:     details.ExpiresAt,
		BranchIDs:     branchIDs,
		AccountExists: accountExists,
	}, nil
}

// AcceptInvitation joins the invited user to the company, applies the role and
// branch assignments, and claims the invitation. The claim is conditional so a
// token can only be used once, even under concurrent requests.
func (s *Service) AcceptInvitation(ctx context.Context, rawToken string, actor *db.User, fullName, password, ip, ua string) (Session, error) {
	preview, err := s.InvitationPreview(ctx, rawToken)
	if err != nil {
		return Session{}, err
	}

	if actor != nil && !strings.EqualFold(actor.Email, preview.Email) {
		return Session{}, ErrInvitationEmailMismatch
	}

	var user db.User
	newUser := false

	switch {
	case actor != nil:
		user = *actor
	default:
		existing, lookupErr := s.Q.GetUserByEmail(ctx, preview.Email)
		if lookupErr == nil {
			_ = existing
			return Session{}, ErrInvitationNeedsLogin
		}
		if !errors.Is(lookupErr, pgx.ErrNoRows) {
			return Session{}, lookupErr
		}
		hash, hashErr := auth.HashPassword(password)
		if hashErr != nil {
			return Session{}, hashErr
		}
		user, err = s.Q.CreateUser(ctx, db.CreateUserParams{
			Email:        preview.Email,
			PasswordHash: hash,
			FullName:     strings.TrimSpace(fullName),
			Column4:      "",
			AccountType:  "company",
		})
		if err != nil {
			if isUniqueViolation(err) {
				return Session{}, ErrInvitationNeedsLogin
			}
			return Session{}, err
		}
		newUser = true
	}

	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		q := s.Q.WithTx(tx)

		if _, txErr := q.ClaimInvitation(ctx, db.ClaimInvitationParams{
			ID:               preview.ID,
			AcceptedByUserID: &user.ID,
		}); txErr != nil {
			if errors.Is(txErr, pgx.ErrNoRows) {
				return ErrInvalidInvitation
			}
			return txErr
		}

		role, txErr := q.GetSystemRoleByKey(ctx, preview.RoleKey)
		if txErr != nil {
			return txErr
		}

		member, txErr := q.GetCompanyMember(ctx, db.GetCompanyMemberParams{
			CompanyID: preview.CompanyID,
			UserID:    user.ID,
		})
		if errors.Is(txErr, pgx.ErrNoRows) {
			now := time.Now()
			member, txErr = q.CreateCompanyMember(ctx, db.CreateCompanyMemberParams{
				CompanyID: preview.CompanyID,
				UserID:    user.ID,
				Status:    "active",
				Column4:   "Invited staff",
				JoinedAt:  &now,
			})
		} else if txErr == nil && (member.Status != "active" || member.DeletedAt != nil) {
			member, txErr = q.ReactivateCompanyMember(ctx, member.ID)
		}
		if txErr != nil {
			return txErr
		}

		if txErr = q.AddMemberRole(ctx, db.AddMemberRoleParams{
			MemberID: member.ID,
			RoleID:   role.ID,
		}); txErr != nil {
			return txErr
		}

		for _, branchID := range preview.BranchIDs {
			if _, branchErr := q.GetBranch(ctx, db.GetBranchParams{ID: branchID, CompanyID: preview.CompanyID}); branchErr != nil {
				return ErrInvalidInvitationBranch
			}
			if branchErr := q.AddBranchMember(ctx, db.AddBranchMemberParams{BranchID: branchID, MemberID: member.ID}); branchErr != nil {
				return branchErr
			}
		}
		if newUser {
			_, txErr = tx.Exec(ctx, `
				INSERT INTO email_jobs (user_id, kind, scheduled_at)
				VALUES ($1, 'welcome_provider', now() + interval '30 minutes')`, user.ID)
			if txErr != nil {
				return txErr
			}
		}

		return nil
	})
	if err != nil {
		if newUser {
			// The invitation failed; do not leave an orphan account behind.
			_, _ = s.DB.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		}
		return Session{}, err
	}

	session, err := s.startSession(ctx, user, uuid.Nil, ip, ua)
	if err != nil {
		return Session{}, err
	}
	if newUser {
		token, tokenErr := s.createEmailVerification(ctx, user)
		if tokenErr != nil {
			s.Log.Warn("could not create email verification for invited user", "error", tokenErr, "user_id", user.ID)
		} else {
			if err := s.Mailer.SendVerification(ctx, user.Email, token); err != nil {
				s.Log.Warn("could not send verification email to invited user", "error", err, "user_id", user.ID)
			}
			if s.Cfg.ExposeDevKeys {
				session.DevVerificationToken = token
			}
		}
	}

	s.Audit(ctx, AuditEntry{
		ActorID:    &user.ID,
		CompanyID:  &preview.CompanyID,
		Action:     "company.invitation_accepted",
		EntityType: "company_invitation",
		EntityID:   &preview.ID,
		Metadata:   map[string]any{"role": preview.RoleKey},
		IP:         ip,
		UserAgent:  ua,
	})
	return session, nil
}

// ---------------------------------------------------------------------------
// Platform administration
// ---------------------------------------------------------------------------

// EnsureSuperAdmin creates the bootstrap super administrator when it does not
// exist yet. It never changes an existing account, so rotating the env
// password does not silently overwrite a password set in the app.
func (s *Service) EnsureSuperAdmin(ctx context.Context, email, password string) (bool, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return false, nil
	}

	existing, err := s.Q.GetUserByEmail(ctx, email)
	if err == nil {
		if existing.AccountType != "platform" || existing.PlatformRole == nil || *existing.PlatformRole != "super_admin" {
			return false, fmt.Errorf("SUPER_ADMIN_EMAIL is already used by an account that is not a super administrator")
		}
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return false, err
	}
	now := time.Now()
	role := "super_admin"
	if _, err := s.Q.CreateUser(ctx, db.CreateUserParams{
		Email:           email,
		PasswordHash:    hash,
		FullName:        "Platform super admin",
		AccountType:     "platform",
		PlatformRole:    &role,
		EmailVerifiedAt: &now,
	}); err != nil {
		return false, err
	}
	return true, nil
}

type PlatformAdminInput struct {
	Email    string
	FullName string
	Password string
	Role     string
}

func (s *Service) CreatePlatformAdmin(ctx context.Context, input PlatformAdminInput) (db.User, error) {
	role := strings.TrimSpace(input.Role)
	if role != "admin" && role != "support" {
		return db.User{}, ErrInvalidPlatformRole
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return db.User{}, err
	}
	now := time.Now()
	user, err := s.Q.CreateUser(ctx, db.CreateUserParams{
		Email:           normalizeEmail(input.Email),
		PasswordHash:    hash,
		FullName:        strings.TrimSpace(input.FullName),
		AccountType:     "platform",
		PlatformRole:    &role,
		EmailVerifiedAt: &now,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.User{}, ErrEmailTaken
		}
		return db.User{}, err
	}
	return user, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *Service) Audit(ctx context.Context, entry AuditEntry) {
	metadata := []byte("{}")
	if len(entry.Metadata) > 0 {
		if encoded, err := json.Marshal(entry.Metadata); err == nil {
			metadata = encoded
		}
	}
	var entityID *uuid.UUID
	if entry.EntityID != nil {
		entityID = entry.EntityID
	}
	var entityType *string
	if entry.EntityType != "" {
		entityType = &entry.EntityType
	}

	_, err := s.Q.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ActorUserID: entry.ActorID,
		CompanyID:   entry.CompanyID,
		Action:      entry.Action,
		EntityType:  entityType,
		EntityID:    entityID,
		Metadata:    metadata,
		Column7:     entry.IP,
		Column8:     entry.UserAgent,
	})
	if err != nil {
		s.Log.Warn("audit log write failed", "action", entry.Action, "error", err)
	}
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func newSlug(name string) string {
	base := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if base == "" {
		base = "company"
	}
	if len(base) > 40 {
		base = strings.Trim(base[:40], "-")
	}
	suffix := make([]byte, 3)
	_, _ = rand.Read(suffix)
	return base + "-" + hex.EncodeToString(suffix)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
