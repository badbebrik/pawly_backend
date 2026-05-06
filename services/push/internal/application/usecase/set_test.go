package usecase

import (
	"context"
	"errors"
	"push/internal/application/ports"
	"push/internal/domain/model"
	"testing"

	"github.com/google/uuid"
)

type stubPushRepo struct {
	upsertDeviceTokenFn      func(context.Context, ports.UpsertDeviceTokenParams) (*model.DeviceToken, error)
	deleteDeviceTokenFn      func(context.Context, ports.DeleteDeviceTokenParams) error
	listDeviceTokensByUserFn func(context.Context, uuid.UUID) ([]model.DeviceToken, error)
	getPetPushSettingsFn     func(context.Context, uuid.UUID, uuid.UUID) (*model.PetPushSettings, error)
	upsertPetPushSettingsFn  func(context.Context, ports.UpsertPetPushSettingsParams) (*model.PetPushSettings, error)
}

func (s *stubPushRepo) UpsertDeviceToken(ctx context.Context, in ports.UpsertDeviceTokenParams) (*model.DeviceToken, error) {
	if s.upsertDeviceTokenFn != nil {
		return s.upsertDeviceTokenFn(ctx, in)
	}
	return &model.DeviceToken{
		ID:        in.ID,
		UserID:    in.UserID,
		DeviceID:  in.DeviceID,
		Platform:  in.Platform,
		PushToken: in.PushToken,
	}, nil
}

func (s *stubPushRepo) DeleteDeviceToken(ctx context.Context, in ports.DeleteDeviceTokenParams) error {
	if s.deleteDeviceTokenFn != nil {
		return s.deleteDeviceTokenFn(ctx, in)
	}
	return nil
}

func (s *stubPushRepo) ListDeviceTokensByUser(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	if s.listDeviceTokensByUserFn != nil {
		return s.listDeviceTokensByUserFn(ctx, userID)
	}
	return []model.DeviceToken{}, nil
}

func (s *stubPushRepo) GetPetPushSettings(ctx context.Context, userID, petID uuid.UUID) (*model.PetPushSettings, error) {
	if s.getPetPushSettingsFn != nil {
		return s.getPetPushSettingsFn(ctx, userID, petID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubPushRepo) UpsertPetPushSettings(ctx context.Context, in ports.UpsertPetPushSettingsParams) (*model.PetPushSettings, error) {
	if s.upsertPetPushSettingsFn != nil {
		return s.upsertPetPushSettingsFn(ctx, in)
	}
	return &model.PetPushSettings{
		UserID:                in.UserID,
		PetID:                 in.PetID,
		ScheduledItemsEnabled: in.ScheduledItemsEnabled,
	}, nil
}

func expectPushErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v want %v", got, want)
	}
}

func TestRegisterDeviceToken_UpsertsNormalizedToken(t *testing.T) {
	userID := uuid.New()
	var upsert ports.UpsertDeviceTokenParams
	set := New(&stubPushRepo{
		upsertDeviceTokenFn: func(_ context.Context, in ports.UpsertDeviceTokenParams) (*model.DeviceToken, error) {
			upsert = in
			return &model.DeviceToken{
				ID:        in.ID,
				UserID:    in.UserID,
				DeviceID:  in.DeviceID,
				Platform:  in.Platform,
				PushToken: in.PushToken,
			}, nil
		},
	})

	out, err := set.RegisterDeviceToken(context.Background(), RegisterDeviceTokenParams{
		UserID:    userID,
		DeviceID:  " phone ",
		Platform:  " ios ",
		PushToken: " token ",
	})
	if err != nil {
		t.Fatalf("RegisterDeviceToken returned error: %v", err)
	}
	if upsert.ID == uuid.Nil || upsert.UserID != userID {
		t.Fatalf("unexpected upsert identity: %+v", upsert)
	}
	if upsert.DeviceID != "phone" || upsert.Platform != model.PlatformIOS || upsert.PushToken != "token" {
		t.Fatalf("unexpected normalized token: %+v", upsert)
	}
	if out.DeviceID != "phone" || out.Platform != model.PlatformIOS || out.PushToken != "token" {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestRegisterDeviceTokenRejectsInvalidInput(t *testing.T) {
	set := New(&stubPushRepo{})

	tests := []struct {
		name string
		in   RegisterDeviceTokenParams
		want error
	}{
		{name: "missing user", in: RegisterDeviceTokenParams{DeviceID: "phone", Platform: "ios", PushToken: "token"}, want: ErrForbidden},
		{name: "missing device", in: RegisterDeviceTokenParams{UserID: uuid.New(), Platform: "ios", PushToken: "token"}, want: ErrInvalidInput},
		{name: "missing token", in: RegisterDeviceTokenParams{UserID: uuid.New(), DeviceID: "phone", Platform: "ios"}, want: ErrInvalidInput},
		{name: "bad platform", in: RegisterDeviceTokenParams{UserID: uuid.New(), DeviceID: "phone", Platform: "web", PushToken: "token"}, want: ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := set.RegisterDeviceToken(context.Background(), tt.in)
			expectPushErr(t, err, tt.want)
		})
	}
}

func TestDeleteDeviceToken_DeletesTrimmedDeviceID(t *testing.T) {
	userID := uuid.New()
	var deleted ports.DeleteDeviceTokenParams
	set := New(&stubPushRepo{
		deleteDeviceTokenFn: func(_ context.Context, in ports.DeleteDeviceTokenParams) error {
			deleted = in
			return nil
		},
	})

	err := set.DeleteDeviceToken(context.Background(), DeleteDeviceTokenParams{
		UserID:   userID,
		DeviceID: " phone ",
	})
	if err != nil {
		t.Fatalf("DeleteDeviceToken returned error: %v", err)
	}
	if deleted.UserID != userID || deleted.DeviceID != "phone" {
		t.Fatalf("unexpected delete params: %+v", deleted)
	}
}

func TestGetPetPushSettings_ReturnsDefaultWhenMissing(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	set := New(&stubPushRepo{})

	out, err := set.GetPetPushSettings(context.Background(), GetPetPushSettingsParams{
		UserID: userID,
		PetID:  petID,
	})
	if err != nil {
		t.Fatalf("GetPetPushSettings returned error: %v", err)
	}
	if out.UserID != userID || out.PetID != petID || !out.ScheduledItemsEnabled {
		t.Fatalf("unexpected default settings: %+v", out)
	}
}

func TestUpdatePetPushSettings_UpsertsSettings(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var upsert ports.UpsertPetPushSettingsParams
	set := New(&stubPushRepo{
		upsertPetPushSettingsFn: func(_ context.Context, in ports.UpsertPetPushSettingsParams) (*model.PetPushSettings, error) {
			upsert = in
			return &model.PetPushSettings{
				UserID:                in.UserID,
				PetID:                 in.PetID,
				ScheduledItemsEnabled: in.ScheduledItemsEnabled,
			}, nil
		},
	})

	out, err := set.UpdatePetPushSettings(context.Background(), UpdatePetPushSettingsParams{
		UserID:                userID,
		PetID:                 petID,
		ScheduledItemsEnabled: false,
	})
	if err != nil {
		t.Fatalf("UpdatePetPushSettings returned error: %v", err)
	}
	if upsert.UserID != userID || upsert.PetID != petID || upsert.ScheduledItemsEnabled {
		t.Fatalf("unexpected upsert params: %+v", upsert)
	}
	if out.ScheduledItemsEnabled {
		t.Fatalf("unexpected settings result: %+v", out)
	}
}

func TestListEligibleDeviceTokens_RespectsSettings(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	device := model.DeviceToken{ID: uuid.New(), UserID: userID, DeviceID: "phone", Platform: model.PlatformIOS, PushToken: "token"}

	t.Run("enabled", func(t *testing.T) {
		set := New(&stubPushRepo{
			getPetPushSettingsFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.PetPushSettings, error) {
				return &model.PetPushSettings{UserID: userID, PetID: petID, ScheduledItemsEnabled: true}, nil
			},
			listDeviceTokensByUserFn: func(_ context.Context, gotUserID uuid.UUID) ([]model.DeviceToken, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected user id: %s", gotUserID)
				}
				return []model.DeviceToken{device}, nil
			},
		})

		out, err := set.ListEligibleDeviceTokens(context.Background(), ListEligibleDeviceTokensParams{UserID: userID, PetID: petID})
		if err != nil {
			t.Fatalf("ListEligibleDeviceTokens returned error: %v", err)
		}
		if len(out) != 1 || out[0].DeviceID != "phone" {
			t.Fatalf("unexpected tokens: %+v", out)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		calledList := false
		set := New(&stubPushRepo{
			getPetPushSettingsFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.PetPushSettings, error) {
				return &model.PetPushSettings{UserID: userID, PetID: petID, ScheduledItemsEnabled: false}, nil
			},
			listDeviceTokensByUserFn: func(context.Context, uuid.UUID) ([]model.DeviceToken, error) {
				calledList = true
				return []model.DeviceToken{device}, nil
			},
		})

		out, err := set.ListEligibleDeviceTokens(context.Background(), ListEligibleDeviceTokensParams{UserID: userID, PetID: petID})
		if err != nil {
			t.Fatalf("ListEligibleDeviceTokens returned error: %v", err)
		}
		if len(out) != 0 {
			t.Fatalf("expected no tokens, got %+v", out)
		}
		if calledList {
			t.Fatal("expected device token list not to be called")
		}
	})
}

func TestMapRepoError(t *testing.T) {
	if err := mapRepoError(ports.ErrNotFound); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mapRepoError(ports.ErrConflict); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

var _ ports.PushRepository = (*stubPushRepo)(nil)
