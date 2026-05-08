package usecase

import (
	"context"
	"errors"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubMedicalRecordRepo struct {
	getFn    func(context.Context, uuid.UUID, uuid.UUID) (*model.MedicalRecord, error)
	listFn   func(context.Context, ports.ListMedicalRecordsQuery) (ports.ListMedicalRecordsResult, error)
	createFn func(context.Context, ports.CreateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error)
	updateFn func(context.Context, ports.UpdateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error)
	deleteFn func(context.Context, ports.DeleteMedicalRecordInput) error
}

func (s *stubMedicalRecordRepo) GetMedicalRecord(ctx context.Context, petID, recordID uuid.UUID) (*model.MedicalRecord, error) {
	if s.getFn != nil {
		return s.getFn(ctx, petID, recordID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubMedicalRecordRepo) ListMedicalRecords(ctx context.Context, query ports.ListMedicalRecordsQuery) (ports.ListMedicalRecordsResult, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return ports.ListMedicalRecordsResult{}, nil
}

func (s *stubMedicalRecordRepo) CreateMedicalRecord(ctx context.Context, input ports.CreateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error) {
	if s.createFn != nil {
		return s.createFn(ctx, input)
	}
	return &model.MedicalRecord{ID: input.ID, PetID: input.PetID, Title: input.Title, Status: input.Status}, ports.AttachmentSync{}, nil
}

func (s *stubMedicalRecordRepo) UpdateMedicalRecord(ctx context.Context, input ports.UpdateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, input)
	}
	return &model.MedicalRecord{ID: input.ID, PetID: input.PetID, Title: input.Title, Status: input.Status, RowVersion: input.RowVersion + 1}, ports.AttachmentSync{}, nil
}

func (s *stubMedicalRecordRepo) DeleteMedicalRecord(ctx context.Context, input ports.DeleteMedicalRecordInput) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, input)
	}
	return nil
}

func TestMedicalRecordsListNormalizesQuery(t *testing.T) {
	var captured ports.ListMedicalRecordsQuery
	uc := NewMedicalRecords(&stubMedicalRecordRepo{
		listFn: func(_ context.Context, query ports.ListMedicalRecordsQuery) (ports.ListMedicalRecordsResult, error) {
			captured = query
			return ports.ListMedicalRecordsResult{Items: []model.MedicalRecordListItem{{ID: uuid.New(), Title: "Allergy"}}}, nil
		},
	}, &stubDictionaryRepo{}, &stubHealthAccess{}, &stubHealthFileClient{})

	out, err := uc.ListMedicalRecords(context.Background(), ListMedicalRecordsParams{
		UserID: uuid.New(),
		PetID:  uuid.New(),
		Q:      " allergy ",
		Status: "active",
		Bucket: "archive",
	})
	if err != nil {
		t.Fatalf("ListMedicalRecords returned error: %v", err)
	}
	if len(out.Items) != 1 || captured.Q != "allergy" || captured.Status == nil || *captured.Status != "ACTIVE" || captured.Bucket != "archive" {
		t.Fatalf("unexpected list query/result: query=%+v out=%+v", captured, out)
	}
}

func TestMedicalRecordsCreateResolvesDictionaryAndSyncsAttachments(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	recordTypeID := uuid.New()
	fileID := uuid.New()
	var created ports.CreateMedicalRecordInput
	var linked []uuid.UUID
	uc := NewMedicalRecords(&stubMedicalRecordRepo{
		createFn: func(_ context.Context, input ports.CreateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error) {
			created = input
			return &model.MedicalRecord{
				ID:     input.ID,
				PetID:  input.PetID,
				Title:  input.Title,
				Status: input.Status,
				Attachments: []model.HealthAttachment{
					{FileID: fileID, FileType: "image"},
				},
			}, ports.AttachmentSync{Add: []uuid.UUID{fileID}}, nil
		},
	}, &stubDictionaryRepo{
		getFn: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.HealthDictionaryItem, error) {
			return &model.HealthDictionaryItem{ID: recordTypeID, Kind: ports.HealthDictionaryKindMedicalRecordType, Name: "Allergy"}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{
		getFilesFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
			return map[uuid.UUID]model.HealthFile{fileID: {ID: fileID, MimeType: "image/png"}}, nil
		},
		linkAttachmentsFn: func(_ context.Context, gotPetID uuid.UUID, entityType string, entityID uuid.UUID, files []uuid.UUID) error {
			if gotPetID != petID || entityType != "MEDICAL_RECORD" || entityID == uuid.Nil {
				t.Fatalf("unexpected link args: pet=%s entity=%s id=%s", gotPetID, entityType, entityID)
			}
			linked = files
			return nil
		},
	})

	out, err := uc.CreateMedicalRecord(context.Background(), CreateMedicalRecordParams{
		UserID:       userID,
		PetID:        petID,
		RecordTypeID: &recordTypeID,
		Status:       " active ",
		Title:        " Allergy ",
		Attachments:  []AttachmentParam{{FileID: fileID}},
	})
	if err != nil {
		t.Fatalf("CreateMedicalRecord returned error: %v", err)
	}
	if out.ID == uuid.Nil || created.Title != "Allergy" || created.Status != "ACTIVE" || created.RecordTypeItemID == nil || *created.RecordTypeItemID != recordTypeID {
		t.Fatalf("unexpected create input/result: input=%+v out=%+v", created, out)
	}
	if len(linked) != 1 || linked[0] != fileID {
		t.Fatalf("unexpected linked files: %v", linked)
	}
	if out.Attachments[0].DownloadURL == nil || out.Attachments[0].PreviewURL == nil {
		t.Fatalf("expected enriched attachment: %+v", out.Attachments[0])
	}
}

func TestMedicalRecordsRejectInvalidStatusAndTitle(t *testing.T) {
	uc := NewMedicalRecords(&stubMedicalRecordRepo{}, &stubDictionaryRepo{}, &stubHealthAccess{}, &stubHealthFileClient{})

	_, err := uc.CreateMedicalRecord(context.Background(), CreateMedicalRecordParams{
		UserID:         uuid.New(),
		PetID:          uuid.New(),
		RecordTypeName: strPtr("Allergy"),
		Status:         "BAD",
		Title:          "Allergy",
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = uc.CreateMedicalRecord(context.Background(), CreateMedicalRecordParams{
		UserID:         uuid.New(),
		PetID:          uuid.New(),
		RecordTypeName: strPtr("Allergy"),
		Status:         "ACTIVE",
		Title:          " ",
	})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestMedicalRecordsDeleteRemovesAttachments(t *testing.T) {
	petID := uuid.New()
	recordID := uuid.New()
	fileID := uuid.New()
	var unlinked []uuid.UUID
	var deleted []uuid.UUID
	uc := NewMedicalRecords(&stubMedicalRecordRepo{
		getFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: recordID, PetID: petID, Attachments: []model.HealthAttachment{{FileID: fileID}}}, nil
		},
	}, &stubDictionaryRepo{}, &stubHealthAccess{}, &stubHealthFileClient{
		unlinkAttachmentsFn: func(_ context.Context, entityType string, entityID uuid.UUID, files []uuid.UUID) error {
			if entityType != "MEDICAL_RECORD" || entityID != recordID {
				t.Fatalf("unexpected unlink args: entity=%s id=%s", entityType, entityID)
			}
			unlinked = files
			return nil
		},
		deleteFilesIfUnlinkedFn: func(context.Context, []uuid.UUID) error {
			deleted = []uuid.UUID{fileID}
			return nil
		},
	})

	err := uc.DeleteMedicalRecord(context.Background(), DeleteMedicalRecordParams{
		UserID: uuid.New(), PetID: petID, RecordID: recordID, RowVersion: 1,
	})
	if err != nil {
		t.Fatalf("DeleteMedicalRecord returned error: %v", err)
	}
	if len(unlinked) != 1 || unlinked[0] != fileID || len(deleted) != 1 || deleted[0] != fileID {
		t.Fatalf("unexpected attachment cleanup: unlinked=%v deleted=%v", unlinked, deleted)
	}
}

type stubDispatchRepo struct {
	items     []model.ScheduledItemOccurrenceListItem
	createErr error
	created   []ports.CreateScheduledItemDispatchParams
}

func (s *stubDispatchRepo) ListDueScheduledItemOccurrences(context.Context, ports.ListDueScheduledItemOccurrencesParams) ([]model.ScheduledItemOccurrenceListItem, error) {
	return s.items, nil
}

func (s *stubDispatchRepo) CreateScheduledItemDispatch(_ context.Context, params ports.CreateScheduledItemDispatchParams) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.created = append(s.created, params)
	return nil
}

type stubPetUserLister struct {
	users []uuid.UUID
	err   error
}

func (s *stubPetUserLister) ListPetUserIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.users, nil
}

type stubPushPublisher struct {
	err       error
	published []model.ScheduledOccurrencePushJob
}

func (s *stubPushPublisher) PublishScheduledOccurrenceDue(_ context.Context, job model.ScheduledOccurrencePushJob) error {
	if s.err != nil {
		return s.err
	}
	s.published = append(s.published, job)
	return nil
}

func TestScheduledDispatcherPublishesDueOccurrences(t *testing.T) {
	petID := uuid.New()
	occurrenceID := uuid.New()
	userID := uuid.New()
	repo := &stubDispatchRepo{items: []model.ScheduledItemOccurrenceListItem{{
		ID:              occurrenceID,
		PetID:           petID,
		ScheduledItemID: uuid.New(),
		ScheduledFor:    time.Now().UTC(),
		Rule:            model.ScheduledItem{SourceType: model.ScheduledItemSourceTypeManual, Title: "Walk", Note: strPtr("Take a walk")},
	}}}
	publisher := &stubPushPublisher{}
	uc := NewScheduledDispatcher(ScheduledDispatcherDependencies{
		Repository:    repo,
		PetUserLister: &stubPetUserLister{users: []uuid.UUID{userID}},
		PushPublisher: publisher,
	})

	out, err := uc.DispatchDueScheduledOccurrences(context.Background(), 10)
	if err != nil {
		t.Fatalf("DispatchDueScheduledOccurrences returned error: %v", err)
	}
	if out.Scanned != 1 || out.Published != 1 || len(repo.created) != 1 || len(publisher.published) != 1 {
		t.Fatalf("unexpected dispatch result: %+v created=%v published=%v", out, repo.created, publisher.published)
	}
	if publisher.published[0].PetID != petID.String() || publisher.published[0].UserIDs[0] != userID.String() {
		t.Fatalf("unexpected push job: %+v", publisher.published[0])
	}
}

func TestScheduledDispatcherSkipsAndRecordsFailures(t *testing.T) {
	petID := uuid.New()
	item := model.ScheduledItemOccurrenceListItem{ID: uuid.New(), PetID: petID, ScheduledItemID: uuid.New(), ScheduledFor: time.Now().UTC()}

	t.Run("no users", func(t *testing.T) {
		uc := NewScheduledDispatcher(ScheduledDispatcherDependencies{
			Repository:    &stubDispatchRepo{items: []model.ScheduledItemOccurrenceListItem{item}},
			PetUserLister: &stubPetUserLister{},
			PushPublisher: &stubPushPublisher{},
		})
		out, err := uc.DispatchDueScheduledOccurrences(context.Background(), 10)
		if err != nil {
			t.Fatalf("dispatch returned error: %v", err)
		}
		if out.Skipped != 1 || out.Published != 0 {
			t.Fatalf("unexpected result: %+v", out)
		}
	})

	t.Run("dispatch conflict", func(t *testing.T) {
		uc := NewScheduledDispatcher(ScheduledDispatcherDependencies{
			Repository:    &stubDispatchRepo{items: []model.ScheduledItemOccurrenceListItem{item}, createErr: ports.ErrConflict},
			PetUserLister: &stubPetUserLister{users: []uuid.UUID{uuid.New()}},
			PushPublisher: &stubPushPublisher{},
		})
		out, err := uc.DispatchDueScheduledOccurrences(context.Background(), 10)
		if err != nil {
			t.Fatalf("dispatch returned error: %v", err)
		}
		if out.Skipped != 1 || out.Failed != 0 {
			t.Fatalf("unexpected result: %+v", out)
		}
	})

	t.Run("publish failure", func(t *testing.T) {
		uc := NewScheduledDispatcher(ScheduledDispatcherDependencies{
			Repository:    &stubDispatchRepo{items: []model.ScheduledItemOccurrenceListItem{item}},
			PetUserLister: &stubPetUserLister{users: []uuid.UUID{uuid.New()}},
			PushPublisher: &stubPushPublisher{err: errors.New("publish failed")},
		})
		out, err := uc.DispatchDueScheduledOccurrences(context.Background(), 10)
		if err != nil {
			t.Fatalf("dispatch returned error: %v", err)
		}
		if out.Failed != 1 || len(out.Failures) != 1 || out.Failures[0].Operation != "publish_push_job" {
			t.Fatalf("unexpected result: %+v", out)
		}
	})
}

var (
	_ ports.MedicalRecordRepository     = (*stubMedicalRecordRepo)(nil)
	_ ports.ScheduledDispatchRepository = (*stubDispatchRepo)(nil)
	_ ports.PetUserLister               = (*stubPetUserLister)(nil)
	_ ports.PushPublisher               = (*stubPushPublisher)(nil)
)
