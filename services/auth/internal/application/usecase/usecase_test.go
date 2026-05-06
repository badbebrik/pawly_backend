package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"
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

type stubOAuthRepo struct {
	createFn                     func(context.Context, *model.OAuthIdentity) error
	getByProviderAndExternalIDFn func(context.Context, string, string) (*model.OAuthIdentity, error)
	getByUserIDFn                func(context.Context, uuid.UUID) ([]model.OAuthIdentity, error)
	getByEmailFn                 func(context.Context, string, string) (*model.OAuthIdentity, error)
}

func (s *stubOAuthRepo) Create(ctx context.Context, identity *model.OAuthIdentity) error {
	if s.createFn != nil {
		return s.createFn(ctx, identity)
	}
	return nil
}

func (s *stubOAuthRepo) GetByProviderAndExternalID(ctx context.Context, provider, externalID string) (*model.OAuthIdentity, error) {
	if s.getByProviderAndExternalIDFn != nil {
		return s.getByProviderAndExternalIDFn(ctx, provider, externalID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubOAuthRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error) {
	if s.getByUserIDFn != nil {
		return s.getByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubOAuthRepo) GetByEmail(ctx context.Context, provider, email string) (*model.OAuthIdentity, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, provider, email)
	}
	return nil, ports.ErrNotFound
}

type stubResetTokenRepo struct {
	consumeOnceFn func(context.Context, string, time.Duration) (bool, error)
}

func (s *stubResetTokenRepo) ConsumeOnce(ctx context.Context, tokenID string, ttl time.Duration) (bool, error) {
	if s.consumeOnceFn != nil {
		return s.consumeOnceFn(ctx, tokenID, ttl)
	}
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

type stubOAuthVerifier struct {
	verifyGoogleIDTokenFn func(context.Context, string) (*ports.OAuthClaims, error)
}

func (s *stubOAuthVerifier) VerifyGoogleIDToken(ctx context.Context, token string) (*ports.OAuthClaims, error) {
	if s.verifyGoogleIDTokenFn != nil {
		return s.verifyGoogleIDTokenFn(ctx, token)
	}
	return nil, ports.ErrOAuthInvalidToken
}

func newTestSet(deps Dependencies) *Set {
	return NewSet(deps)
}

func baseAuthDeps() Dependencies {
	return Dependencies{
		Users:        &stubUserRepo{},
		Sessions:     &stubSessionRepo{},
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
			generateResetTokenFn: func(userID, _ string) (string, error) {
				return "reset-" + userID, nil
			},
		},
		Profiles:    &stubProfileProvisioner{},
		OAuthVerify: &stubOAuthVerifier{},
		Clock:       fixedClock{now: time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)},
	}
}

func expectErr(t *testing.T, got error, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v want %v", got, want)
	}
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
		Password:  "Password123",
		FirstName: "Ivan",
		LastName:  "Ivanov",
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
	passwordHash, err := security.HashPassword("Password123")
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
		Password: "Password123",
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

func TestUsecases_RejectBasicInvalidInput(t *testing.T) {
	set := newTestSet(baseAuthDeps())

	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "register invalid email",
			run: func() error {
				_, err := set.RegisterEmail.Execute(context.Background(), RegisterEmailParams{Email: "bad", Password: "Password123"})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "register weak password",
			run: func() error {
				_, err := set.RegisterEmail.Execute(context.Background(), RegisterEmailParams{Email: "user@example.com", Password: "weak"})
				return err
			},
			want: ErrWeakPassword,
		},
		{
			name: "login empty password",
			run: func() error {
				_, err := set.LoginEmail.Execute(context.Background(), LoginEmailParams{Email: "user@example.com"})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "refresh empty token",
			run: func() error {
				_, err := set.Refresh.Execute(context.Background(), RefreshParams{})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "verify email invalid code",
			run: func() error {
				_, err := set.VerifyEmail.Execute(context.Background(), VerifyEmailParams{Email: "user@example.com", Code: "123"})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "resend invalid email",
			run: func() error {
				_, err := set.ResendEmailVerification.Execute(context.Background(), ResendEmailVerificationParams{Email: "bad"})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "password reset request invalid email",
			run: func() error {
				return set.PasswordResetRequest.Execute(context.Background(), PasswordResetRequestParams{Email: "bad"})
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "password reset verify invalid code",
			run: func() error {
				_, err := set.PasswordResetVerify.Execute(context.Background(), PasswordResetVerifyParams{Email: "user@example.com", Code: "abc"})
				return err
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "password reset confirm weak password",
			run: func() error {
				return set.PasswordResetConfirm.Execute(context.Background(), PasswordResetConfirmParams{ResetToken: "token", NewPassword: "weak"})
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "change password missing old password",
			run: func() error {
				return set.ChangePassword.Execute(context.Background(), ChangePasswordParams{AccessToken: "access", NewPassword: "Password123"})
			},
			want: ErrIncorrectFormat,
		},
		{
			name: "oauth unsupported provider",
			run: func() error {
				_, err := set.LoginOAuth.Execute(context.Background(), LoginOAuthParams{Provider: "github", IDToken: "token"})
				return err
			},
			want: ErrIncorrectFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.run(), tt.want)
		})
	}
}

func TestVerifyEmail_VerifiesUserAndCreatesSession(t *testing.T) {
	userID := uuid.New()
	var verifiedUserID uuid.UUID
	var createdSession *model.Session

	deps := baseAuthDeps()
	deps.Users = &stubUserRepo{
		getByEmailFn: func(context.Context, string) (*model.User, error) {
			return &model.User{
				ID:         userID,
				Email:      "user@example.com",
				IsVerified: false,
				IsActive:   true,
			}, nil
		},
		setVerifiedFn: func(_ context.Context, id uuid.UUID) error {
			verifiedUserID = id
			return nil
		},
	}
	deps.Sessions = &stubSessionRepo{
		createFn: func(_ context.Context, session *model.Session) error {
			createdSession = session
			return nil
		},
	}
	deps.Verification = &stubVerificationStore{
		verifyCodeFn: func(_ context.Context, email, purpose, code string) error {
			if email != "user@example.com" || purpose != "registration" || code != "123456" {
				t.Fatalf("unexpected verification args: %s %s %s", email, purpose, code)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	out, err := set.VerifyEmail.Execute(context.Background(), VerifyEmailParams{
		Email: "user@example.com",
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}

	if verifiedUserID != userID {
		t.Fatalf("unexpected verified user id: %s", verifiedUserID)
	}
	if createdSession == nil || createdSession.UserID != userID {
		t.Fatal("expected session for verified user")
	}
	if out.UserID != userID || out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("unexpected verify result: %+v", out)
	}
}

func TestResendEmailVerification_ReturnsMetaForUnverifiedUser(t *testing.T) {
	userID := uuid.New()
	deps := baseAuthDeps()
	deps.Users = &stubUserRepo{
		getByEmailFn: func(context.Context, string) (*model.User, error) {
			return &model.User{
				ID:         userID,
				Email:      "user@example.com",
				IsVerified: false,
				IsActive:   true,
			}, nil
		},
	}
	deps.Verification = &stubVerificationStore{
		requestCodeFn: func(context.Context, string, string) (string, int, int, error) {
			return "123456", 300, 60, nil
		},
	}

	set := newTestSet(deps)
	out, err := set.ResendEmailVerification.Execute(context.Background(), ResendEmailVerificationParams{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ResendEmailVerification returned error: %v", err)
	}
	if out.UserID != userID {
		t.Fatalf("unexpected user id: %s", out.UserID)
	}
	if out.Verification.Channel != "email" || out.Verification.CodeTTLSeconds != 300 || out.Verification.CanResendInSeconds != 60 {
		t.Fatalf("unexpected verification meta: %+v", out.Verification)
	}
}

func TestPasswordResetVerify_ReturnsResetToken(t *testing.T) {
	userID := uuid.New()
	deps := baseAuthDeps()
	deps.Users = &stubUserRepo{
		getByEmailFn: func(context.Context, string) (*model.User, error) {
			return &model.User{
				ID:         userID,
				Email:      "user@example.com",
				IsVerified: true,
				IsActive:   true,
			}, nil
		},
	}
	deps.Verification = &stubVerificationStore{
		verifyCodeFn: func(_ context.Context, email, purpose, code string) error {
			if email != "user@example.com" || purpose != "password_reset" || code != "123456" {
				t.Fatalf("unexpected verification args: %s %s %s", email, purpose, code)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	out, err := set.PasswordResetVerify.Execute(context.Background(), PasswordResetVerifyParams{
		Email: "user@example.com",
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("PasswordResetVerify returned error: %v", err)
	}
	if out.ResetToken != "reset-"+userID.String() {
		t.Fatalf("unexpected reset token: %s", out.ResetToken)
	}
}

func TestPasswordResetConfirm_UpdatesPasswordAndRevokesSessions(t *testing.T) {
	userID := uuid.New()
	resetTokenID := uuid.New().String()
	var consumedTokenID string
	var updatedPasswordHash string
	var revokedUserID uuid.UUID

	deps := baseAuthDeps()
	deps.Users = &stubUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*model.User, error) {
			if id != userID {
				t.Fatalf("unexpected user id: %s", id)
			}
			return &model.User{ID: userID, Email: "user@example.com", IsActive: true}, nil
		},
		updatePasswordHashFn: func(_ context.Context, id uuid.UUID, hash string) error {
			if id != userID {
				t.Fatalf("unexpected user id for password update: %s", id)
			}
			updatedPasswordHash = hash
			return nil
		},
	}
	deps.ResetTokens = &stubResetTokenRepo{
		consumeOnceFn: func(_ context.Context, tokenID string, ttl time.Duration) (bool, error) {
			consumedTokenID = tokenID
			if ttl <= 0 {
				t.Fatalf("expected positive reset token ttl, got %s", ttl)
			}
			return true, nil
		},
	}
	deps.Sessions = &stubSessionRepo{
		revokeAllFn: func(_ context.Context, id uuid.UUID) error {
			revokedUserID = id
			return nil
		},
	}
	deps.Tokens = &stubTokenManager{
		validateTokenFn: func(token string) (*ports.TokenClaims, error) {
			if token != "reset-token" {
				t.Fatalf("unexpected reset token: %s", token)
			}
			return &ports.TokenClaims{
				Subject:   userID.String(),
				SessionID: resetTokenID,
				Type:      ports.TokenTypeReset,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
		ensureTokenTypeFn: func(_ *ports.TokenClaims, tokenType string) error {
			if tokenType != ports.TokenTypeReset {
				t.Fatalf("unexpected token type: %s", tokenType)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	err := set.PasswordResetConfirm.Execute(context.Background(), PasswordResetConfirmParams{
		ResetToken:  "reset-token",
		NewPassword: "NewPassword123",
	})
	if err != nil {
		t.Fatalf("PasswordResetConfirm returned error: %v", err)
	}
	if consumedTokenID != resetTokenID {
		t.Fatalf("unexpected consumed token id: %s", consumedTokenID)
	}
	if err := security.ComparePassword(updatedPasswordHash, "NewPassword123"); err != nil {
		t.Fatalf("password hash was not updated: %v", err)
	}
	if revokedUserID != userID {
		t.Fatalf("unexpected revoked user id: %s", revokedUserID)
	}
}

func TestChangePassword_UpdatesPasswordAndRevokesSessions(t *testing.T) {
	userID := uuid.New()
	oldHash, err := security.HashPassword("OldPassword123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	var updatedPasswordHash string
	var revokedUserID uuid.UUID

	deps := baseAuthDeps()
	deps.Users = &stubUserRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Email:        "user@example.com",
				PasswordHash: &oldHash,
				IsActive:     true,
			}, nil
		},
		updatePasswordHashFn: func(_ context.Context, id uuid.UUID, hash string) error {
			if id != userID {
				t.Fatalf("unexpected user id for password update: %s", id)
			}
			updatedPasswordHash = hash
			return nil
		},
	}
	deps.Sessions = &stubSessionRepo{
		revokeAllFn: func(_ context.Context, id uuid.UUID) error {
			revokedUserID = id
			return nil
		},
	}
	deps.Tokens = &stubTokenManager{
		validateTokenFn: func(token string) (*ports.TokenClaims, error) {
			if token != "access-token" {
				t.Fatalf("unexpected access token: %s", token)
			}
			return &ports.TokenClaims{Subject: userID.String(), Type: ports.TokenTypeAccess}, nil
		},
		ensureTokenTypeFn: func(_ *ports.TokenClaims, tokenType string) error {
			if tokenType != ports.TokenTypeAccess {
				t.Fatalf("unexpected token type: %s", tokenType)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	err = set.ChangePassword.Execute(context.Background(), ChangePasswordParams{
		AccessToken: "access-token",
		OldPassword: "OldPassword123",
		NewPassword: "NewPassword123",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	if err := security.ComparePassword(updatedPasswordHash, "NewPassword123"); err != nil {
		t.Fatalf("password hash was not updated: %v", err)
	}
	if revokedUserID != userID {
		t.Fatalf("unexpected revoked user id: %s", revokedUserID)
	}
}

func TestLogout_RevokesSession(t *testing.T) {
	sessionID := uuid.New()
	var revokedSessionID uuid.UUID

	deps := baseAuthDeps()
	deps.Sessions = &stubSessionRepo{
		revokeFn: func(_ context.Context, id uuid.UUID) error {
			revokedSessionID = id
			return nil
		},
	}
	deps.Tokens = &stubTokenManager{
		validateTokenFn: func(string) (*ports.TokenClaims, error) {
			return &ports.TokenClaims{SessionID: sessionID.String(), Type: ports.TokenTypeAccess}, nil
		},
		ensureTokenTypeFn: func(_ *ports.TokenClaims, tokenType string) error {
			if tokenType != ports.TokenTypeAccess {
				t.Fatalf("unexpected token type: %s", tokenType)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	if err := set.Logout.Execute(context.Background(), "access-token"); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	if revokedSessionID != sessionID {
		t.Fatalf("unexpected revoked session id: %s", revokedSessionID)
	}
}

func TestLogoutAll_RevokesUserSessions(t *testing.T) {
	userID := uuid.New()
	var revokedUserID uuid.UUID

	deps := baseAuthDeps()
	deps.Sessions = &stubSessionRepo{
		revokeAllFn: func(_ context.Context, id uuid.UUID) error {
			revokedUserID = id
			return nil
		},
	}
	deps.Tokens = &stubTokenManager{
		validateTokenFn: func(string) (*ports.TokenClaims, error) {
			return &ports.TokenClaims{Subject: userID.String(), Type: ports.TokenTypeAccess}, nil
		},
		ensureTokenTypeFn: func(_ *ports.TokenClaims, tokenType string) error {
			if tokenType != ports.TokenTypeAccess {
				t.Fatalf("unexpected token type: %s", tokenType)
			}
			return nil
		},
	}

	set := newTestSet(deps)
	if err := set.LogoutAll.Execute(context.Background(), "access-token"); err != nil {
		t.Fatalf("LogoutAll returned error: %v", err)
	}
	if revokedUserID != userID {
		t.Fatalf("unexpected revoked user id: %s", revokedUserID)
	}
}

func TestLoginOAuth_ExistingIdentityCreatesSession(t *testing.T) {
	userID := uuid.New()
	var createdSession *model.Session

	deps := baseAuthDeps()
	deps.OAuthVerify = &stubOAuthVerifier{
		verifyGoogleIDTokenFn: func(context.Context, string) (*ports.OAuthClaims, error) {
			return &ports.OAuthClaims{
				Subject:       "google-subject",
				Email:         "user@example.com",
				EmailVerified: true,
			}, nil
		},
	}
	deps.OAuth = &stubOAuthRepo{
		getByProviderAndExternalIDFn: func(context.Context, string, string) (*model.OAuthIdentity, error) {
			return &model.OAuthIdentity{
				ID:         uuid.New(),
				UserID:     userID,
				Provider:   "google",
				ExternalID: "google-subject",
			}, nil
		},
	}
	deps.Users = &stubUserRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.User, error) {
			return &model.User{
				ID:         userID,
				Email:      "user@example.com",
				IsVerified: true,
				IsActive:   true,
			}, nil
		},
	}
	deps.Sessions = &stubSessionRepo{
		createFn: func(_ context.Context, session *model.Session) error {
			createdSession = session
			return nil
		},
	}

	set := newTestSet(deps)
	out, err := set.LoginOAuth.Execute(context.Background(), LoginOAuthParams{
		Provider: "google",
		IDToken:  "id-token",
	})
	if err != nil {
		t.Fatalf("LoginOAuth returned error: %v", err)
	}
	if createdSession == nil || createdSession.UserID != userID {
		t.Fatal("expected oauth login to create session")
	}
	if out.UserID != userID || out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatalf("unexpected oauth result: %+v", out)
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
