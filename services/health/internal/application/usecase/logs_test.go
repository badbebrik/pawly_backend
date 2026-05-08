package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLogsCreateMetricValidatesInputKinds(t *testing.T) {
	uc := NewLogs(&stubLogsRepo{}, &stubHealthAccess{}, &stubHealthFileClient{})
	userID := uuid.New()
	petID := uuid.New()

	_, err := uc.CreateMetric(context.Background(), CreateMetricParams{
		UserID:    userID,
		PetID:     petID,
		Name:      "Has symptom",
		InputKind: "BOOLEAN",
		Unit:      strPtr("kg"),
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = uc.CreateMetric(context.Background(), CreateMetricParams{
		UserID:    userID,
		PetID:     petID,
		Name:      "Mood",
		InputKind: "SCALE",
	})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestLogsCreateMetricNormalizesAndCreates(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var captured ports.CreateMetricInput
	uc := NewLogs(&stubLogsRepo{
		createMetricFn: func(_ context.Context, input ports.CreateMetricInput) (*model.Metric, error) {
			captured = input
			return &model.Metric{ID: input.ID, Name: input.Name, InputKind: input.InputKind, Unit: input.Unit}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{})

	out, err := uc.CreateMetric(context.Background(), CreateMetricParams{
		UserID:    userID,
		PetID:     petID,
		Name:      " Weight ",
		InputKind: " numeric ",
		Unit:      strPtr(" kg "),
	})
	if err != nil {
		t.Fatalf("CreateMetric returned error: %v", err)
	}
	if out.ID == uuid.Nil || captured.Name != "Weight" || captured.InputKind != "NUMERIC" || captured.Unit == nil || *captured.Unit != "kg" {
		t.Fatalf("unexpected metric input/result: input=%+v out=%+v", captured, out)
	}
}

func TestLogsUpdateMetricRejectsExistingValuesOutOfRange(t *testing.T) {
	uc := NewLogs(&stubLogsRepo{
		hasMetricValuesOutOfRangeFn: func(context.Context, uuid.UUID, uuid.UUID, *float64, *float64) (bool, error) {
			return true, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{})

	_, err := uc.UpdateMetric(context.Background(), UpdateMetricParams{
		UserID:     uuid.New(),
		PetID:      uuid.New(),
		MetricID:   uuid.New(),
		RowVersion: 1,
		Name:       "Mood",
		InputKind:  "SCALE",
		MinValue:   floatPtr(1),
		MaxValue:   floatPtr(5),
	})
	expectHealthErr(t, err, ErrConflict)
}

func TestLogsCreateLogValidatesRequiredMetricAndLinksAttachment(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	logTypeID := uuid.New()
	metricID := uuid.New()
	fileID := uuid.New()
	var linkedEntityID uuid.UUID
	var linkedFiles []uuid.UUID
	var createdInput ports.CreateLogInput

	uc := NewLogs(&stubLogsRepo{
		getMetricsByIDsFn: func(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]model.Metric, error) {
			return map[uuid.UUID]model.Metric{
				metricID: {ID: metricID, InputKind: "SCALE", MinValue: floatPtr(1), MaxValue: floatPtr(5)},
			}, nil
		},
		getLogTypeByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.LogType, error) {
			return &model.LogType{
				ID: logTypeID,
				MetricRequirements: []model.LogTypeMetricRequirement{
					{MetricID: metricID, IsRequired: true},
				},
			}, nil
		},
		createLogFn: func(_ context.Context, input ports.CreateLogInput) (*model.Log, error) {
			createdInput = input
			return &model.Log{
				ID:         input.ID,
				PetID:      input.PetID,
				OccurredAt: input.OccurredAt,
				LogTypeID:  input.LogTypeID,
				Attachments: []model.LogAttachment{
					{FileID: fileID, FileType: "image"},
				},
			}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{
		getFilesFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
			return map[uuid.UUID]model.HealthFile{fileID: {ID: fileID, MimeType: "image/png"}}, nil
		},
		linkAttachmentsFn: func(_ context.Context, gotPetID uuid.UUID, entityType string, entityID uuid.UUID, files []uuid.UUID) error {
			if gotPetID != petID || entityType != "LOG" {
				t.Fatalf("unexpected link args: pet=%s entity=%s", gotPetID, entityType)
			}
			linkedEntityID = entityID
			linkedFiles = files
			return nil
		},
	})

	out, err := uc.CreateLog(context.Background(), CreateLogParams{
		UserID:     userID,
		PetID:      petID,
		OccurredAt: time.Now().UTC(),
		LogTypeID:  &logTypeID,
		MetricValues: []CreateOrUpdateMetricValue{
			{MetricID: metricID, ValueNum: 3},
		},
		Attachments: []AttachmentParam{{FileID: fileID}},
	})
	if err != nil {
		t.Fatalf("CreateLog returned error: %v", err)
	}
	if createdInput.ID == uuid.Nil || len(createdInput.MetricValues) != 1 || len(createdInput.Attachments) != 1 {
		t.Fatalf("unexpected create log input: %+v", createdInput)
	}
	if linkedEntityID != out.ID || len(linkedFiles) != 1 || linkedFiles[0] != fileID {
		t.Fatalf("unexpected linked files: entity=%s files=%v", linkedEntityID, linkedFiles)
	}
	if out.Attachments[0].DownloadURL == nil || out.Attachments[0].PreviewURL == nil {
		t.Fatalf("expected attachment URLs: %+v", out.Attachments[0])
	}
}

func TestLogsCreateLogRejectsInvalidMetricValues(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	metricID := uuid.New()
	uc := NewLogs(&stubLogsRepo{
		getMetricsByIDsFn: func(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]model.Metric, error) {
			return map[uuid.UUID]model.Metric{
				metricID: {ID: metricID, InputKind: "BOOLEAN"},
			}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{})

	_, err := uc.CreateLog(context.Background(), CreateLogParams{
		UserID:     userID,
		PetID:      petID,
		OccurredAt: time.Now().UTC(),
		MetricValues: []CreateOrUpdateMetricValue{
			{MetricID: metricID, ValueNum: 2},
		},
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = uc.CreateLog(context.Background(), CreateLogParams{
		UserID:     userID,
		PetID:      petID,
		OccurredAt: time.Now().UTC(),
		MetricValues: []CreateOrUpdateMetricValue{
			{MetricID: metricID, ValueNum: 1},
			{MetricID: metricID, ValueNum: 1},
		},
	})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestLogsUpdateLogSyncsAttachmentDiff(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	logID := uuid.New()
	oldFileID := uuid.New()
	newFileID := uuid.New()
	var removedFiles []uuid.UUID
	var deletedFiles []uuid.UUID
	var addedFiles []uuid.UUID

	uc := NewLogs(&stubLogsRepo{
		getLogFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.Log, error) {
			return &model.Log{ID: logID, PetID: petID, Attachments: []model.LogAttachment{{FileID: oldFileID}}}, nil
		},
		getMetricsByIDsFn: func(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]model.Metric, error) {
			return map[uuid.UUID]model.Metric{}, nil
		},
		updateLogFn: func(_ context.Context, input ports.UpdateLogInput) (*model.Log, error) {
			return &model.Log{ID: input.ID, PetID: input.PetID, Attachments: []model.LogAttachment{{FileID: newFileID}}}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{
		getFilesFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
			return map[uuid.UUID]model.HealthFile{newFileID: {ID: newFileID, MimeType: "application/pdf"}}, nil
		},
		linkAttachmentsFn: func(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, files []uuid.UUID) error {
			addedFiles = files
			return nil
		},
		unlinkAttachmentsFn: func(_ context.Context, _ string, _ uuid.UUID, files []uuid.UUID) error {
			removedFiles = files
			return nil
		},
		deleteFilesIfUnlinkedFn: func(_ context.Context, files []uuid.UUID) error {
			deletedFiles = files
			return nil
		},
	})

	_, err := uc.UpdateLog(context.Background(), UpdateLogParams{
		UserID:      userID,
		PetID:       petID,
		LogID:       logID,
		RowVersion:  1,
		OccurredAt:  time.Now().UTC(),
		Attachments: []AttachmentParam{{FileID: newFileID}},
	})
	if err != nil {
		t.Fatalf("UpdateLog returned error: %v", err)
	}
	if len(addedFiles) != 1 || addedFiles[0] != newFileID || len(removedFiles) != 1 || removedFiles[0] != oldFileID || len(deletedFiles) != 1 || deletedFiles[0] != oldFileID {
		t.Fatalf("unexpected attachment diff: add=%v remove=%v delete=%v", addedFiles, removedFiles, deletedFiles)
	}
}

func TestLogsBootstrapIncludesCatalogWhenRequested(t *testing.T) {
	uc := NewLogs(&stubLogsRepo{
		listRecentLogTypesFn: func(context.Context, uuid.UUID, int) ([]model.LogType, error) {
			return []model.LogType{{ID: uuid.New(), Name: "Walk"}}, nil
		},
		listLogTypesFn: func(_ context.Context, input ports.ListLogTypesInput) ([]model.LogType, error) {
			return []model.LogType{{ID: uuid.New(), Scope: input.Scope}}, nil
		},
		listMetricsFn: func(_ context.Context, input ports.ListMetricsInput) ([]model.Metric, error) {
			return []model.Metric{{ID: uuid.New(), Scope: input.Scope, Name: "Weight"}}, nil
		},
	}, &stubHealthAccess{}, &stubHealthFileClient{})

	out, err := uc.GetLogsBootstrap(context.Background(), GetLogsBootstrapParams{
		UserID:         uuid.New(),
		PetID:          uuid.New(),
		IncludeCatalog: true,
	})
	if err != nil {
		t.Fatalf("GetLogsBootstrap returned error: %v", err)
	}
	if !out.Permissions.LogRead || len(out.RecentLogTypes) != 1 || len(out.SystemLogTypes) != 1 || len(out.CustomLogTypes) != 1 || len(out.SystemMetrics) != 1 || len(out.CustomMetrics) != 1 {
		t.Fatalf("unexpected bootstrap: %+v", out)
	}
}
