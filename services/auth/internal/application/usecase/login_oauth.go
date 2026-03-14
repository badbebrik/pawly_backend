package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type LoginOAuthUseCase struct {
	deps *dependencies
}

type LoginOAuthInput struct {
	Provider string
	IDToken  string
	Locale   string
}

type LoginOAuthOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (uc *LoginOAuthUseCase) Execute(ctx context.Context, in LoginOAuthInput) (*LoginOAuthOutput, error) {
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider != "google" || strings.TrimSpace(in.IDToken) == "" {
		return nil, ErrIncorrectFormat
	}

	claims, err := uc.deps.oauthVerify.VerifyGoogleIDToken(ctx, in.IDToken)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrOAuthInvalidToken):
			return nil, ErrOAuthInvalidToken
		case errors.Is(err, ports.ErrOAuthProviderUnavailable):
			return nil, ErrOAuthProviderUnavailable
		default:
			return nil, err
		}
	}

	identity, err := uc.deps.oauth.GetByProviderAndExternalID(ctx, provider, claims.Subject)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return nil, err
	}

	var user *model.User
	createdUser := false

	if identity != nil {
		user, err = uc.deps.users.GetByID(ctx, identity.UserID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return nil, ErrOAuthInvalidToken
			}
			return nil, err
		}
	} else {
		email := security.NormalizeEmail(claims.Email)
		if !security.ValidateEmail(email) || !claims.EmailVerified {
			return nil, ErrOAuthInvalidToken
		}

		user, err = uc.deps.users.GetByEmail(ctx, email)
		if err != nil && !errors.Is(err, ports.ErrNotFound) {
			return nil, err
		}

		if user == nil {
			user = &model.User{
				ID:         uuid.New(),
				Email:      email,
				IsVerified: true,
				IsActive:   true,
			}
			if err := uc.deps.createUserWithProfile(ctx, user, normalizeLocale(in.Locale), claims.FirstName, claims.LastName); err != nil {
				if !errors.Is(err, ports.ErrConflict) {
					return nil, err
				}
				user, err = uc.deps.users.GetByEmail(ctx, email)
				if err != nil {
					return nil, err
				}
			} else {
				createdUser = true
			}
		}

		emailCopy := email
		newIdentity := &model.OAuthIdentity{
			ID:         uuid.New(),
			UserID:     user.ID,
			Provider:   provider,
			ExternalID: claims.Subject,
			Email:      &emailCopy,
		}
		if err := uc.deps.oauth.Create(ctx, newIdentity); err != nil {
			if createdUser {
				uc.deps.compensateUserWithProfile(ctx, user.ID)
			}
			if existingIdentity, getErr := uc.deps.oauth.GetByProviderAndExternalID(ctx, provider, claims.Subject); getErr == nil {
				user, err = uc.deps.users.GetByID(ctx, existingIdentity.UserID)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	if err := user.RequireActive(); err != nil {
		return nil, ErrUserBlocked
	}

	pair, err := uc.deps.createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOAuthOutput{
		UserID:       user.ID,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int(uc.deps.tokens.AccessTTL().Seconds()),
	}, nil
}
