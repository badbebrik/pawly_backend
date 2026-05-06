package usecase

import (
	"context"
	"errors"
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubPetRepo struct {
	createFn           func(context.Context, ports.CreatePetInput) (*model.Pet, error)
	deleteByIDFn       func(context.Context, uuid.UUID) error
	getByIDFn          func(context.Context, uuid.UUID) (*model.Pet, error)
	listByIDsFn        func(context.Context, []uuid.UUID, bool, int, int) ([]model.Pet, int, error)
	listSpeciesFn      func(context.Context) ([]model.Species, error)
	listBreedsFn       func(context.Context) ([]model.Breed, error)
	listPatternsFn     func(context.Context) ([]model.Pattern, error)
	listColorPresetsFn func(context.Context) ([]model.ColorPreset, error)
	updateFn           func(context.Context, uuid.UUID, int, model.Pet) (*model.Pet, error)
	updateOwnerFn      func(context.Context, uuid.UUID, int, uuid.UUID) (*model.Pet, error)
	updatePhotoFn      func(context.Context, uuid.UUID, int, *uuid.UUID) (*model.Pet, error)
	updateStatusFn     func(context.Context, uuid.UUID, int, string, *time.Time, *time.Time) (*model.Pet, error)
}

func (s *stubPetRepo) Create(ctx context.Context, in ports.CreatePetInput) (*model.Pet, error) {
	if s.createFn != nil {
		return s.createFn(ctx, in)
	}
	return &in.Pet, nil
}

func (s *stubPetRepo) DeleteByID(ctx context.Context, petID uuid.UUID) error {
	if s.deleteByIDFn != nil {
		return s.deleteByIDFn(ctx, petID)
	}
	return nil
}

func (s *stubPetRepo) GetByID(ctx context.Context, petID uuid.UUID) (*model.Pet, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, petID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubPetRepo) ListByIDs(ctx context.Context, ids []uuid.UUID, includeArchived bool, offset, limit int) ([]model.Pet, int, error) {
	if s.listByIDsFn != nil {
		return s.listByIDsFn(ctx, ids, includeArchived, offset, limit)
	}
	return []model.Pet{}, 0, nil
}

func (s *stubPetRepo) ListSpecies(ctx context.Context) ([]model.Species, error) {
	if s.listSpeciesFn != nil {
		return s.listSpeciesFn(ctx)
	}
	return []model.Species{}, nil
}

func (s *stubPetRepo) ListBreeds(ctx context.Context) ([]model.Breed, error) {
	if s.listBreedsFn != nil {
		return s.listBreedsFn(ctx)
	}
	return []model.Breed{}, nil
}

func (s *stubPetRepo) ListPatterns(ctx context.Context) ([]model.Pattern, error) {
	if s.listPatternsFn != nil {
		return s.listPatternsFn(ctx)
	}
	return []model.Pattern{}, nil
}

func (s *stubPetRepo) ListColorPresets(ctx context.Context) ([]model.ColorPreset, error) {
	if s.listColorPresetsFn != nil {
		return s.listColorPresetsFn(ctx)
	}
	return []model.ColorPreset{}, nil
}

func (s *stubPetRepo) Update(ctx context.Context, petID uuid.UUID, rowVersion int, pet model.Pet) (*model.Pet, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, petID, rowVersion, pet)
	}
	pet.ID = petID
	pet.RowVersion = rowVersion + 1
	return &pet, nil
}

func (s *stubPetRepo) UpdateOwner(ctx context.Context, petID uuid.UUID, rowVersion int, ownerUserID uuid.UUID) (*model.Pet, error) {
	if s.updateOwnerFn != nil {
		return s.updateOwnerFn(ctx, petID, rowVersion, ownerUserID)
	}
	return &model.Pet{ID: petID, OwnerUserID: ownerUserID, RowVersion: rowVersion + 1}, nil
}

func (s *stubPetRepo) UpdatePhoto(ctx context.Context, petID uuid.UUID, rowVersion int, fileID *uuid.UUID) (*model.Pet, error) {
	if s.updatePhotoFn != nil {
		return s.updatePhotoFn(ctx, petID, rowVersion, fileID)
	}
	return &model.Pet{ID: petID, RowVersion: rowVersion + 1, ProfilePhotoFileID: fileID}, nil
}

func (s *stubPetRepo) UpdateStatus(ctx context.Context, petID uuid.UUID, rowVersion int, status string, missingSince *time.Time, archivedAt *time.Time) (*model.Pet, error) {
	if s.updateStatusFn != nil {
		return s.updateStatusFn(ctx, petID, rowVersion, status, missingSince, archivedAt)
	}
	return &model.Pet{ID: petID, RowVersion: rowVersion + 1, Status: status, MissingSince: missingSince, ArchivedAt: archivedAt}, nil
}

type stubACLClient struct {
	checkFn                 func(context.Context, uuid.UUID, uuid.UUID, string) (bool, error)
	listPetsForUserFn       func(context.Context, uuid.UUID) ([]ACLMembership, error)
	createOwnerMembershipFn func(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	transferOwnershipFn     func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (ACLTransferOwnershipResult, error)
}

func (s *stubACLClient) Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error) {
	if s.checkFn != nil {
		return s.checkFn(ctx, petID, userID, action)
	}
	return true, nil
}

func (s *stubACLClient) ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]ACLMembership, error) {
	if s.listPetsForUserFn != nil {
		return s.listPetsForUserFn(ctx, userID)
	}
	return []ACLMembership{}, nil
}

func (s *stubACLClient) CreateOwnerMembership(ctx context.Context, petID, userID uuid.UUID) (uuid.UUID, error) {
	if s.createOwnerMembershipFn != nil {
		return s.createOwnerMembershipFn(ctx, petID, userID)
	}
	return uuid.New(), nil
}

func (s *stubACLClient) TransferOwnership(ctx context.Context, petID, requesterUserID, targetMemberID uuid.UUID) (ACLTransferOwnershipResult, error) {
	if s.transferOwnershipFn != nil {
		return s.transferOwnershipFn(ctx, petID, requesterUserID, targetMemberID)
	}
	return ACLTransferOwnershipResult{CurrentOwnerUserID: uuid.New(), PreviousOwnerMemberID: uuid.New()}, nil
}

type stubFileClient struct {
	initUploadFn           func(context.Context, string, int64, string) (uuid.UUID, UploadInfo, error)
	confirmUploadFn        func(context.Context, uuid.UUID, int64) error
	getDownloadURLFn       func(context.Context, uuid.UUID) (string, time.Time, error)
	batchGetDownloadURLsFn func(context.Context, []uuid.UUID) (map[uuid.UUID]string, error)
	linkPetAvatarFn        func(context.Context, uuid.UUID, uuid.UUID) error
	unlinkPetAvatarFn      func(context.Context, uuid.UUID, uuid.UUID) error
	deleteFileIfUnlinkedFn func(context.Context, uuid.UUID) error
}

func (s *stubFileClient) InitUpload(ctx context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, UploadInfo, error) {
	if s.initUploadFn != nil {
		return s.initUploadFn(ctx, mimeType, expectedSize, originalFilename)
	}
	return uuid.New(), UploadInfo{Method: "PUT", URL: "upload-url"}, nil
}

func (s *stubFileClient) ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error {
	if s.confirmUploadFn != nil {
		return s.confirmUploadFn(ctx, fileID, sizeBytes)
	}
	return nil
}

func (s *stubFileClient) GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error) {
	if s.getDownloadURLFn != nil {
		return s.getDownloadURLFn(ctx, fileID)
	}
	return "download-url", time.Now().Add(time.Hour), nil
}

func (s *stubFileClient) BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if s.batchGetDownloadURLsFn != nil {
		return s.batchGetDownloadURLsFn(ctx, fileIDs)
	}
	return map[uuid.UUID]string{}, nil
}

func (s *stubFileClient) LinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error {
	if s.linkPetAvatarFn != nil {
		return s.linkPetAvatarFn(ctx, fileID, petID)
	}
	return nil
}

func (s *stubFileClient) UnlinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error {
	if s.unlinkPetAvatarFn != nil {
		return s.unlinkPetAvatarFn(ctx, fileID, petID)
	}
	return nil
}

func (s *stubFileClient) DeleteFileIfUnlinked(ctx context.Context, fileID uuid.UUID) error {
	if s.deleteFileIfUnlinkedFn != nil {
		return s.deleteFileIfUnlinkedFn(ctx, fileID)
	}
	return nil
}

func newTestPet(repo ports.PetRepository, acl ports.ACLClient, file ports.FileClient) *Pet {
	return New(repo, acl, file)
}

func expectPetErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v want %v", got, want)
	}
}

func TestPetUsecases_RejectBasicInvalidInput(t *testing.T) {
	uc := newTestPet(&stubPetRepo{}, &stubACLClient{}, &stubFileClient{})
	userID := uuid.New()
	petID := uuid.New()

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "create missing user",
			run: func() error {
				_, err := uc.CreatePet(context.Background(), CreatePetParams{Name: "Barsik", CustomSpeciesName: stringPtr("cat")})
				return err
			},
		},
		{
			name: "create missing species choice",
			run: func() error {
				_, err := uc.CreatePet(context.Background(), CreatePetParams{UserID: userID, Name: "Barsik"})
				return err
			},
		},
		{
			name: "update missing row version",
			run: func() error {
				_, err := uc.UpdatePet(context.Background(), UpdatePetParams{UserID: userID, PetID: petID, Name: "Barsik", CustomSpeciesName: stringPtr("cat")})
				return err
			},
		},
		{
			name: "missing status without date",
			run: func() error {
				_, err := uc.ChangePetStatus(context.Background(), ChangePetStatusParams{UserID: userID, PetID: petID, RowVersion: 1, Status: "MISSING"})
				return err
			},
		},
		{
			name: "transfer missing target member",
			run: func() error {
				_, err := uc.TransferOwnership(context.Background(), TransferPetOwnershipParams{UserID: userID, PetID: petID, RowVersion: 1})
				return err
			},
		},
		{
			name: "photo upload invalid mime",
			run: func() error {
				_, _, err := uc.InitPetPhotoUpload(context.Background(), InitPetPhotoUploadParams{
					UserID: userID, PetID: petID, MimeType: "text/plain", ExpectedSizeBytes: 10,
				})
				return err
			},
		},
		{
			name: "confirm photo missing file",
			run: func() error {
				_, err := uc.ConfirmPetPhotoUpload(context.Background(), ConfirmPetPhotoUploadParams{
					UserID: userID, PetID: petID, RowVersion: 1, SizeBytes: 10,
				})
				return err
			},
		},
		{
			name: "get missing user",
			run: func() error {
				_, err := uc.GetPet(context.Background(), uuid.Nil, petID)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectPetErr(t, tt.run(), ErrInvalidInput)
		})
	}
}

func TestCreatePet_CreatesPetAndOwnerMembership(t *testing.T) {
	userID := uuid.New()
	var createdPet model.Pet
	var ownerPetID uuid.UUID
	var ownerUserID uuid.UUID

	uc := newTestPet(&stubPetRepo{
		createFn: func(_ context.Context, in ports.CreatePetInput) (*model.Pet, error) {
			createdPet = in.Pet
			return &in.Pet, nil
		},
	}, &stubACLClient{
		createOwnerMembershipFn: func(_ context.Context, petID, uid uuid.UUID) (uuid.UUID, error) {
			ownerPetID = petID
			ownerUserID = uid
			return uuid.New(), nil
		},
	}, &stubFileClient{})

	out, err := uc.CreatePet(context.Background(), CreatePetParams{
		UserID:            userID,
		Name:              "  Barsik  ",
		CustomSpeciesName: stringPtr(" Cat "),
	})
	if err != nil {
		t.Fatalf("CreatePet returned error: %v", err)
	}
	if out.ID == uuid.Nil {
		t.Fatal("expected generated pet id")
	}
	if createdPet.OwnerUserID != userID || createdPet.Name != "Barsik" || createdPet.Status != "ACTIVE" {
		t.Fatalf("unexpected created pet: %+v", createdPet)
	}
	if createdPet.CustomSpeciesName == nil || *createdPet.CustomSpeciesName != "Cat" {
		t.Fatalf("unexpected custom species: %v", createdPet.CustomSpeciesName)
	}
	if ownerPetID != out.ID || ownerUserID != userID {
		t.Fatalf("unexpected owner membership args: pet=%s user=%s", ownerPetID, ownerUserID)
	}
}

func TestGetPet_ChecksReadAccess(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var checkedAction string

	uc := newTestPet(&stubPetRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.Pet, error) {
			return &model.Pet{ID: petID, Name: "Barsik"}, nil
		},
	}, &stubACLClient{
		checkFn: func(_ context.Context, gotPetID, gotUserID uuid.UUID, action string) (bool, error) {
			if gotPetID != petID || gotUserID != userID {
				t.Fatalf("unexpected acl args: pet=%s user=%s", gotPetID, gotUserID)
			}
			checkedAction = action
			return true, nil
		},
	}, &stubFileClient{})

	out, err := uc.GetPet(context.Background(), userID, petID)
	if err != nil {
		t.Fatalf("GetPet returned error: %v", err)
	}
	if out.ID != petID {
		t.Fatalf("unexpected pet id: %s", out.ID)
	}
	if checkedAction != ActionPetRead {
		t.Fatalf("unexpected acl action: %s", checkedAction)
	}
}

func TestListPets_ReturnsAccessAndPhotoURL(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	memberID := uuid.New()
	fileID := uuid.New()

	uc := newTestPet(&stubPetRepo{
		listByIDsFn: func(_ context.Context, ids []uuid.UUID, includeArchived bool, offset, limit int) ([]model.Pet, int, error) {
			if len(ids) != 1 || ids[0] != petID || includeArchived || offset != 0 || limit != 50 {
				t.Fatalf("unexpected list args: ids=%v include=%v offset=%d limit=%d", ids, includeArchived, offset, limit)
			}
			return []model.Pet{{ID: petID, Name: "Barsik", ProfilePhotoFileID: &fileID}}, 1, nil
		},
	}, &stubACLClient{
		listPetsForUserFn: func(context.Context, uuid.UUID) ([]ACLMembership, error) {
			return []ACLMembership{{PetID: petID, MemberID: memberID}}, nil
		},
	}, &stubFileClient{
		batchGetDownloadURLsFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]string, error) {
			return map[uuid.UUID]string{fileID: "https://cdn.example/avatar.png"}, nil
		},
	})

	items, total, err := uc.ListPets(context.Background(), ListPetsParams{UserID: userID, Limit: -1})
	if err != nil {
		t.Fatalf("ListPets returned error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("unexpected list result: total=%d items=%d", total, len(items))
	}
	if items[0].MyAccess == nil || items[0].MyAccess.MemberID != memberID {
		t.Fatalf("unexpected access: %+v", items[0].MyAccess)
	}
	if items[0].ProfilePhotoDownloadURL == nil || *items[0].ProfilePhotoDownloadURL == "" {
		t.Fatal("expected profile photo download url")
	}
}

func TestUpdatePet_ChecksWriteAccessAndUpdatesPet(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var updatedName string

	uc := newTestPet(&stubPetRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.Pet, error) {
			return &model.Pet{ID: petID, Status: "ACTIVE"}, nil
		},
		updateFn: func(_ context.Context, gotPetID uuid.UUID, rowVersion int, pet model.Pet) (*model.Pet, error) {
			if gotPetID != petID || rowVersion != 2 {
				t.Fatalf("unexpected update args: pet=%s row=%d", gotPetID, rowVersion)
			}
			updatedName = pet.Name
			pet.ID = gotPetID
			pet.RowVersion = rowVersion + 1
			return &pet, nil
		},
	}, &stubACLClient{}, &stubFileClient{})

	out, err := uc.UpdatePet(context.Background(), UpdatePetParams{
		UserID:            userID,
		PetID:             petID,
		RowVersion:        2,
		Name:              "  Barsik  ",
		CustomSpeciesName: stringPtr("cat"),
	})
	if err != nil {
		t.Fatalf("UpdatePet returned error: %v", err)
	}
	if updatedName != "Barsik" || out.RowVersion != 3 {
		t.Fatalf("unexpected update result: name=%s out=%+v", updatedName, out)
	}
}

func TestChangePetStatus_UpdatesMissingAndArchivedStates(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	missingSince := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	uc := newTestPet(&stubPetRepo{}, &stubACLClient{}, &stubFileClient{})

	missing, err := uc.ChangePetStatus(context.Background(), ChangePetStatusParams{
		UserID: userID, PetID: petID, RowVersion: 1, Status: "missing", MissingSince: &missingSince,
	})
	if err != nil {
		t.Fatalf("ChangePetStatus missing returned error: %v", err)
	}
	if missing.Status != "MISSING" || missing.MissingSince == nil || !missing.MissingSince.Equal(missingSince) {
		t.Fatalf("unexpected missing pet: %+v", missing)
	}

	archived, err := uc.ChangePetStatus(context.Background(), ChangePetStatusParams{
		UserID: userID, PetID: petID, RowVersion: 2, Status: "archived",
	})
	if err != nil {
		t.Fatalf("ChangePetStatus archived returned error: %v", err)
	}
	if archived.Status != "ARCHIVED" || archived.ArchivedAt == nil {
		t.Fatalf("unexpected archived pet: %+v", archived)
	}
}

func TestTransferOwnership_UpdatesOwnerFromACLResult(t *testing.T) {
	userID := uuid.New()
	newOwnerID := uuid.New()
	petID := uuid.New()
	targetMemberID := uuid.New()

	uc := newTestPet(&stubPetRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.Pet, error) {
			return &model.Pet{ID: petID, OwnerUserID: userID, RowVersion: 1}, nil
		},
		updateOwnerFn: func(_ context.Context, gotPetID uuid.UUID, rowVersion int, ownerUserID uuid.UUID) (*model.Pet, error) {
			if gotPetID != petID || rowVersion != 1 || ownerUserID != newOwnerID {
				t.Fatalf("unexpected update owner args: pet=%s row=%d owner=%s", gotPetID, rowVersion, ownerUserID)
			}
			return &model.Pet{ID: petID, OwnerUserID: ownerUserID, RowVersion: 2}, nil
		},
	}, &stubACLClient{
		transferOwnershipFn: func(_ context.Context, gotPetID, requesterID, gotTargetMemberID uuid.UUID) (ACLTransferOwnershipResult, error) {
			if gotPetID != petID || requesterID != userID || gotTargetMemberID != targetMemberID {
				t.Fatalf("unexpected transfer args: pet=%s requester=%s target=%s", gotPetID, requesterID, gotTargetMemberID)
			}
			return ACLTransferOwnershipResult{
				PreviousOwnerMemberID: uuid.New(),
				PreviousOwnerUserID:   userID,
				CurrentOwnerMemberID:  targetMemberID,
				CurrentOwnerUserID:    newOwnerID,
			}, nil
		},
	}, &stubFileClient{})

	out, err := uc.TransferOwnership(context.Background(), TransferPetOwnershipParams{
		UserID: userID, PetID: petID, RowVersion: 1, TargetMemberID: targetMemberID,
	})
	if err != nil {
		t.Fatalf("TransferOwnership returned error: %v", err)
	}
	if out.OwnerUserID != newOwnerID {
		t.Fatalf("unexpected owner: %s", out.OwnerUserID)
	}
}

func TestPetPhotoUploadFlow_BasicSuccess(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	fileID := uuid.New()
	var confirmedFileID uuid.UUID
	var linkedFileID uuid.UUID
	var unlinkedOldFileID uuid.UUID
	var deletedOldFileID uuid.UUID
	oldFileID := uuid.New()

	uc := newTestPet(&stubPetRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.Pet, error) {
			return &model.Pet{ID: petID, Status: "ACTIVE", ProfilePhotoFileID: &oldFileID}, nil
		},
		updatePhotoFn: func(_ context.Context, gotPetID uuid.UUID, rowVersion int, gotFileID *uuid.UUID) (*model.Pet, error) {
			if gotPetID != petID || rowVersion != 1 || gotFileID == nil || *gotFileID != fileID {
				t.Fatalf("unexpected update photo args: pet=%s row=%d file=%v", gotPetID, rowVersion, gotFileID)
			}
			return &model.Pet{ID: petID, RowVersion: 2, ProfilePhotoFileID: gotFileID}, nil
		},
	}, &stubACLClient{}, &stubFileClient{
		initUploadFn: func(_ context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, UploadInfo, error) {
			if mimeType != "image/png" || expectedSize != 1024 || originalFilename != "avatar.png" {
				t.Fatalf("unexpected init upload args: %s %d %s", mimeType, expectedSize, originalFilename)
			}
			return fileID, UploadInfo{Method: "PUT", URL: "upload-url"}, nil
		},
		confirmUploadFn: func(_ context.Context, gotFileID uuid.UUID, sizeBytes int64) error {
			confirmedFileID = gotFileID
			if sizeBytes != 1024 {
				t.Fatalf("unexpected confirmed size: %d", sizeBytes)
			}
			return nil
		},
		linkPetAvatarFn: func(_ context.Context, gotFileID, gotPetID uuid.UUID) error {
			if gotPetID != petID {
				t.Fatalf("unexpected link pet id: %s", gotPetID)
			}
			linkedFileID = gotFileID
			return nil
		},
		unlinkPetAvatarFn: func(_ context.Context, gotFileID, gotPetID uuid.UUID) error {
			if gotPetID != petID {
				t.Fatalf("unexpected unlink pet id: %s", gotPetID)
			}
			unlinkedOldFileID = gotFileID
			return nil
		},
		deleteFileIfUnlinkedFn: func(_ context.Context, gotFileID uuid.UUID) error {
			deletedOldFileID = gotFileID
			return nil
		},
	})

	createdFileID, upload, err := uc.InitPetPhotoUpload(context.Background(), InitPetPhotoUploadParams{
		UserID: userID, PetID: petID, MimeType: "IMAGE/PNG", OriginalFilename: "avatar.png", ExpectedSizeBytes: 1024,
	})
	if err != nil {
		t.Fatalf("InitPetPhotoUpload returned error: %v", err)
	}
	if createdFileID != fileID || upload.Method != "PUT" {
		t.Fatalf("unexpected upload result: file=%s upload=%+v", createdFileID, upload)
	}

	out, err := uc.ConfirmPetPhotoUpload(context.Background(), ConfirmPetPhotoUploadParams{
		UserID: userID, PetID: petID, RowVersion: 1, FileID: fileID, SizeBytes: 1024,
	})
	if err != nil {
		t.Fatalf("ConfirmPetPhotoUpload returned error: %v", err)
	}
	if confirmedFileID != fileID || linkedFileID != fileID {
		t.Fatalf("expected file to be confirmed and linked: confirmed=%s linked=%s", confirmedFileID, linkedFileID)
	}
	if unlinkedOldFileID != oldFileID || deletedOldFileID != oldFileID {
		t.Fatalf("expected old file cleanup: unlinked=%s deleted=%s", unlinkedOldFileID, deletedOldFileID)
	}
	if out.ProfilePhotoFileID == nil || *out.ProfilePhotoFileID != fileID {
		t.Fatalf("unexpected updated photo: %+v", out)
	}
}

func TestDeletePetPhoto_ClearsExistingPhoto(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	oldFileID := uuid.New()
	var updatedWithNil bool
	var unlinkedFileID uuid.UUID
	var deletedFileID uuid.UUID

	uc := newTestPet(&stubPetRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.Pet, error) {
			return &model.Pet{ID: petID, Status: "ACTIVE", ProfilePhotoFileID: &oldFileID}, nil
		},
		updatePhotoFn: func(_ context.Context, gotPetID uuid.UUID, rowVersion int, fileID *uuid.UUID) (*model.Pet, error) {
			if gotPetID != petID || rowVersion != 3 {
				t.Fatalf("unexpected update photo args: pet=%s row=%d", gotPetID, rowVersion)
			}
			updatedWithNil = fileID == nil
			return &model.Pet{ID: petID, RowVersion: 4}, nil
		},
	}, &stubACLClient{}, &stubFileClient{
		unlinkPetAvatarFn: func(_ context.Context, fileID, gotPetID uuid.UUID) error {
			if gotPetID != petID {
				t.Fatalf("unexpected pet id: %s", gotPetID)
			}
			unlinkedFileID = fileID
			return nil
		},
		deleteFileIfUnlinkedFn: func(_ context.Context, fileID uuid.UUID) error {
			deletedFileID = fileID
			return nil
		},
	})

	out, err := uc.DeletePetPhoto(context.Background(), DeletePetPhotoParams{UserID: userID, PetID: petID, RowVersion: 3})
	if err != nil {
		t.Fatalf("DeletePetPhoto returned error: %v", err)
	}
	if !updatedWithNil {
		t.Fatal("expected photo to be cleared")
	}
	if unlinkedFileID != oldFileID || deletedFileID != oldFileID {
		t.Fatalf("expected old file cleanup: unlinked=%s deleted=%s", unlinkedFileID, deletedFileID)
	}
	if out.ProfilePhotoFileID != nil {
		t.Fatalf("expected empty photo in result: %+v", out)
	}
}

func stringPtr(v string) *string {
	return &v
}

var (
	_ ports.PetRepository = (*stubPetRepo)(nil)
	_ ports.ACLClient     = (*stubACLClient)(nil)
	_ ports.FileClient    = (*stubFileClient)(nil)
)
