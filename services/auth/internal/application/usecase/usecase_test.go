package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type stubUserRepo struct {
	createFn             func(context.Context, *model.User) error
	getByIDFn            func(context.Context, uuid.UUID) (*model.User, error)
	getByEmailFn         func(context.Context, string) (*model.User, error)
	setVerifiedFn        func(context.Context, uuid.UUID) error
	updatePasswordHashFn func(context.Context, uuid.UUID, string) error
	updateEmailFn        func(context.Context, uuid.UUID, string) error
	setActiveFn          func(context.Context, uuid.UUID, bool) error
	updateLastLoginAtFn  func(context.Context, uuid.UUID, time.Time) error
	softDeleteFn         func(context.Context, uuid.UUID) error
	deleteFn             func(context.Context, uuid.UUID) error
}

func (s *stubUserRepo) Create(ctx context.Context, user *model.User) error {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}
	return nil
}

func (s *stubUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, ports.ErrNotFound
}

func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, email)
	}
	return nil, ports.ErrNotFound
}

func (s *stubUserRepo) SetVerified(ctx context.Context, id uuid.UUID) error {
	if s.setVerifiedFn != nil {
		return s.setVerifiedFn(ctx, id)
	}
	return nil
}

func (s *stubUserRepo) UpdatePasswordHash(ctx context.Context, id uuid.UUID, newHash string) error {
	if s.updatePasswordHashFn != nil {
		return s.updatePasswordHashFn(ctx, id, newHash)
	}
	return nil
}

func (s *stubUserRepo) UpdateEmail(ctx context.Context, id uuid.UUID, newEmail string) error {
	if s.updateEmailFn != nil {
		return s.updateEmailFn(ctx, id, newEmail)
	}
	return nil
}

func (s *stubUserRepo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	if s.setActiveFn != nil {
		return s.setActiveFn(ctx, id, active)
	}
	return nil
}

func (s *stubUserRepo) UpdateLastLoginAt(ctx context.Context, id uuid.UUID, at time.Time) error {
	if s.updateLastLoginAtFn != nil {
		return s.updateLastLoginAtFn(ctx, id, at)
	}
	return nil
}

func (s *stubUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if s.softDeleteFn != nil {
		return s.softDeleteFn(ctx, id)
	}
	return nil
}

func (s *stubUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

type stubSessionRepo struct {
	createFn             func(context.Context, *model.Session) error
	getByIDFn            func(context.Context, uuid.UUID) (*model.Session, error)
	updateRefreshTokenFn func(context.Context, uuid.UUID, string, string, time.Time) error
	revokeFn             func(context.Context, uuid.UUID) error
	revokeAllFn          func(context.Context, uuid.UUID) error
}

func (s *stubSessionRepo) Create(ctx context.Context, session *model.Session) error {
	if s.createFn != nil {
		return s.createFn(ctx, session)
	}
	return nil
}

func (s *stubSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, ports.ErrNotFound
}

func (s *stubSessionRepo) UpdateRefreshToken(ctx context.Context, id uuid.UUID, oldHash string, newHash string, newExpiresAt time.Time) error {
	if s.updateRefreshTokenFn != nil {
		return s.updateRefreshTokenFn(ctx, id, oldHash, newHash, newExpiresAt)
	}
	return nil
}

func (s *stubSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if s.revokeFn != nil {
		return s.revokeFn(ctx, id)
	}
	return nil
}

func (s *stubSessionRepo) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	if s.revokeAllFn != nil {
		return s.revokeAllFn(ctx, userID)
	}
	return nil
}

type stubOAuthRepo struct{}

func (s *stubOAuthRepo) Create(context.Context, *model.OAuthIdentity) error {
	return nil
}

func (s *stubOAuthRepo) GetByProviderAndExternalID(context.Context, string, string) (*model.OAuthIdentity, error) {
	return nil, ports.ErrNotFound
}

func (s *stubOAuthRepo) GetByUserID(context.Context, uuid.UUID) ([]model.OAuthIdentity, error) {
	return nil, nil
}

func (s *stubOAuthRepo) GetByEmail(context.Context, string, string) (*model.OAuthIdentity, error) {
	return nil, ports.ErrNotFound
}

type stubResetTokenRepo struct{}

func (s *stubResetTokenRepo) ConsumeOnce(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}

type stubVerificationStore struct {
	requestCodeFn func(context.Context, string, string) (string, int, int, error)
	verifyCodeFn  func(context.Context, string, string, string) error
}

func (s *stubVerificationStore) RequestCode(ctx context.Context, email, purpose string) (string, int, int, error) {
	if s.requestCodeFn != nil {
		return s.requestCodeFn(ctx, email, purpose)
	}
	return "", 0, 0, nil
}

func (s *stubVerificationStore) VerifyCode(ctx context.Context, email, purpose, code string) error {
	if s.verifyCodeFn != nil {
		return s.verifyCodeFn(ctx, email, purpose, code)
	}
	return nil
}

type stubNotifier struct {
	sendEmailVerificationFn func(context.Context, ports.EmailVerificationMessage) error
	sendPasswordResetFn     func(context.Context, ports.PasswordResetMessage) error
	sendWelcomeEmailFn      func(context.Context, ports.WelcomeEmailMessage) error
}

func (s *stubNotifier) SendEmailVerification(ctx context.Context, msg ports.EmailVerificationMessage) error {
	if s.sendEmailVerificationFn != nil {
		return s.sendEmailVerificationFn(ctx, msg)
	}
	return nil
}

func (s *stubNotifier) SendPasswordReset(ctx context.Context, msg ports.PasswordResetMessage) error {
	if s.sendPasswordResetFn != nil {
		return s.sendPasswordResetFn(ctx, msg)
	}
	return nil
}

func (s *stubNotifier) SendWelcomeEmail(ctx context.Context, msg ports.WelcomeEmailMessage) error {
	if s.sendWelcomeEmailFn != nil {
		return s.sendWelcomeEmailFn(ctx, msg)
	}
	return nil
}

type stubTokenManager struct {
	validateTokenFn        func(string) (*ports.TokenClaims, error)
	ensureTokenTypeFn      func(*ports.TokenClaims, string) error
	generateAccessTokenFn  func(string, string) (string, error)
	generateRefreshTokenFn func(string, string) (string, error)
	generateResetTokenFn   func(string, string) (string, error)
	accessTTL              time.Duration
	refreshTTL             time.Duration
}

func (s *stubTokenManager) GenerateAccessToken(userID, sessionID string) (string, error) {
	if s.generateAccessTokenFn != nil {
		return s.generateAccessTokenFn(userID, sessionID)
	}
	return "", nil
}

func (s *stubTokenManager) GenerateRefreshToken(userID, sessionID string) (string, error) {
	if s.generateRefreshTokenFn != nil {
		return s.generateRefreshTokenFn(userID, sessionID)
	}
	return "", nil
}

func (s *stubTokenManager) GeneratePasswordResetToken(userID, email string) (string, error) {
	if s.generateResetTokenFn != nil {
		return s.generateResetTokenFn(userID, email)
	}
	return "", nil
}

func (s *stubTokenManager) ValidateToken(tokenStr string) (*ports.TokenClaims, error) {
	if s.validateTokenFn != nil {
		return s.validateTokenFn(tokenStr)
	}
	return nil, nil
}

func (s *stubTokenManager) EnsureTokenType(claims *ports.TokenClaims, tokenType string) error {
	if s.ensureTokenTypeFn != nil {
		return s.ensureTokenTypeFn(claims, tokenType)
	}
	return nil
}

func (s *stubTokenManager) AccessTTL() time.Duration {
	if s.accessTTL != 0 {
		return s.accessTTL
	}
	return 15 * time.Minute
}

func (s *stubTokenManager) RefreshTTL() time.Duration {
	if s.refreshTTL != 0 {
		return s.refreshTTL
	}
	return 30 * 24 * time.Hour
}

type stubProfileProvisioner struct {
	createProfileFn func(context.Context, uuid.UUID, string, string, string, string) error
	deleteProfileFn func(context.Context, uuid.UUID) error
}

func (s *stubProfileProvisioner) CreateProfile(ctx context.Context, userID uuid.UUID, locale string, timezone string, firstName string, lastName string) error {
	if s.createProfileFn != nil {
		return s.createProfileFn(ctx, userID, locale, timezone, firstName, lastName)
	}
	return nil
}

func (s *stubProfileProvisioner) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	if s.deleteProfileFn != nil {
		return s.deleteProfileFn(ctx, userID)
	}
	return nil
}

type stubOAuthVerifier struct{}

func (s *stubOAuthVerifier) VerifyGoogleIDToken(context.Context, string) (*ports.OAuthClaims, error) {
	return nil, ports.ErrOAuthInvalidToken
}

func newTestSet(deps Dependencies) *Set {
	return NewSet(deps)
}

func TestRegisterEmail_UsesExplicitLocaleForProfileAndVerification(t *testing.T) {
	var createdUser *model.User
	var createdProfileLocale string
	var verificationLocale string

	set := newTestSet(Dependencies{
		Users: &stubUserRepo{
			getByEmailFn: func(context.Context, string) (*model.User, error) {
				return nil, ports.ErrNotFound
			},
			createFn: func(_ context.Context, user *model.User) error {
				createdUser = user
				return nil
			},
		},
		Sessions:    &stubSessionRepo{},
		OAuth:       &stubOAuthRepo{},
		ResetTokens: &stubResetTokenRepo{},
		Verification: &stubVerificationStore{
			requestCodeFn: func(_ context.Context, email, purpose string) (string, int, int, error) {
				if email != "user@example.com" {
					t.Fatalf("unexpected email: %s", email)
				}
				if purpose != "registration" {
					t.Fatalf("unexpected purpose: %s", purpose)
				}
				return "123456", 300, 60, nil
			},
		},
		Notifier: &stubNotifier{
			sendEmailVerificationFn: func(_ context.Context, msg ports.EmailVerificationMessage) error {
				verificationLocale = msg.Locale
				return nil
			},
		},
		Tokens: &stubTokenManager{},
		Profiles: &stubProfileProvisioner{
			createProfileFn: func(_ context.Context, _ uuid.UUID, locale string, _ string, _, _ string) error {
				createdProfileLocale = locale
				return nil
			},
		},
		OAuthVerify: &stubOAuthVerifier{},
	})

	out, err := set.RegisterEmail.Execute(context.Background(), RegisterEmailParams{
		Email:     "User@Example.com",
		Password:  "StrongPass123",
		FirstName: "Vika",
		LastName:  "Petrova",
		Locale:    "EN",
	})
	if err != nil {
		t.Fatalf("RegisterEmail returned error: %v", err)
	}

	if createdUser == nil {
		t.Fatal("expected user to be created")
	}
	if createdUser.Email != "user@example.com" {
		t.Fatalf("unexpected normalized email: %s", createdUser.Email)
	}
	if createdProfileLocale != "en" {
		t.Fatalf("unexpected profile locale: %s", createdProfileLocale)
	}
	if verificationLocale != "en" {
		t.Fatalf("unexpected verification locale: %s", verificationLocale)
	}
	if out.Verification.CodeTTLSeconds != 300 {
		t.Fatalf("unexpected code ttl: %d", out.Verification.CodeTTLSeconds)
	}
}

func TestLoginEmail_CreatesSessionAndTouchesLastLogin(t *testing.T) {
	passwordHash, err := security.HashPassword("StrongPass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	userID := uuid.New()
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)

	var createdSession *model.Session
	var lastLoginAt time.Time

	set := newTestSet(Dependencies{
		Users: &stubUserRepo{
			getByEmailFn: func(context.Context, string) (*model.User, error) {
				return &model.User{
					ID:           userID,
					Email:        "user@example.com",
					PasswordHash: &passwordHash,
					IsVerified:   true,
					IsActive:     true,
				}, nil
			},
			updateLastLoginAtFn: func(_ context.Context, id uuid.UUID, at time.Time) error {
				if id != userID {
					t.Fatalf("unexpected user id for last login: %s", id)
				}
				lastLoginAt = at
				return nil
			},
		},
		Sessions: &stubSessionRepo{
			createFn: func(_ context.Context, session *model.Session) error {
				createdSession = session
				return nil
			},
		},
		OAuth:        &stubOAuthRepo{},
		ResetTokens:  &stubResetTokenRepo{},
		Verification: &stubVerificationStore{},
		Notifier:     &stubNotifier{},
		Tokens: &stubTokenManager{
			generateAccessTokenFn: func(_, sessionID string) (string, error) {
				return "access-" + sessionID, nil
			},
			generateRefreshTokenFn: func(_, sessionID string) (string, error) {
				return "refresh-" + sessionID, nil
			},
			accessTTL:  15 * time.Minute,
			refreshTTL: 30 * 24 * time.Hour,
		},
		Profiles:    &stubProfileProvisioner{},
		OAuthVerify: &stubOAuthVerifier{},
		Clock:       fixedClock{now: now},
	})

	out, err := set.LoginEmail.Execute(context.Background(), LoginEmailParams{
		Email:    "user@example.com",
		Password: "StrongPass123",
		Locale:   "ru",
	})
	if err != nil {
		t.Fatalf("LoginEmail returned error: %v", err)
	}

	if createdSession == nil {
		t.Fatal("expected session to be created")
	}
	if got, want := createdSession.RefreshTokenHash, security.HashTokenSHA256(out.RefreshToken); got != want {
		t.Fatalf("unexpected refresh token hash: got %s want %s", got, want)
	}
	if got, want := createdSession.ExpiresAt, now.Add(30*24*time.Hour); !got.Equal(want) {
		t.Fatalf("unexpected session expiry: got %s want %s", got, want)
	}
	if !lastLoginAt.Equal(now) {
		t.Fatalf("unexpected last login time: %s", lastLoginAt)
	}
	if out.AccessToken != "access-"+createdSession.ID.String() {
		t.Fatalf("unexpected access token: %s", out.AccessToken)
	}
}

func TestRefresh_RotatesSessionUsingClock(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	oldRefresh := "old-refresh-token"

	var rotatedOldHash string
	var rotatedNewHash string
	var rotatedExpiresAt time.Time
	var touchedAt time.Time

	set := newTestSet(Dependencies{
		Users: &stubUserRepo{
			updateLastLoginAtFn: func(_ context.Context, id uuid.UUID, at time.Time) error {
				if id != userID {
					t.Fatalf("unexpected user id for last login: %s", id)
				}
				touchedAt = at
				return nil
			},
		},
		Sessions: &stubSessionRepo{
			getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Session, error) {
				if id != sessionID {
					t.Fatalf("unexpected session id: %s", id)
				}
				return &model.Session{
					ID:               sessionID,
					UserID:           userID,
					RefreshTokenHash: security.HashTokenSHA256(oldRefresh),
					ExpiresAt:        now.Add(time.Hour),
					IsRevoked:        false,
				}, nil
			},
			updateRefreshTokenFn: func(_ context.Context, id uuid.UUID, oldHash string, newHash string, expiresAt time.Time) error {
				if id != sessionID {
					t.Fatalf("unexpected session id on rotate: %s", id)
				}
				rotatedOldHash = oldHash
				rotatedNewHash = newHash
				rotatedExpiresAt = expiresAt
				return nil
			},
		},
		OAuth:        &stubOAuthRepo{},
		ResetTokens:  &stubResetTokenRepo{},
		Verification: &stubVerificationStore{},
		Notifier:     &stubNotifier{},
		Tokens: &stubTokenManager{
			validateTokenFn: func(token string) (*ports.TokenClaims, error) {
				if token != oldRefresh {
					t.Fatalf("unexpected refresh token: %s", token)
				}
				return &ports.TokenClaims{
					Subject:   userID.String(),
					SessionID: sessionID.String(),
					Type:      ports.TokenTypeRefresh,
				}, nil
			},
			ensureTokenTypeFn: func(_ *ports.TokenClaims, tokenType string) error {
				if tokenType != ports.TokenTypeRefresh {
					t.Fatalf("unexpected token type: %s", tokenType)
				}
				return nil
			},
			generateAccessTokenFn: func(_, sid string) (string, error) {
				return "access-" + sid, nil
			},
			generateRefreshTokenFn: func(_, sid string) (string, error) {
				return "refresh-" + sid, nil
			},
			accessTTL:  15 * time.Minute,
			refreshTTL: 10 * 24 * time.Hour,
		},
		Profiles:    &stubProfileProvisioner{},
		OAuthVerify: &stubOAuthVerifier{},
		Clock:       fixedClock{now: now},
	})

	out, err := set.Refresh.Execute(context.Background(), RefreshParams{RefreshToken: oldRefresh})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	if rotatedOldHash != security.HashTokenSHA256(oldRefresh) {
		t.Fatalf("unexpected old hash: %s", rotatedOldHash)
	}
	if rotatedNewHash != security.HashTokenSHA256(out.RefreshToken) {
		t.Fatalf("unexpected new hash: %s", rotatedNewHash)
	}
	if want := now.Add(10 * 24 * time.Hour); !rotatedExpiresAt.Equal(want) {
		t.Fatalf("unexpected rotated expiry: got %s want %s", rotatedExpiresAt, want)
	}
	if !touchedAt.Equal(now) {
		t.Fatalf("unexpected touch time: %s", touchedAt)
	}
}

func TestRequestPasswordReset_UsesExplicitLocale(t *testing.T) {
	userID := uuid.New()
	var sentLocale string

	set := newTestSet(Dependencies{
		Users: &stubUserRepo{
			getByEmailFn: func(context.Context, string) (*model.User, error) {
				return &model.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsActive:   true,
				}, nil
			},
		},
		Sessions:    &stubSessionRepo{},
		OAuth:       &stubOAuthRepo{},
		ResetTokens: &stubResetTokenRepo{},
		Verification: &stubVerificationStore{
			requestCodeFn: func(_ context.Context, email, purpose string) (string, int, int, error) {
				if purpose != "password_reset" {
					t.Fatalf("unexpected purpose: %s", purpose)
				}
				return "654321", 180, 0, nil
			},
		},
		Notifier: &stubNotifier{
			sendPasswordResetFn: func(_ context.Context, msg ports.PasswordResetMessage) error {
				sentLocale = msg.Locale
				return nil
			},
		},
		Tokens:      &stubTokenManager{},
		Profiles:    &stubProfileProvisioner{},
		OAuthVerify: &stubOAuthVerifier{},
	})

	err := set.PasswordResetRequest.Execute(context.Background(), PasswordResetRequestParams{
		Email:  "user@example.com",
		Locale: "EN",
	})
	if err != nil {
		t.Fatalf("PasswordResetRequest returned error: %v", err)
	}

	if sentLocale != "en" {
		t.Fatalf("unexpected reset locale: %s", sentLocale)
	}
}

var (
	_ ports.UserRepository          = (*stubUserRepo)(nil)
	_ ports.SessionRepository       = (*stubSessionRepo)(nil)
	_ ports.OAuthIdentityRepository = (*stubOAuthRepo)(nil)
	_ ports.ResetTokenRepository    = (*stubResetTokenRepo)(nil)
	_ ports.VerificationStore       = (*stubVerificationStore)(nil)
	_ ports.Notifier                = (*stubNotifier)(nil)
	_ ports.TokenManager            = (*stubTokenManager)(nil)
	_ ports.ProfileProvisioner      = (*stubProfileProvisioner)(nil)
	_ ports.OAuthVerifier           = (*stubOAuthVerifier)(nil)
	_ ports.Clock                   = fixedClock{}
)
