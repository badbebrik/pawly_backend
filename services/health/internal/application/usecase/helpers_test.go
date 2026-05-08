package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"testing"

	"github.com/google/uuid"
)

func TestPrepareHealthAttachmentsValidatesAndDetectsFileType(t *testing.T) {
	imageID := uuid.New()
	pdfID := uuid.New()
	imageName := " image.png "

	out, err := prepareHealthAttachments(context.Background(), &stubHealthFileClient{
		getFilesFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
			return map[uuid.UUID]model.HealthFile{
				imageID: {ID: imageID, MimeType: "image/png"},
				pdfID:   {ID: pdfID, MimeType: "application/pdf", FileName: strPtr("record.pdf")},
			}, nil
		},
	}, []AttachmentParam{
		{FileID: imageID, FileName: &imageName},
		{FileID: pdfID},
	})
	if err != nil {
		t.Fatalf("prepareHealthAttachments returned error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected two attachments, got %+v", out)
	}
	if out[0].FileName == nil || *out[0].FileName != "image.png" || out[0].FileType != "image" {
		t.Fatalf("unexpected image attachment: %+v", out[0])
	}
	if out[1].FileName == nil || *out[1].FileName != "record.pdf" || out[1].FileType != "pdf" {
		t.Fatalf("unexpected document attachment: %+v", out[1])
	}
}

func TestPrepareHealthAttachmentsRejectsDuplicatesAndMissingFiles(t *testing.T) {
	fileID := uuid.New()

	_, err := prepareHealthAttachments(context.Background(), &stubHealthFileClient{}, []AttachmentParam{
		{FileID: fileID},
		{FileID: fileID},
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = prepareHealthAttachments(context.Background(), &stubHealthFileClient{
		getFilesFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
			return map[uuid.UUID]model.HealthFile{}, nil
		},
	}, []AttachmentParam{{FileID: fileID}})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestSyncHealthAttachmentsLinksAndDeletes(t *testing.T) {
	addID := uuid.New()
	removeID := uuid.New()
	var linked []uuid.UUID
	var unlinked []uuid.UUID
	var deleted []uuid.UUID

	err := syncHealthAttachments(context.Background(), &stubHealthFileClient{
		linkAttachmentsFn: func(_ context.Context, _ uuid.UUID, entityType string, _ uuid.UUID, fileIDs []uuid.UUID) error {
			if entityType != "MEDICAL_RECORD" {
				t.Fatalf("unexpected entity type: %s", entityType)
			}
			linked = fileIDs
			return nil
		},
		unlinkAttachmentsFn: func(_ context.Context, entityType string, _ uuid.UUID, fileIDs []uuid.UUID) error {
			if entityType != "MEDICAL_RECORD" {
				t.Fatalf("unexpected entity type: %s", entityType)
			}
			unlinked = fileIDs
			return nil
		},
		deleteFilesIfUnlinkedFn: func(_ context.Context, fileIDs []uuid.UUID) error {
			deleted = fileIDs
			return nil
		},
	}, uuid.New(), "MEDICAL_RECORD", uuid.New(), ports.AttachmentSync{
		Add:    []uuid.UUID{addID},
		Remove: []uuid.UUID{removeID},
	})
	if err != nil {
		t.Fatalf("syncHealthAttachments returned error: %v", err)
	}
	if len(linked) != 1 || linked[0] != addID || len(unlinked) != 1 || unlinked[0] != removeID || len(deleted) != 1 || deleted[0] != removeID {
		t.Fatalf("unexpected sync result: linked=%v unlinked=%v deleted=%v", linked, unlinked, deleted)
	}
}

func TestResolveDictionaryItemUsesExistingOrCreatesCustom(t *testing.T) {
	petID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	customID := uuid.New()
	repo := &stubDictionaryRepo{
		getFn: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.HealthDictionaryItem, error) {
			return &model.HealthDictionaryItem{ID: itemID, Kind: ports.HealthDictionaryKindMedicalRecordType, Name: "Allergy"}, nil
		},
		getOrCreateFn: func(_ context.Context, in ports.GetOrCreateCustomHealthDictionaryItemInput) (*model.HealthDictionaryItem, error) {
			if in.Name != "Checkup" || in.CreatedBy != userID || in.UpdatedBy != userID {
				t.Fatalf("unexpected custom dictionary input: %+v", in)
			}
			return &model.HealthDictionaryItem{ID: customID, Kind: in.Kind, Name: in.Name}, nil
		},
	}

	existing, err := resolveDictionaryItem(context.Background(), repo, petID, userID, ports.HealthDictionaryKindMedicalRecordType, &itemID, nil)
	if err != nil {
		t.Fatalf("resolve existing returned error: %v", err)
	}
	if existing.ID != itemID {
		t.Fatalf("unexpected existing item: %+v", existing)
	}

	custom, err := resolveDictionaryItem(context.Background(), repo, petID, userID, ports.HealthDictionaryKindMedicalRecordType, nil, strPtr(" Checkup "))
	if err != nil {
		t.Fatalf("resolve custom returned error: %v", err)
	}
	if custom.ID != customID {
		t.Fatalf("unexpected custom item: %+v", custom)
	}
}

func TestResolveDictionaryItemRejectsArchivedAndEmpty(t *testing.T) {
	itemID := uuid.New()
	repo := &stubDictionaryRepo{
		getFn: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.HealthDictionaryItem, error) {
			return &model.HealthDictionaryItem{ID: itemID, IsArchived: true}, nil
		},
	}

	_, err := resolveDictionaryItem(context.Background(), repo, uuid.New(), uuid.New(), ports.HealthDictionaryKindMedicalRecordType, &itemID, nil)
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = resolveDictionaryItem(context.Background(), repo, uuid.New(), uuid.New(), ports.HealthDictionaryKindMedicalRecordType, nil, strPtr(" "))
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestResolveDictionaryItemRefsDeduplicates(t *testing.T) {
	itemID := uuid.New()
	out, err := resolveDictionaryItemRefs(context.Background(), &stubDictionaryRepo{
		getFn: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.HealthDictionaryItem, error) {
			return &model.HealthDictionaryItem{ID: itemID, Kind: ports.HealthDictionaryKindVaccinationTarget, Name: "Rabies"}, nil
		},
	}, uuid.New(), uuid.New(), ports.HealthDictionaryKindVaccinationTarget, []HealthDictionaryItemRefParam{
		{ID: &itemID},
		{ID: &itemID},
	})
	if err != nil {
		t.Fatalf("resolveDictionaryItemRefs returned error: %v", err)
	}
	if len(out) != 1 || out[0] != itemID {
		t.Fatalf("unexpected refs: %v", out)
	}
}
