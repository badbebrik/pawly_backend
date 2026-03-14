package service

import (
	"context"
	"errors"
	"profile/internal/config"
	"profile/internal/model"
	"profile/internal/repository"
	"regexp"
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
	Timezone  *string
	FirstName *string
	LastName  *string
}

func (s *ProfileService) CreateProfile(ctx context.Context, in CreateProfileInput) (*model.Profile, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	locale, err := s.normalizeLocale(in.Locale)
	if err != nil {
		return nil, err
	}
	timezone, err := s.normalizeTimezone(in.Timezone)
	if err != nil {
		return nil, err
	}

	firstName := normalizeOptionalString(in.FirstName)
	lastName := normalizeOptionalString(in.LastName)

	profile := &model.Profile{
		UserID:    in.UserID,
		FirstName: firstName,
		LastName:  lastName,
		Locale:    locale,
		Timezone:  timezone,
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

type ProfileBrief struct {
	UserID            uuid.UUID
	FirstName         *string
	LastName          *string
	AvatarFileID      *uuid.UUID
	AvatarDownloadURL *string
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ProfileService) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, userID)
}

type Preferences struct {
	UserID   uuid.UUID
	Locale   string
	Timezone string
}

func (s *ProfileService) GetPreferences(ctx context.Context, userID uuid.UUID) (*Preferences, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &Preferences{
		UserID:   profile.UserID,
		Locale:   profile.Locale,
		Timezone: profile.Timezone,
	}, nil
}

func (s *ProfileService) BatchGetPreferences(ctx context.Context, userIDs []uuid.UUID) ([]Preferences, []uuid.UUID, error) {
	if len(userIDs) == 0 {
		return []Preferences{}, []uuid.UUID{}, nil
	}

	unique := uniqueUserIDs(userIDs)
	if len(unique) == 0 {
		return []Preferences{}, []uuid.UUID{}, nil
	}

	profiles, err := s.repo.GetByUserIDs(ctx, unique)
	if err != nil {
		return nil, nil, err
	}

	profilesByUser := make(map[uuid.UUID]model.Profile, len(profiles))
	for i := range profiles {
		profilesByUser[profiles[i].UserID] = profiles[i]
	}

	items := make([]Preferences, 0, len(profilesByUser))
	notFound := make([]uuid.UUID, 0)
	for i := range unique {
		profile, ok := profilesByUser[unique[i]]
		if !ok {
			notFound = append(notFound, unique[i])
			continue
		}
		items = append(items, Preferences{
			UserID:   profile.UserID,
			Locale:   profile.Locale,
			Timezone: profile.Timezone,
		})
	}

	return items, notFound, nil
}

func (s *ProfileService) BatchGetProfilesBrief(ctx context.Context, userIDs []uuid.UUID) ([]ProfileBrief, []uuid.UUID, error) {
	if len(userIDs) == 0 {
		return []ProfileBrief{}, []uuid.UUID{}, nil
	}

	unique := uniqueUserIDs(userIDs)
	if len(unique) == 0 {
		return []ProfileBrief{}, []uuid.UUID{}, nil
	}

	profiles, err := s.repo.GetByUserIDs(ctx, unique)
	if err != nil {
		return nil, nil, err
	}

	profilesByUser := make(map[uuid.UUID]model.Profile, len(profiles))
	avatarIDs := make([]uuid.UUID, 0, len(profiles))
	for i := range profiles {
		profile := profiles[i]
		profilesByUser[profile.UserID] = profile
		if profile.AvatarFileID != nil {
			avatarIDs = append(avatarIDs, *profile.AvatarFileID)
		}
	}

	avatarURLByID := map[uuid.UUID]string{}
	if s.fileClient != nil && len(avatarIDs) > 0 {
		if urls, err := s.fileClient.BatchGetDownloadURLs(ctx, avatarIDs); err == nil {
			avatarURLByID = urls
		}
	}

	items := make([]ProfileBrief, 0, len(profilesByUser))
	notFound := make([]uuid.UUID, 0)
	for i := range unique {
		userID := unique[i]
		profile, ok := profilesByUser[userID]
		if !ok {
			notFound = append(notFound, userID)
			continue
		}

		item := ProfileBrief{
			UserID:       profile.UserID,
			FirstName:    profile.FirstName,
			LastName:     profile.LastName,
			AvatarFileID: profile.AvatarFileID,
		}
		if profile.AvatarFileID != nil {
			if url, ok := avatarURLByID[*profile.AvatarFileID]; ok && strings.TrimSpace(url) != "" {
				urlCopy := url
				item.AvatarDownloadURL = &urlCopy
			}
		}
		items = append(items, item)
	}

	return items, notFound, nil
}

func (s *ProfileService) UpdateProfileInfo(ctx context.Context, userID uuid.UUID, patch UpdateProfileInfoInput) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if patch.FirstName != nil {
		p.FirstName = normalizeOptionalString(patch.FirstName)
	}
	if patch.LastName != nil {
		p.LastName = normalizeOptionalString(patch.LastName)
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

type UpdateProfileInfoInput struct {
	FirstName *string
	LastName  *string
}

type UpdatePreferencesInput struct {
	Locale   *string
	Timezone *string
}

func (s *ProfileService) UpdatePreferences(ctx context.Context, userID uuid.UUID, patch UpdatePreferencesInput) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	p, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if patch.Locale != nil {
		locale, err := s.normalizeLocale(patch.Locale)
		if err != nil {
			return nil, err
		}
		p.Locale = locale
	}
	if patch.Timezone != nil {
		timezone, err := s.normalizeTimezone(patch.Timezone)
		if err != nil {
			return nil, err
		}
		p.Timezone = timezone
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
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

var localePattern = regexp.MustCompile(`^[a-z]{2}(-[a-z]{2})?$`)

func (s *ProfileService) normalizeLocale(raw *string) (string, error) {
	locale := strings.ToLower(strings.TrimSpace(s.cfg.DefaultLocale))
	if locale == "" {
		locale = "ru"
	}
	if raw != nil {
		candidate := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(*raw, "_", "-")))
		if candidate != "" {
			locale = candidate
		}
	}
	if !localePattern.MatchString(locale) {
		return "", ErrInvalidLocale
	}
	return locale, nil
}

func (s *ProfileService) normalizeTimezone(raw *string) (string, error) {
	timezone := strings.TrimSpace(s.cfg.DefaultTimezone)
	if timezone == "" {
		timezone = "UTC"
	}
	if raw != nil {
		candidate := strings.TrimSpace(*raw)
		if candidate != "" {
			timezone = candidate
		}
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", ErrInvalidTimezone
	}
	return timezone, nil
}

func uniqueUserIDs(userIDs []uuid.UUID) []uuid.UUID {
	unique := make([]uuid.UUID, 0, len(userIDs))
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for i := range userIDs {
		id := userIDs[i]
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
