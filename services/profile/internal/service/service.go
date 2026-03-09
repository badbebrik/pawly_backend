package service

import (
	"context"
	"errors"
	"fmt"
	"profile/internal/config"
	"profile/internal/model"
	"profile/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProfileService struct {
	repo       repository.ProfileRepository
	cfg        *config.Config
	fileClient FileClient
}

func NewService(repo repository.ProfileRepository, cfg *config.Config, fileClient FileClient) *ProfileService {
	return &ProfileService{repo: repo, cfg: cfg, fileClient: fileClient}
}

type CreateProfileInput struct {
	UserID    uuid.UUID
	Locale    *string
	FirstName *string
	LastName  *string
}

func (s *ProfileService) CreateProfile(ctx context.Context, in CreateProfileInput) (*model.Profile, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	firstName := normalizeOptionalString(in.FirstName)
	lastName := normalizeOptionalString(in.LastName)

	p, err := s.repo.GetByUserID(ctx, in.UserID)
	if err == nil {
		changed := false
		if p.FirstName == nil && firstName != nil {
			p.FirstName = firstName
			changed = true
		}
		if p.LastName == nil && lastName != nil {
			p.LastName = lastName
			changed = true
		}
		if changed {
			if err := s.repo.Update(ctx, p); err != nil {
				return nil, err
			}
		}
		return p, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	locale := s.cfg.DefaultLocale
	if in.Locale != nil && *in.Locale != "" {
		locale = *in.Locale
	}

	profile := &model.Profile{
		UserID:        in.UserID,
		FirstName:     firstName,
		LastName:      lastName,
		Locale:        locale,
		Timezone:      s.cfg.DefaultTimezone,
		DateFormat:    s.cfg.DefaultDateFmt,
		PublicContact: defaultPublicContactSettings(),
		ExtraContacts: model.ExtraContacts{},
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return s.repo.GetByUserID(ctx, in.UserID)
		}
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
	if patch.PublicContact != nil {
		p.PublicContact = *patch.PublicContact
	}
	if patch.ExtraContacts != nil {
		if err := validateExtraContacts(*patch.ExtraContacts); err != nil {
			return nil, err
		}
		p.ExtraContacts = *patch.ExtraContacts
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

type UpdateProfileInput struct {
	FirstName     *string
	LastName      *string
	Phone         *string
	Locale        *string
	Timezone      *string
	DateFormat    *string
	PublicContact *model.PublicContactSettings
	ExtraContacts *model.ExtraContacts
}

func defaultPublicContactSettings() model.PublicContactSettings {
	return model.PublicContactSettings{
		ShowOwnerName:     true,
		ShowPhone:         true,
		ShowEmail:         false,
		ShowExtraContacts: false,
	}
}

func (s *ProfileService) GetAvatarDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error) {
	if s.fileClient == nil {
		return "", time.Time{}, errors.New("file client not configured")
	}
	return s.fileClient.GetDownloadURL(ctx, fileID)
}

func (s *ProfileService) InitAvatarUpload(ctx context.Context, mimeType string, expectedSize *int64, userID uuid.UUID) (uuid.UUID, UploadInfo, error) {
	if s.fileClient == nil {
		return uuid.Nil, UploadInfo{}, errors.New("file client not configured")
	}
	size := int64(0)
	if expectedSize != nil {
		size = *expectedSize
	}
	return s.fileClient.InitUpload(ctx, mimeType, size, userID)
}

func (s *ProfileService) ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID, fileID uuid.UUID, sizeBytes int64) (*model.Profile, error) {
	if s.fileClient == nil {
		return nil, errors.New("file client not configured")
	}
	if sizeBytes <= 0 {
		return nil, errors.New("invalid size_bytes")
	}

	if err := s.fileClient.ConfirmUpload(ctx, fileID, sizeBytes); err != nil {
		return nil, err
	}
	if err := s.fileClient.LinkAvatar(ctx, fileID, userID); err != nil {
		return nil, err
	}

	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	p.AvatarFileID = &fileID
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func validateExtraContacts(m model.ExtraContacts) error {
	for k := range m {
		switch k {
		case "telegram", "whatsapp", "vk":
		default:
			return fmt.Errorf("invalid extra_contact key: %s", k)
		}
	}
	return nil
}

func normalizeOptionalString(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
