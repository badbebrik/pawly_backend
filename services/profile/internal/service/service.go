package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"profile/internal/config"
	"profile/internal/model"
	"profile/internal/repository"
)

type ProfileService struct {
	repo repository.ProfileRepository
	cfg  *config.Config
}

func NewService(repo repository.ProfileRepository, cfg *config.Config) *ProfileService {
	return &ProfileService{repo: repo, cfg: cfg}
}

func (s *ProfileService) CreateProfile(ctx context.Context, userID uuid.UUID, localeOpt *string) (*model.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err == nil {
		return p, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	locale := s.cfg.DefaultLocale
	if localeOpt != nil && *localeOpt != "" {
		locale = *localeOpt
	}

	profile := &model.Profile{
		UserID:        userID,
		Locale:        locale,
		Timezone:      s.cfg.DefaultTimezone,
		DateFormat:    s.cfg.DefaultDateFmt,
		Notifications: defaultNotifications(locale),
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, patch UpdateProfileInput) (*model.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if patch.FirstName != nil {
		p.FirstName = patch.FirstName
	}
	if patch.LastName != nil {
		p.LastName = patch.LastName
	}
	if patch.AvatarURL != nil {
		p.AvatarURL = patch.AvatarURL
	}
	if patch.Phone != nil {
		p.Phone = patch.Phone
	}
	if patch.Locale != nil && *patch.Locale != "" {
		p.Locale = *patch.Locale
	}
	if patch.Timezone != nil && *patch.Timezone != "" {
		p.Timezone = *patch.Timezone
	}
	if patch.DateFormat != nil && *patch.DateFormat != "" {
		p.DateFormat = *patch.DateFormat
	}
	if patch.Notifications != nil {
		p.Notifications = patch.Notifications
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

type UpdateProfileInput struct {
	FirstName     *string
	LastName      *string
	AvatarURL     *string
	Phone         *string
	Locale        *string
	Timezone      *string
	DateFormat    *string
	Notifications map[string]any
}

func defaultNotifications(locale string) map[string]any {
	return map[string]any{
		"email": map[string]any{
			"task_reminders":    false,
			"medical_reminders": false,
		},
		"push": map[string]any{
			"task_reminders":    false,
			"medical_reminders": false,
			"chat":              false,
		},
	}
}
