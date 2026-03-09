package service

import (
	"auth/internal/domain/model"
	"auth/internal/infrastructure/oauth"
	"auth/internal/repository"
	"auth/internal/security"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LoginOAuthInput struct {
	Provider string
	IDToken  string
}

type LoginOAuthOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (s *Service) LoginOAuth(ctx context.Context, in LoginOAuthInput) (*LoginOAuthOutput, error) {
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider != "google" || strings.TrimSpace(in.IDToken) == "" {
		return nil, ErrIncorrectFormat
	}

	claims, err := s.oauthVerify.VerifyGoogleIDToken(ctx, in.IDToken)
	if err != nil {
		switch {
		case errors.Is(err, oauth.ErrInvalidToken):
			return nil, ErrOAuthInvalidToken
		case errors.Is(err, oauth.ErrProviderUnavailable):
			return nil, ErrOAuthProviderUnavailable
		default:
			return nil, err
		}
	}

	identity, err := s.oauth.GetByProviderAndExternalID(ctx, provider, claims.Subject)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	var user *model.User
	createdUser := false

	if identity != nil {
		user, err = s.users.GetByID(ctx, identity.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrOAuthInvalidToken
			}
			return nil, err
		}
	} else {
		email := security.NormalizeEmail(claims.Email)
		if !security.ValidateEmail(email) || !claims.EmailVerified {
			return nil, ErrOAuthInvalidToken
		}

		user, err = s.users.GetByEmail(ctx, email)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}

		if user == nil {
			user = &model.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: nil,
				IsVerified:   true,
				IsActive:     true,
			}
			if err := s.users.Create(ctx, user); err != nil {
				if !errors.Is(err, repository.ErrConflict) {
					return nil, err
				}
				user, err = s.users.GetByEmail(ctx, email)
				if err != nil {
					return nil, err
				}
			} else {
				createdUser = true
				if err := s.profile.CreateProfile(ctx, user.ID, "", claims.FirstName, claims.LastName); err != nil {
					_ = s.users.Delete(ctx, user.ID)
					return nil, ErrProfileCreationFailed
				}
			}
		}

		em := email
		newIdentity := &model.OAuthIdentity{
			ID:         uuid.New(),
			UserID:     user.ID,
			Provider:   provider,
			ExternalID: claims.Subject,
			Email:      &em,
		}
		if err := s.oauth.Create(ctx, newIdentity); err != nil {
			if createdUser {
				_ = s.users.Delete(ctx, user.ID)
			}
			if identity, getErr := s.oauth.GetByProviderAndExternalID(ctx, provider, claims.Subject); getErr == nil {
				user, err = s.users.GetByID(ctx, identity.UserID)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	if !user.IsActive {
		return nil, ErrUserBlocked
	}

	sessionID := uuid.New()
	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}

	hash := security.HashTokenSHA256(refreshToken)
	expiresAt := time.Now().Add(time.Duration(s.refreshTTLDays) * 24 * time.Hour)

	sess := &model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hash,
		ExpiresAt:        expiresAt,
		IsRevoked:        false,
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, err
	}

	_ = s.users.UpdateLastLoginAt(ctx, user.ID, time.Now())

	return &LoginOAuthOutput{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.accessTTLSeconds,
	}, nil
}
