package usecase

import (
	"context"
	"profile/internal/model"
	"strings"

	"github.com/google/uuid"
)

type Preferences struct {
	UserID   uuid.UUID
	Locale   string
	Timezone string
}

type GetPreferencesUseCase struct{ deps *dependencies }
type BatchGetPreferencesUseCase struct{ deps *dependencies }
type BatchProfilesBriefUseCase struct{ deps *dependencies }

func (uc *GetPreferencesUseCase) Execute(ctx context.Context, userID uuid.UUID) (*Preferences, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Preferences{UserID: profile.UserID, Locale: profile.Locale, Timezone: profile.Timezone}, nil
}

func (uc *BatchGetPreferencesUseCase) Execute(ctx context.Context, userIDs []uuid.UUID) ([]Preferences, []uuid.UUID, error) {
	unique := uniqueUserIDs(userIDs)
	if len(unique) == 0 {
		return []Preferences{}, []uuid.UUID{}, nil
	}
	profiles, err := uc.deps.profiles.GetByUserIDs(ctx, unique)
	if err != nil {
		return nil, nil, err
	}
	byUser := make(map[uuid.UUID]model.Profile, len(profiles))
	for i := range profiles {
		byUser[profiles[i].UserID] = profiles[i]
	}
	items := make([]Preferences, 0, len(byUser))
	notFound := make([]uuid.UUID, 0)
	for i := range unique {
		p, ok := byUser[unique[i]]
		if !ok {
			notFound = append(notFound, unique[i])
			continue
		}
		items = append(items, Preferences{UserID: p.UserID, Locale: p.Locale, Timezone: p.Timezone})
	}
	return items, notFound, nil
}

type ProfileBrief struct {
	UserID            uuid.UUID
	FirstName         *string
	LastName          *string
	AvatarFileID      *uuid.UUID
	AvatarDownloadURL *string
}

func (uc *BatchProfilesBriefUseCase) Execute(ctx context.Context, userIDs []uuid.UUID) ([]ProfileBrief, []uuid.UUID, error) {
	unique := uniqueUserIDs(userIDs)
	if len(unique) == 0 {
		return []ProfileBrief{}, []uuid.UUID{}, nil
	}
	profiles, err := uc.deps.profiles.GetByUserIDs(ctx, unique)
	if err != nil {
		return nil, nil, err
	}
	byUser := make(map[uuid.UUID]model.Profile, len(profiles))
	for i := range profiles {
		byUser[profiles[i].UserID] = profiles[i]
	}
	avatarURLByID := resolveAvatarDownloadURLs(ctx, uc.deps.files, profiles)
	items := make([]ProfileBrief, 0, len(byUser))
	notFound := make([]uuid.UUID, 0)
	for i := range unique {
		p, ok := byUser[unique[i]]
		if !ok {
			notFound = append(notFound, unique[i])
			continue
		}
		item := ProfileBrief{
			UserID:       p.UserID,
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			AvatarFileID: p.AvatarFileID,
		}
		if p.AvatarFileID != nil {
			if url := strings.TrimSpace(avatarURLByID[*p.AvatarFileID]); url != "" {
				urlCopy := url
				item.AvatarDownloadURL = &urlCopy
			}
		}
		items = append(items, item)
	}
	return items, notFound, nil
}
