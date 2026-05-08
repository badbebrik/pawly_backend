package usecase

import (
	"context"
	"errors"
	"profile/internal/application/ports"
	"profile/internal/domain/model"
	"testing"

	"github.com/google/uuid"
)

func TestCreateProfileNormalizesDataAndUsesDefaults(t *testing.T) {
	ctx := context.Background()
	set, repo, _ := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	profile, err := set.CreateProfile.Execute(ctx, CreateProfileParams{
		UserID:    userID,
		Locale:    profileStringPtr("RU_RU"),
		Timezone:  profileStringPtr("Europe/Moscow"),
		FirstName: profileStringPtr("  Ivan  "),
		LastName:  profileStringPtr("  Ivanov  "),
	})
	if err != nil {
		t.Fatalf("CreateProfile.Execute error: %v", err)
	}

	if profile.UserID != userID || profile.Locale != "ru-ru" || profile.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if profile.FirstName == nil || *profile.FirstName != "Ivan" || profile.LastName == nil || *profile.LastName != "Ivanov" {
		t.Fatalf("expected trimmed names, got first=%v last=%v", profile.FirstName, profile.LastName)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected profile create call, got %d", len(repo.created))
	}
}

func TestProfileUsecasesRejectBasicInvalidInput(t *testing.T) {
	ctx := context.Background()
	set, _, _ := baseProfileDeps()

	if _, err := set.CreateProfile.Execute(ctx, CreateProfileParams{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected create invalid input, got %v", err)
	}
	if _, err := set.GetProfile.Execute(ctx, uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected get invalid input, got %v", err)
	}
	if err := set.DeleteProfile.Execute(ctx, uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected delete invalid input, got %v", err)
	}
	if _, err := set.UpdateProfileInfo.Execute(ctx, uuid.Nil, UpdateProfileInfoParams{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected update info invalid input, got %v", err)
	}
	if _, err := set.UpdatePreferences.Execute(ctx, uuid.Nil, UpdatePreferencesParams{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected update preferences invalid input, got %v", err)
	}
	if _, err := set.GetPreferences.Execute(ctx, uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected preferences invalid input, got %v", err)
	}
	if _, _, err := set.InitAvatarUpload.Execute(ctx, uuid.Nil, "image/jpeg", nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected init avatar invalid input, got %v", err)
	}
	if _, err := set.ConfirmAvatarUpload.Execute(ctx, uuid.New(), uuid.New(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected confirm avatar invalid input, got %v", err)
	}
	if _, err := set.DeleteAvatar.Execute(ctx, uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected delete avatar invalid input, got %v", err)
	}
	if _, err := set.GetAvatarDownloadURL.Execute(ctx, uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected get avatar url invalid input, got %v", err)
	}
}

func TestCreateAndUpdatePreferencesValidateLocaleAndTimezone(t *testing.T) {
	ctx := context.Background()
	set, repo, _ := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo.profiles[userID] = *profileFixture(userID)

	if _, err := set.CreateProfile.Execute(ctx, CreateProfileParams{
		UserID: userID,
		Locale: profileStringPtr("bad-locale-value"),
	}); !errors.Is(err, ErrInvalidLocale) {
		t.Fatalf("expected invalid locale, got %v", err)
	}

	if _, err := set.UpdatePreferences.Execute(ctx, userID, UpdatePreferencesParams{
		Timezone: profileStringPtr("Bad/Timezone"),
	}); !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected invalid timezone, got %v", err)
	}

	profile, err := set.UpdatePreferences.Execute(ctx, userID, UpdatePreferencesParams{
		Locale:   profileStringPtr("EN_US"),
		Timezone: profileStringPtr("Europe/Moscow"),
	})
	if err != nil {
		t.Fatalf("UpdatePreferences.Execute error: %v", err)
	}
	if profile.Locale != "en-us" || profile.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected preferences: %+v", profile)
	}
}

func TestGetUpdateAndDeleteProfileUseRepository(t *testing.T) {
	ctx := context.Background()
	set, repo, _ := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo.profiles[userID] = *profileFixture(userID)

	got, err := set.GetProfile.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetProfile.Execute error: %v", err)
	}
	if got.FirstName == nil || *got.FirstName != "Ivan" {
		t.Fatalf("unexpected profile: %+v", got)
	}

	updated, err := set.UpdateProfileInfo.Execute(ctx, userID, UpdateProfileInfoParams{
		FirstName: profileStringPtr("  Petr  "),
		LastName:  profileStringPtr("   "),
	})
	if err != nil {
		t.Fatalf("UpdateProfileInfo.Execute error: %v", err)
	}
	if updated.FirstName == nil || *updated.FirstName != "Petr" || updated.LastName != nil {
		t.Fatalf("expected updated names, got first=%v last=%v", updated.FirstName, updated.LastName)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected update call, got %d", len(repo.updated))
	}

	if err := set.DeleteProfile.Execute(ctx, userID); err != nil {
		t.Fatalf("DeleteProfile.Execute error: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != userID {
		t.Fatalf("expected delete call for user, got %+v", repo.deleted)
	}
}

func TestBatchGetPreferencesDeduplicatesAndReturnsMissing(t *testing.T) {
	ctx := context.Background()
	set, repo, _ := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	missingID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	repo.profiles[userID] = *profileFixture(userID)

	items, notFound, err := set.BatchGetPreferences.Execute(ctx, []uuid.UUID{uuid.Nil, userID, userID, missingID})
	if err != nil {
		t.Fatalf("BatchGetPreferences.Execute error: %v", err)
	}
	if len(repo.getManyInput) != 2 || repo.getManyInput[0] != userID || repo.getManyInput[1] != missingID {
		t.Fatalf("expected deduplicated input, got %+v", repo.getManyInput)
	}
	if len(items) != 1 || items[0].UserID != userID || items[0].Locale != "ru" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if len(notFound) != 1 || notFound[0] != missingID {
		t.Fatalf("unexpected notFound: %+v", notFound)
	}
}

func TestBatchProfilesBriefResolvesAvatarURLs(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	missingID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	avatarID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	profile := profileFixture(userID)
	profile.AvatarFileID = &avatarID
	repo.profiles[userID] = *profile
	files.batchURLs[avatarID] = "https://files.example/avatar.jpg"

	items, notFound, err := set.BatchProfilesBrief.Execute(ctx, []uuid.UUID{userID, missingID, userID})
	if err != nil {
		t.Fatalf("BatchProfilesBrief.Execute error: %v", err)
	}
	if len(items) != 1 || items[0].AvatarDownloadURL == nil || *items[0].AvatarDownloadURL != "https://files.example/avatar.jpg" {
		t.Fatalf("unexpected brief items: %+v", items)
	}
	if len(notFound) != 1 || notFound[0] != missingID {
		t.Fatalf("unexpected notFound: %+v", notFound)
	}
	if len(files.batchInputs) != 1 || len(files.batchInputs[0]) != 1 || files.batchInputs[0][0] != avatarID {
		t.Fatalf("expected avatar URL batch lookup, got %+v", files.batchInputs)
	}
}

func TestAvatarUploadFlowReplacesOldAvatar(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	oldAvatarID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	newAvatarID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	profile := profileFixture(userID)
	profile.AvatarFileID = &oldAvatarID
	repo.profiles[userID] = *profile

	fileID, upload, err := set.InitAvatarUpload.Execute(ctx, userID, "image/jpeg", profileInt64Ptr(1024))
	if err != nil {
		t.Fatalf("InitAvatarUpload.Execute error: %v", err)
	}
	if fileID != newAvatarID || upload.URL == "" {
		t.Fatalf("unexpected upload response: fileID=%s upload=%+v", fileID, upload)
	}
	if len(files.initCalls) != 1 || files.initCalls[0].expectedSize != 1024 || files.initCalls[0].userID != userID {
		t.Fatalf("unexpected init calls: %+v", files.initCalls)
	}

	updated, err := set.ConfirmAvatarUpload.Execute(ctx, userID, newAvatarID, 900)
	if err != nil {
		t.Fatalf("ConfirmAvatarUpload.Execute error: %v", err)
	}
	if updated.AvatarFileID == nil || *updated.AvatarFileID != newAvatarID {
		t.Fatalf("expected new avatar, got %+v", updated.AvatarFileID)
	}
	if len(files.confirmCalls) != 1 || files.confirmCalls[0].fileID != newAvatarID {
		t.Fatalf("expected confirm call, got %+v", files.confirmCalls)
	}
	if len(files.linkCalls) != 1 || files.linkCalls[0].fileID != newAvatarID || files.linkCalls[0].userID != userID {
		t.Fatalf("expected link call, got %+v", files.linkCalls)
	}
	if len(files.unlinks) != 1 || files.unlinks[0].fileID != oldAvatarID || len(files.deletions) != 1 || files.deletions[0] != oldAvatarID {
		t.Fatalf("expected old avatar cleanup, unlinks=%+v deletions=%+v", files.unlinks, files.deletions)
	}
}

func TestConfirmAvatarUploadRollsBackLinkOnProfileUpdateError(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	avatarID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo.profiles[userID] = *profileFixture(userID)
	repo.updateErr = errors.New("update failed")

	errProfile, err := set.ConfirmAvatarUpload.Execute(ctx, userID, avatarID, 900)
	if !errors.Is(err, repo.updateErr) {
		t.Fatalf("expected repo update error, got profile=%+v err=%v", errProfile, err)
	}
	if len(files.unlinks) != 1 || files.unlinks[0].fileID != avatarID || files.unlinks[0].userID != userID {
		t.Fatalf("expected rollback unlink, got %+v", files.unlinks)
	}
	if len(files.deletions) != 0 {
		t.Fatalf("did not expect file deletion on rollback, got %+v", files.deletions)
	}
}

func TestDeleteAvatarClearsProfileAndCleansFile(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	avatarID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	profile := profileFixture(userID)
	profile.AvatarFileID = &avatarID
	repo.profiles[userID] = *profile

	updated, err := set.DeleteAvatar.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteAvatar.Execute error: %v", err)
	}
	if updated.AvatarFileID != nil {
		t.Fatalf("expected avatar to be cleared, got %+v", updated.AvatarFileID)
	}
	if len(files.unlinks) != 1 || files.unlinks[0].fileID != avatarID || len(files.deletions) != 1 || files.deletions[0] != avatarID {
		t.Fatalf("expected avatar cleanup, unlinks=%+v deletions=%+v", files.unlinks, files.deletions)
	}
}

func TestDeleteAvatarNoopsWhenProfileHasNoAvatar(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo.profiles[userID] = *profileFixture(userID)

	profile, err := set.DeleteAvatar.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteAvatar.Execute error: %v", err)
	}
	if profile.AvatarFileID != nil {
		t.Fatalf("expected no avatar, got %+v", profile.AvatarFileID)
	}
	if len(files.unlinks) != 0 || len(files.deletions) != 0 {
		t.Fatalf("did not expect file calls, unlinks=%+v deletions=%+v", files.unlinks, files.deletions)
	}
}

func TestAvatarDownloadURLMapsFileClientErrors(t *testing.T) {
	ctx := context.Background()
	set, _, files := baseProfileDeps()
	avatarID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	url, err := set.GetAvatarDownloadURL.Execute(ctx, avatarID)
	if err != nil {
		t.Fatalf("GetAvatarDownloadURL.Execute error: %v", err)
	}
	if url != "https://files.example/avatar.jpg" {
		t.Fatalf("unexpected download url: %q", url)
	}

	files.downloadURL = " "
	if _, err := set.GetAvatarDownloadURL.Execute(ctx, avatarID); !errors.Is(err, ErrAvatarUpload) {
		t.Fatalf("expected avatar upload error for empty URL, got %v", err)
	}
}

func TestProfileUsecasesMapRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	set, repo, files := baseProfileDeps()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo.profiles[userID] = *profileFixture(userID)

	repo.getErr = ports.ErrNotFound
	if _, err := set.GetProfile.Execute(ctx, userID); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("expected repo not found, got %v", err)
	}
	repo.getErr = nil

	repo.deleteErr = ports.ErrConflict
	if err := set.DeleteProfile.Execute(ctx, userID); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected repo conflict, got %v", err)
	}

	files.confirmErr = errors.New("confirm failed")
	if _, err := set.ConfirmAvatarUpload.Execute(ctx, userID, uuid.New(), 100); !errors.Is(err, ErrAvatarUpload) {
		t.Fatalf("expected avatar upload error, got %v", err)
	}
}

func profileFixture(userID uuid.UUID) *model.Profile {
	firstName := "Ivan"
	lastName := "Ivanov"
	return &model.Profile{
		UserID:    userID,
		FirstName: &firstName,
		LastName:  &lastName,
		Locale:    "ru",
		Timezone:  "UTC",
	}
}
