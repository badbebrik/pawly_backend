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

type stubHealthAccess struct {
	allowed map[string]bool
	err     error
}

func (s *stubHealthAccess) Check(_ context.Context, _, _ uuid.UUID, action string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	if s.allowed == nil {
		return true, nil
	}
	return s.allowed[action], nil
}

type stubHealthFileClient struct {
	initUploadFn            func(context.Context, string, int64, string) (uuid.UUID, ports.UploadInfo, error)
	confirmUploadFn         func(context.Context, uuid.UUID, int64) (*ports.UploadedFile, error)
	batchGetDownloadURLsFn  func(context.Context, []uuid.UUID) (map[uuid.UUID]string, error)
	getFilesFn              func(context.Context, []uuid.UUID) (map[uuid.UUID]model.HealthFile, error)
	linkAttachmentsFn       func(context.Context, uuid.UUID, string, uuid.UUID, []uuid.UUID) error
	unlinkAttachmentsFn     func(context.Context, string, uuid.UUID, []uuid.UUID) error
	deleteFilesIfUnlinkedFn func(context.Context, []uuid.UUID) error
}

func (s *stubHealthFileClient) InitUpload(ctx context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, ports.UploadInfo, error) {
	if s.initUploadFn != nil {
		return s.initUploadFn(ctx, mimeType, expectedSize, originalFilename)
	}
	return uuid.New(), ports.UploadInfo{Method: "PUT", URL: "upload-url"}, nil
}

func (s *stubHealthFileClient) ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) (*ports.UploadedFile, error) {
	if s.confirmUploadFn != nil {
		return s.confirmUploadFn(ctx, fileID, sizeBytes)
	}
	return &ports.UploadedFile{ID: fileID, MimeType: "image/png", SizeBytes: sizeBytes}, nil
}

func (s *stubHealthFileClient) BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if s.batchGetDownloadURLsFn != nil {
		return s.batchGetDownloadURLsFn(ctx, fileIDs)
	}
	out := make(map[uuid.UUID]string, len(fileIDs))
	for _, id := range fileIDs {
		out[id] = "download-url"
	}
	return out, nil
}

func (s *stubHealthFileClient) GetFiles(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
	if s.getFilesFn != nil {
		return s.getFilesFn(ctx, fileIDs)
	}
	out := make(map[uuid.UUID]model.HealthFile, len(fileIDs))
	for _, id := range fileIDs {
		out[id] = model.HealthFile{ID: id, MimeType: "image/png"}
	}
	return out, nil
}

func (s *stubHealthFileClient) LinkAttachments(ctx context.Context, petID uuid.UUID, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error {
	if s.linkAttachmentsFn != nil {
		return s.linkAttachmentsFn(ctx, petID, entityType, entityID, fileIDs)
	}
	return nil
}

func (s *stubHealthFileClient) UnlinkAttachments(ctx context.Context, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error {
	if s.unlinkAttachmentsFn != nil {
		return s.unlinkAttachmentsFn(ctx, entityType, entityID, fileIDs)
	}
	return nil
}

func (s *stubHealthFileClient) DeleteFilesIfUnlinked(ctx context.Context, fileIDs []uuid.UUID) error {
	if s.deleteFilesIfUnlinkedFn != nil {
		return s.deleteFilesIfUnlinkedFn(ctx, fileIDs)
	}
	return nil
}

type stubLogsRepo struct {
	getLogFn                    func(context.Context, uuid.UUID, uuid.UUID) (*model.Log, error)
	listLogsFn                  func(context.Context, ports.ListLogsQuery) (ports.ListLogsResult, error)
	createLogFn                 func(context.Context, ports.CreateLogInput) (*model.Log, error)
	updateLogFn                 func(context.Context, ports.UpdateLogInput) (*model.Log, error)
	softDeleteLogFn             func(context.Context, ports.DeleteLogInput) error
	getLogTypeByIDFn            func(context.Context, uuid.UUID, uuid.UUID) (*model.LogType, error)
	getMetricsByIDsFn           func(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID]model.Metric, error)
	listRecentLogTypesFn        func(context.Context, uuid.UUID, int) ([]model.LogType, error)
	listLogTypesFn              func(context.Context, ports.ListLogTypesInput) ([]model.LogType, error)
	createLogTypeFn             func(context.Context, ports.CreateLogTypeInput) (*model.LogType, error)
	updateLogTypeFn             func(context.Context, ports.UpdateLogTypeInput) (*model.LogType, error)
	archiveLogTypeFn            func(context.Context, ports.ArchiveLogTypeInput) error
	listMetricsFn               func(context.Context, ports.ListMetricsInput) ([]model.Metric, error)
	createMetricFn              func(context.Context, ports.CreateMetricInput) (*model.Metric, error)
	updateMetricFn              func(context.Context, ports.UpdateMetricInput) (*model.Metric, error)
	archiveMetricFn             func(context.Context, ports.ArchiveMetricInput) error
	hasMetricValuesOutOfRangeFn func(context.Context, uuid.UUID, uuid.UUID, *float64, *float64) (bool, error)
	getMetricByIDFn             func(context.Context, uuid.UUID, uuid.UUID) (*model.Metric, error)
	listAnalyticsMetricsFn      func(context.Context, ports.ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error)
	listMetricSeriesFn          func(context.Context, ports.ListMetricSeriesInput) ([]model.MetricSeriesPoint, *model.MetricSeriesSummary, error)
}

func (s *stubLogsRepo) GetLog(ctx context.Context, petID, logID uuid.UUID) (*model.Log, error) {
	if s.getLogFn != nil {
		return s.getLogFn(ctx, petID, logID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubLogsRepo) ListLogs(ctx context.Context, query ports.ListLogsQuery) (ports.ListLogsResult, error) {
	if s.listLogsFn != nil {
		return s.listLogsFn(ctx, query)
	}
	return ports.ListLogsResult{}, nil
}

func (s *stubLogsRepo) CreateLog(ctx context.Context, input ports.CreateLogInput) (*model.Log, error) {
	if s.createLogFn != nil {
		return s.createLogFn(ctx, input)
	}
	return &model.Log{ID: input.ID, PetID: input.PetID, OccurredAt: input.OccurredAt, LogTypeID: input.LogTypeID}, nil
}

func (s *stubLogsRepo) UpdateLog(ctx context.Context, input ports.UpdateLogInput) (*model.Log, error) {
	if s.updateLogFn != nil {
		return s.updateLogFn(ctx, input)
	}
	return &model.Log{ID: input.ID, PetID: input.PetID, OccurredAt: input.OccurredAt, LogTypeID: input.LogTypeID}, nil
}

func (s *stubLogsRepo) SoftDeleteLog(ctx context.Context, input ports.DeleteLogInput) error {
	if s.softDeleteLogFn != nil {
		return s.softDeleteLogFn(ctx, input)
	}
	return nil
}

func (s *stubLogsRepo) GetLogTypeByID(ctx context.Context, petID, logTypeID uuid.UUID) (*model.LogType, error) {
	if s.getLogTypeByIDFn != nil {
		return s.getLogTypeByIDFn(ctx, petID, logTypeID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubLogsRepo) GetMetricsByIDs(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID]model.Metric, error) {
	if s.getMetricsByIDsFn != nil {
		return s.getMetricsByIDsFn(ctx, petID, metricIDs)
	}
	return map[uuid.UUID]model.Metric{}, nil
}

func (s *stubLogsRepo) ListRecentLogTypes(ctx context.Context, petID uuid.UUID, limit int) ([]model.LogType, error) {
	if s.listRecentLogTypesFn != nil {
		return s.listRecentLogTypesFn(ctx, petID, limit)
	}
	return []model.LogType{}, nil
}

func (s *stubLogsRepo) ListLogTypes(ctx context.Context, input ports.ListLogTypesInput) ([]model.LogType, error) {
	if s.listLogTypesFn != nil {
		return s.listLogTypesFn(ctx, input)
	}
	return []model.LogType{}, nil
}

func (s *stubLogsRepo) CreateLogType(ctx context.Context, input ports.CreateLogTypeInput) (*model.LogType, error) {
	if s.createLogTypeFn != nil {
		return s.createLogTypeFn(ctx, input)
	}
	return &model.LogType{ID: input.ID, Name: input.Name, MetricRequirements: []model.LogTypeMetricRequirement{}}, nil
}

func (s *stubLogsRepo) UpdateLogType(ctx context.Context, input ports.UpdateLogTypeInput) (*model.LogType, error) {
	if s.updateLogTypeFn != nil {
		return s.updateLogTypeFn(ctx, input)
	}
	return &model.LogType{ID: input.ID, Name: input.Name, RowVersion: input.RowVersion + 1}, nil
}

func (s *stubLogsRepo) ArchiveLogType(ctx context.Context, input ports.ArchiveLogTypeInput) error {
	if s.archiveLogTypeFn != nil {
		return s.archiveLogTypeFn(ctx, input)
	}
	return nil
}

func (s *stubLogsRepo) ListMetrics(ctx context.Context, input ports.ListMetricsInput) ([]model.Metric, error) {
	if s.listMetricsFn != nil {
		return s.listMetricsFn(ctx, input)
	}
	return []model.Metric{}, nil
}

func (s *stubLogsRepo) CreateMetric(ctx context.Context, input ports.CreateMetricInput) (*model.Metric, error) {
	if s.createMetricFn != nil {
		return s.createMetricFn(ctx, input)
	}
	return &model.Metric{ID: input.ID, Name: input.Name, InputKind: input.InputKind, Unit: input.Unit, MinValue: input.MinValue, MaxValue: input.MaxValue}, nil
}

func (s *stubLogsRepo) UpdateMetric(ctx context.Context, input ports.UpdateMetricInput) (*model.Metric, error) {
	if s.updateMetricFn != nil {
		return s.updateMetricFn(ctx, input)
	}
	return &model.Metric{ID: input.ID, Name: input.Name, InputKind: input.InputKind, Unit: input.Unit, MinValue: input.MinValue, MaxValue: input.MaxValue}, nil
}

func (s *stubLogsRepo) ArchiveMetric(ctx context.Context, input ports.ArchiveMetricInput) error {
	if s.archiveMetricFn != nil {
		return s.archiveMetricFn(ctx, input)
	}
	return nil
}

func (s *stubLogsRepo) HasMetricValuesOutOfRange(ctx context.Context, petID, metricID uuid.UUID, minValue, maxValue *float64) (bool, error) {
	if s.hasMetricValuesOutOfRangeFn != nil {
		return s.hasMetricValuesOutOfRangeFn(ctx, petID, metricID, minValue, maxValue)
	}
	return false, nil
}

func (s *stubLogsRepo) GetMetricByID(ctx context.Context, petID, metricID uuid.UUID) (*model.Metric, error) {
	if s.getMetricByIDFn != nil {
		return s.getMetricByIDFn(ctx, petID, metricID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubLogsRepo) ListAnalyticsMetrics(ctx context.Context, input ports.ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error) {
	if s.listAnalyticsMetricsFn != nil {
		return s.listAnalyticsMetricsFn(ctx, input)
	}
	return []model.AnalyticsMetricSummary{}, nil
}

func (s *stubLogsRepo) ListMetricSeries(ctx context.Context, input ports.ListMetricSeriesInput) ([]model.MetricSeriesPoint, *model.MetricSeriesSummary, error) {
	if s.listMetricSeriesFn != nil {
		return s.listMetricSeriesFn(ctx, input)
	}
	return []model.MetricSeriesPoint{}, nil, nil
}

type stubScheduledRepo struct {
	getScheduledItemFn                           func(context.Context, uuid.UUID, uuid.UUID) (*model.ScheduledItem, error)
	getScheduledItemBySourceFn                   func(context.Context, uuid.UUID, string, uuid.UUID) (*model.ScheduledItem, error)
	listScheduledItemsFn                         func(context.Context, ports.ListScheduledItemsQuery) (ports.ListScheduledItemsResult, error)
	listRecurringScheduledItemsForHorizonFn      func(context.Context, ports.ListRecurringScheduledItemsForHorizonParams) ([]model.ScheduledItem, error)
	createScheduledItemFn                        func(context.Context, ports.CreateScheduledItemInput) (*model.ScheduledItem, error)
	updateScheduledItemFn                        func(context.Context, ports.UpdateScheduledItemInput) (*model.ScheduledItem, error)
	updateScheduledItemReminderSettingsFn        func(context.Context, ports.UpdateScheduledItemReminderSettingsInput) (*model.ScheduledItem, error)
	deleteScheduledItemFn                        func(context.Context, ports.DeleteScheduledItemInput) error
	upsertHealthScheduledItemFn                  func(context.Context, ports.UpsertHealthScheduledItemInput) (*model.ScheduledItem, error)
	deleteHealthScheduledItemFn                  func(context.Context, ports.DeleteHealthScheduledItemInput) error
	getScheduledItemOccurrenceFn                 func(context.Context, uuid.UUID, uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error)
	listScheduledItemOccurrencesFn               func(context.Context, ports.ListScheduledItemOccurrencesQuery) (ports.ListScheduledItemOccurrencesResult, error)
	listCalendarDayScheduledOccurrencesFn        func(context.Context, uuid.UUID, time.Time, time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
	listCalendarDayScheduledOccurrencesForPetsFn func(context.Context, []uuid.UUID, time.Time, time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
	createScheduledItemOccurrenceFn              func(context.Context, ports.CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error)
	deleteScheduledItemOccurrencesFromFn         func(context.Context, ports.DeleteScheduledItemOccurrencesFromInput) error
	markScheduledItemOccurrencesGeneratedUntilFn func(context.Context, ports.MarkScheduledItemOccurrencesGeneratedUntilInput) error
}

func (s *stubScheduledRepo) GetScheduledItem(ctx context.Context, petID, itemID uuid.UUID) (*model.ScheduledItem, error) {
	if s.getScheduledItemFn != nil {
		return s.getScheduledItemFn(ctx, petID, itemID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubScheduledRepo) GetScheduledItemBySource(ctx context.Context, petID uuid.UUID, sourceType string, sourceID uuid.UUID) (*model.ScheduledItem, error) {
	if s.getScheduledItemBySourceFn != nil {
		return s.getScheduledItemBySourceFn(ctx, petID, sourceType, sourceID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubScheduledRepo) ListScheduledItems(ctx context.Context, query ports.ListScheduledItemsQuery) (ports.ListScheduledItemsResult, error) {
	if s.listScheduledItemsFn != nil {
		return s.listScheduledItemsFn(ctx, query)
	}
	return ports.ListScheduledItemsResult{}, nil
}

func (s *stubScheduledRepo) ListRecurringScheduledItemsForHorizon(ctx context.Context, params ports.ListRecurringScheduledItemsForHorizonParams) ([]model.ScheduledItem, error) {
	if s.listRecurringScheduledItemsForHorizonFn != nil {
		return s.listRecurringScheduledItemsForHorizonFn(ctx, params)
	}
	return []model.ScheduledItem{}, nil
}

func (s *stubScheduledRepo) CreateScheduledItem(ctx context.Context, input ports.CreateScheduledItemInput) (*model.ScheduledItem, error) {
	if s.createScheduledItemFn != nil {
		return s.createScheduledItemFn(ctx, input)
	}
	return &model.ScheduledItem{ID: input.ID, PetID: input.PetID, SourceType: input.SourceType, SourceID: input.SourceID, Title: input.Title, StartsAt: input.StartsAt, PushEnabled: input.PushEnabled, RemindOffsetMinutes: input.RemindOffsetMinutes, RecurrenceRule: input.RecurrenceRule, RecurrenceInterval: input.RecurrenceInterval, RecurrenceUntil: input.RecurrenceUntil, RowVersion: 1}, nil
}

func (s *stubScheduledRepo) UpdateScheduledItem(ctx context.Context, input ports.UpdateScheduledItemInput) (*model.ScheduledItem, error) {
	if s.updateScheduledItemFn != nil {
		return s.updateScheduledItemFn(ctx, input)
	}
	return &model.ScheduledItem{ID: input.ID, PetID: input.PetID, Title: input.Title, StartsAt: input.StartsAt, RowVersion: input.RowVersion + 1}, nil
}

func (s *stubScheduledRepo) UpdateScheduledItemReminderSettings(ctx context.Context, input ports.UpdateScheduledItemReminderSettingsInput) (*model.ScheduledItem, error) {
	if s.updateScheduledItemReminderSettingsFn != nil {
		return s.updateScheduledItemReminderSettingsFn(ctx, input)
	}
	return &model.ScheduledItem{ID: input.ID, PetID: input.PetID, RowVersion: input.RowVersion + 1, PushEnabled: input.PushEnabled, RemindOffsetMinutes: input.RemindOffsetMinutes}, nil
}

func (s *stubScheduledRepo) DeleteScheduledItem(ctx context.Context, input ports.DeleteScheduledItemInput) error {
	if s.deleteScheduledItemFn != nil {
		return s.deleteScheduledItemFn(ctx, input)
	}
	return nil
}

func (s *stubScheduledRepo) UpsertHealthScheduledItem(ctx context.Context, input ports.UpsertHealthScheduledItemInput) (*model.ScheduledItem, error) {
	if s.upsertHealthScheduledItemFn != nil {
		return s.upsertHealthScheduledItemFn(ctx, input)
	}
	sourceID := input.SourceID
	return &model.ScheduledItem{ID: uuid.New(), PetID: input.PetID, SourceType: input.SourceType, SourceID: &sourceID, Title: input.Title, StartsAt: input.StartsAt, RowVersion: 1}, nil
}

func (s *stubScheduledRepo) DeleteHealthScheduledItem(ctx context.Context, input ports.DeleteHealthScheduledItemInput) error {
	if s.deleteHealthScheduledItemFn != nil {
		return s.deleteHealthScheduledItemFn(ctx, input)
	}
	return nil
}

func (s *stubScheduledRepo) GetScheduledItemOccurrence(ctx context.Context, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error) {
	if s.getScheduledItemOccurrenceFn != nil {
		return s.getScheduledItemOccurrenceFn(ctx, petID, occurrenceID)
	}
	return nil, ports.ErrNotFound
}

func (s *stubScheduledRepo) ListScheduledItemOccurrences(ctx context.Context, query ports.ListScheduledItemOccurrencesQuery) (ports.ListScheduledItemOccurrencesResult, error) {
	if s.listScheduledItemOccurrencesFn != nil {
		return s.listScheduledItemOccurrencesFn(ctx, query)
	}
	return ports.ListScheduledItemOccurrencesResult{}, nil
}

func (s *stubScheduledRepo) ListCalendarDayScheduledOccurrences(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error) {
	if s.listCalendarDayScheduledOccurrencesFn != nil {
		return s.listCalendarDayScheduledOccurrencesFn(ctx, petID, dayStart, dayEnd)
	}
	return []model.ScheduledItemOccurrenceListItem{}, nil
}

func (s *stubScheduledRepo) ListCalendarDayScheduledOccurrencesForPets(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error) {
	if s.listCalendarDayScheduledOccurrencesForPetsFn != nil {
		return s.listCalendarDayScheduledOccurrencesForPetsFn(ctx, petIDs, dayStart, dayEnd)
	}
	return []model.ScheduledItemOccurrenceListItem{}, nil
}

func (s *stubScheduledRepo) CreateScheduledItemOccurrence(ctx context.Context, input ports.CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error) {
	if s.createScheduledItemOccurrenceFn != nil {
		return s.createScheduledItemOccurrenceFn(ctx, input)
	}
	return &model.ScheduledItemOccurrence{ID: input.ID, PetID: input.PetID, ScheduledItemID: input.ScheduledItemID, ScheduledFor: input.ScheduledFor}, nil
}

func (s *stubScheduledRepo) DeleteScheduledItemOccurrencesFrom(ctx context.Context, input ports.DeleteScheduledItemOccurrencesFromInput) error {
	if s.deleteScheduledItemOccurrencesFromFn != nil {
		return s.deleteScheduledItemOccurrencesFromFn(ctx, input)
	}
	return nil
}

func (s *stubScheduledRepo) MarkScheduledItemOccurrencesGeneratedUntil(ctx context.Context, input ports.MarkScheduledItemOccurrencesGeneratedUntilInput) error {
	if s.markScheduledItemOccurrencesGeneratedUntilFn != nil {
		return s.markScheduledItemOccurrencesGeneratedUntilFn(ctx, input)
	}
	return nil
}

type stubDictionaryRepo struct {
	getFn         func(context.Context, uuid.UUID, uuid.UUID, string) (*model.HealthDictionaryItem, error)
	getOrCreateFn func(context.Context, ports.GetOrCreateCustomHealthDictionaryItemInput) (*model.HealthDictionaryItem, error)
	listFn        func(context.Context, ports.ListHealthDictionaryItemsInput) ([]model.HealthDictionaryItem, error)
}

func (s *stubDictionaryRepo) ListHealthDictionaryItems(ctx context.Context, in ports.ListHealthDictionaryItemsInput) ([]model.HealthDictionaryItem, error) {
	if s.listFn != nil {
		return s.listFn(ctx, in)
	}
	return []model.HealthDictionaryItem{}, nil
}

func (s *stubDictionaryRepo) GetHealthDictionaryItem(ctx context.Context, petID, itemID uuid.UUID, kind string) (*model.HealthDictionaryItem, error) {
	if s.getFn != nil {
		return s.getFn(ctx, petID, itemID, kind)
	}
	return &model.HealthDictionaryItem{ID: itemID, Kind: kind, Name: "Item"}, nil
}

func (s *stubDictionaryRepo) GetOrCreateCustomHealthDictionaryItem(ctx context.Context, in ports.GetOrCreateCustomHealthDictionaryItemInput) (*model.HealthDictionaryItem, error) {
	if s.getOrCreateFn != nil {
		return s.getOrCreateFn(ctx, in)
	}
	return &model.HealthDictionaryItem{ID: uuid.New(), Kind: in.Kind, Name: in.Name}, nil
}

func expectHealthErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v want %v", got, want)
	}
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

var (
	_ ports.HealthAccessChecker        = (*stubHealthAccess)(nil)
	_ ports.HealthFileClient           = (*stubHealthFileClient)(nil)
	_ ports.LogsRepository             = (*stubLogsRepo)(nil)
	_ ports.ScheduledRepository        = (*stubScheduledRepo)(nil)
	_ ports.HealthDictionaryRepository = (*stubDictionaryRepo)(nil)
)
