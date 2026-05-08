package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubDocumentsRepo struct {
	listFn   func(context.Context, ports.ListPetDocumentsQuery) (ports.ListPetDocumentsResult, error)
	getFn    func(context.Context, uuid.UUID, uuid.UUID) (*model.PetDocument, error)
	renameFn func(context.Context, ports.RenamePetDocumentInput) (*model.PetDocument, error)
}

func (s *stubDocumentsRepo) ListPetDocuments(ctx context.Context, query ports.ListPetDocumentsQuery) (ports.ListPetDocumentsResult, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return ports.ListPetDocumentsResult{}, nil
}

func (s *stubDocumentsRepo) GetPetDocument(ctx context.Context, petID, documentID uuid.UUID) (*model.PetDocument, error) {
	if s.getFn != nil {
		return s.getFn(ctx, petID, documentID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubDocumentsRepo) RenamePetDocument(ctx context.Context, input ports.RenamePetDocumentInput) (*model.PetDocument, error) {
	if s.renameFn != nil {
		return s.renameFn(ctx, input)
	}
	name := input.FileName
	return &model.PetDocument{ID: input.ID, PetID: input.PetID, FileName: &name, FileID: uuid.New(), FileType: "document"}, nil
}

func TestDocumentsListExcludesLogWhenLogReadDenied(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var captured ports.ListPetDocumentsQuery
	fileID := uuid.New()
	uc := NewDocuments(&stubDocumentsRepo{
		listFn: func(_ context.Context, query ports.ListPetDocumentsQuery) (ports.ListPetDocumentsResult, error) {
			captured = query
			return ports.ListPetDocumentsResult{
				Items: []model.PetDocument{{ID: uuid.New(), PetID: query.PetID, FileID: fileID, FileType: "image"}},
			}, nil
		},
	}, &stubHealthAccess{allowed: map[string]bool{
		ActionHealthRead: true,
		ActionLogRead:    false,
	}}, &stubHealthFileClient{
		batchGetDownloadURLsFn: func(context.Context, []uuid.UUID) (map[uuid.UUID]string, error) {
			return map[uuid.UUID]string{fileID: "download-url"}, nil
		},
	})

	out, err := uc.ListPetDocuments(context.Background(), ListPetDocumentsParams{
		UserID:   userID,
		PetID:    petID,
		Q:        " record ",
		FileType: strPtr(" image "),
	})
	if err != nil {
		t.Fatalf("ListPetDocuments returned error: %v", err)
	}
	if !captured.ExcludeLog || captured.Q != "record" || captured.FileType == nil || *captured.FileType != "image" {
		t.Fatalf("unexpected list query: %+v", captured)
	}
	if len(out.Items) != 1 || out.Items[0].DownloadURL == nil || out.Items[0].PreviewURL == nil {
		t.Fatalf("expected enriched image document: %+v", out.Items)
	}
}

func TestDocumentsListLogFilterReturnsEmptyWithoutLogRead(t *testing.T) {
	uc := NewDocuments(&stubDocumentsRepo{
		listFn: func(context.Context, ports.ListPetDocumentsQuery) (ports.ListPetDocumentsResult, error) {
			t.Fatal("repo should not be called for LOG filter without log_read")
			return ports.ListPetDocumentsResult{}, nil
		},
	}, &stubHealthAccess{allowed: map[string]bool{
		ActionHealthRead: true,
		ActionLogRead:    false,
	}}, &stubHealthFileClient{})

	out, err := uc.ListPetDocuments(context.Background(), ListPetDocumentsParams{
		UserID:     uuid.New(),
		PetID:      uuid.New(),
		EntityType: strPtr("log"),
	})
	if err != nil {
		t.Fatalf("ListPetDocuments returned error: %v", err)
	}
	if len(out.Items) != 0 {
		t.Fatalf("expected empty list, got %+v", out.Items)
	}
}

func TestDocumentsRenameChecksEntitySpecificPermission(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	docID := uuid.New()

	t.Run("log requires log write", func(t *testing.T) {
		uc := NewDocuments(&stubDocumentsRepo{
			getFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.PetDocument, error) {
				return &model.PetDocument{ID: docID, PetID: petID, EntityType: "LOG", FileID: uuid.New(), FileType: "document"}, nil
			},
			renameFn: func(_ context.Context, input ports.RenamePetDocumentInput) (*model.PetDocument, error) {
				return &model.PetDocument{ID: input.ID, PetID: input.PetID, EntityType: "LOG", FileID: uuid.New(), FileName: &input.FileName}, nil
			},
		}, &stubHealthAccess{allowed: map[string]bool{ActionLogWrite: true}}, &stubHealthFileClient{})

		out, err := uc.RenamePetDocument(context.Background(), RenamePetDocumentParams{
			UserID: userID, PetID: petID, DocumentID: docID, FileName: " record.pdf ",
		})
		if err != nil {
			t.Fatalf("RenamePetDocument returned error: %v", err)
		}
		if out.FileName == nil || *out.FileName != "record.pdf" {
			t.Fatalf("unexpected rename result: %+v", out)
		}
	})

	t.Run("health entity requires health write", func(t *testing.T) {
		uc := NewDocuments(&stubDocumentsRepo{
			getFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.PetDocument, error) {
				return &model.PetDocument{ID: docID, PetID: petID, EntityType: "MEDICAL_RECORD", FileID: uuid.New(), FileType: "document"}, nil
			},
		}, &stubHealthAccess{allowed: map[string]bool{ActionHealthWrite: false}}, &stubHealthFileClient{})

		_, err := uc.RenamePetDocument(context.Background(), RenamePetDocumentParams{
			UserID: userID, PetID: petID, DocumentID: docID, FileName: "record.pdf",
		})
		expectHealthErr(t, err, ErrForbidden)
	})
}

func TestAnalyticsListNormalizesLimitAndTypeIDs(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	typeID := uuid.New()
	var captured ports.ListAnalyticsMetricsInput
	uc := NewAnalytics(&stubLogsRepo{
		listAnalyticsMetricsFn: func(_ context.Context, input ports.ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error) {
			captured = input
			return []model.AnalyticsMetricSummary{{MetricID: uuid.New(), MetricName: "Weight"}}, nil
		},
	}, &stubHealthAccess{})

	out, err := uc.ListAnalyticsMetrics(context.Background(), ListAnalyticsMetricsParams{
		UserID:  userID,
		PetID:   petID,
		Q:       " weight ",
		TypeIDs: []uuid.UUID{typeID, typeID},
		Limit:   999,
	})
	if err != nil {
		t.Fatalf("ListAnalyticsMetrics returned error: %v", err)
	}
	if len(out) != 1 || captured.Q != "weight" || captured.Limit != 500 || len(captured.TypeIDs) != 1 || captured.TypeIDs[0] != typeID {
		t.Fatalf("unexpected analytics query/result: query=%+v out=%+v", captured, out)
	}
}

func TestAnalyticsRejectsInvalidDateRangeAndSort(t *testing.T) {
	now := time.Now().UTC()
	before := now.Add(-time.Hour)
	uc := NewAnalytics(&stubLogsRepo{}, &stubHealthAccess{})

	_, err := uc.ListAnalyticsMetrics(context.Background(), ListAnalyticsMetricsParams{
		UserID:   uuid.New(),
		PetID:    uuid.New(),
		DateFrom: &now,
		DateTo:   &before,
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = uc.GetMetricSeries(context.Background(), GetMetricSeriesParams{
		UserID:   uuid.New(),
		PetID:    uuid.New(),
		MetricID: uuid.New(),
		Sort:     "bad",
	})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestAnalyticsGetMetricSeriesReturnsMetricPointsAndSummary(t *testing.T) {
	metricID := uuid.New()
	summary := &model.MetricSeriesSummary{PointsCount: 1, LastValueNum: 10}
	uc := NewAnalytics(&stubLogsRepo{
		getMetricByIDFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.Metric, error) {
			return &model.Metric{ID: metricID, Name: "Weight"}, nil
		},
		listMetricSeriesFn: func(_ context.Context, input ports.ListMetricSeriesInput) ([]model.MetricSeriesPoint, *model.MetricSeriesSummary, error) {
			if input.Sort != "occurred_at_desc" || !input.IncludeSummary {
				t.Fatalf("unexpected series input: %+v", input)
			}
			return []model.MetricSeriesPoint{{LogID: uuid.New(), ValueNum: 10}}, summary, nil
		},
	}, &stubHealthAccess{})

	out, err := uc.GetMetricSeries(context.Background(), GetMetricSeriesParams{
		UserID:         uuid.New(),
		PetID:          uuid.New(),
		MetricID:       metricID,
		Sort:           "occurred_at_desc",
		IncludeSummary: true,
	})
	if err != nil {
		t.Fatalf("GetMetricSeries returned error: %v", err)
	}
	if out.Metric.ID != metricID || len(out.Points) != 1 || out.Summary != summary {
		t.Fatalf("unexpected series result: %+v", out)
	}
}

var _ ports.DocumentsRepository = (*stubDocumentsRepo)(nil)
