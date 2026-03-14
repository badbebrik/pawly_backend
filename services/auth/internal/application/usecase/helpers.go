package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultLocale = "ru"

type tokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshTokenHash string
	RefreshExpiresAt time.Time
}

func (d *dependencies) now() time.Time {
	return d.clock.Now()
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		return defaultLocale
	}
	return locale
}

func (d *dependencies) issueTokenPair(userID, sessionID uuid.UUID) (*tokenPair, error) {
	refreshToken, err := d.tokens.GenerateRefreshToken(userID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}

	accessToken, err := d.tokens.GenerateAccessToken(userID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}

	return &tokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshTokenHash: security.HashTokenSHA256(refreshToken),
		RefreshExpiresAt: d.now().Add(d.tokens.RefreshTTL()),
	}, nil
}

func (d *dependencies) createSession(ctx context.Context, userID uuid.UUID) (*tokenPair, error) {
	sessionID := uuid.New()

	pair, err := d.issueTokenPair(userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := d.sessions.Create(ctx, &model.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: pair.RefreshTokenHash,
		ExpiresAt:        pair.RefreshExpiresAt,
		IsRevoked:        false,
	}); err != nil {
		return nil, err
	}

	d.touchLastLogin(ctx, userID)

	return pair, nil
}

func (d *dependencies) rotateSession(ctx context.Context, userID, sessionID uuid.UUID, currentRefreshHash string) (*tokenPair, error) {
	pair, err := d.issueTokenPair(userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := d.sessions.UpdateRefreshToken(ctx, sessionID, currentRefreshHash, pair.RefreshTokenHash, pair.RefreshExpiresAt); err != nil {
		return nil, err
	}

	d.touchLastLogin(ctx, userID)

	return pair, nil
}

func (d *dependencies) touchLastLogin(ctx context.Context, userID uuid.UUID) {
	_ = d.users.UpdateLastLoginAt(ctx, userID, d.now())
}

func (d *dependencies) createUserWithProfile(ctx context.Context, user *model.User, locale, firstName, lastName string) error {
	if err := d.users.Create(ctx, user); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.ErrConflict
		}
		return err
	}

	if err := d.profiles.CreateProfile(ctx, user.ID, locale, firstName, lastName); err != nil {
		_ = d.users.Delete(ctx, user.ID)
		return ErrProfileCreationFailed
	}

	return nil
}

func (d *dependencies) compensateUserWithProfile(ctx context.Context, userID uuid.UUID) {
	_ = d.profiles.DeleteProfile(ctx, userID)
	_ = d.users.Delete(ctx, userID)
}
