package usecase

import (
	"context"
	"profile/internal/application/ports"
	"profile/internal/model"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var localePattern = regexp.MustCompile(`^[a-z]{2}(-[a-z]{2})?$`)

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

func (d *dependencies) normalizeLocale(raw *string) (string, error) {
	locale := strings.ToLower(strings.TrimSpace(d.config.DefaultLocale))
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

func (d *dependencies) normalizeTimezone(raw *string) (string, error) {
	timezone := strings.TrimSpace(d.config.DefaultTimezone)
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

func resolveAvatarDownloadURLs(ctx context.Context, files ports.FileGateway, profiles []model.Profile) map[uuid.UUID]string {
	avatarIDs := make([]uuid.UUID, 0, len(profiles))
	for i := range profiles {
		if profiles[i].AvatarFileID != nil {
			avatarIDs = append(avatarIDs, *profiles[i].AvatarFileID)
		}
	}
	if files == nil || len(avatarIDs) == 0 {
		return map[uuid.UUID]string{}
	}
	urls, err := files.BatchGetDownloadURLs(ctx, avatarIDs)
	if err != nil {
		return map[uuid.UUID]string{}
	}
	return urls
}
