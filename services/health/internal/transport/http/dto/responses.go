package dto

type HealthBootstrapPermissionsResponse struct {
	HealthRead  bool `json:"health_read"`
	HealthWrite bool `json:"health_write"`
	LogRead     bool `json:"log_read"`
}

type HealthBootstrapEnumsResponse struct {
	VetVisitStatuses       []string                       `json:"vet_visit_statuses"`
	VetVisitTypes          []string                       `json:"vet_visit_types"`
	VaccinationStatuses    []string                       `json:"vaccination_statuses"`
	VaccinationTargets     []HealthDictionaryItemResponse `json:"vaccination_targets"`
	ProcedureStatuses      []string                       `json:"procedure_statuses"`
	ProcedureTypeItems     []HealthDictionaryItemResponse `json:"procedure_type_items"`
	MedicalRecordTypeItems []HealthDictionaryItemResponse `json:"medical_record_type_items"`
	MedicalRecordStatuses  []string                       `json:"medical_record_statuses"`
}

type HealthBootstrapResponse struct {
	Permissions HealthBootstrapPermissionsResponse `json:"permissions"`
	Enums       HealthBootstrapEnumsResponse       `json:"enums"`
}

type RecurrenceResponse struct {
	Rule     string  `json:"rule"`
	Interval *int    `json:"interval"`
	Until    *string `json:"until"`
}

type RelatedLogResponse struct {
	ID                 string  `json:"id"`
	OccurredAt         string  `json:"occurred_at"`
	LogTypeName        *string `json:"log_type_name"`
	DescriptionPreview *string `json:"description_preview"`
}

type HealthDictionaryItemResponse struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	PetID      *string `json:"pet_id"`
	Code       *string `json:"code"`
	Name       string  `json:"name"`
	IsSystem   bool    `json:"is_system"`
	IsArchived bool    `json:"is_archived"`
}

type HealthAttachmentResponse struct {
	ID            string  `json:"id"`
	FileID        string  `json:"file_id"`
	FileName      *string `json:"file_name"`
	FileType      string  `json:"file_type"`
	DownloadURL   *string `json:"download_url"`
	PreviewURL    *string `json:"preview_url"`
	AddedByUserID string  `json:"added_by_user_id"`
	AddedAt       string  `json:"added_at"`
}

type VetVisitListItemResponse struct {
	ID               string  `json:"id"`
	PetID            string  `json:"pet_id"`
	Status           string  `json:"status"`
	VisitType        string  `json:"visit_type"`
	Title            *string `json:"title"`
	ScheduledAt      *string `json:"scheduled_at"`
	CompletedAt      *string `json:"completed_at"`
	ReasonText       *string `json:"reason_text"`
	ResultText       *string `json:"result_text"`
	ClinicName       *string `json:"clinic_name"`
	VetName          *string `json:"vet_name"`
	RelatedLogsCount int     `json:"related_logs_count"`
	AttachmentsCount int     `json:"attachments_count"`
	RowVersion       int     `json:"row_version"`
	CreatedAt        string  `json:"created_at"`
	CreatedByUserID  string  `json:"created_by_user_id"`
	UpdatedAt        string  `json:"updated_at"`
	UpdatedByUserID  string  `json:"updated_by_user_id"`
}

type VetVisitResponse struct {
	ID            string                     `json:"id"`
	PetID         string                     `json:"pet_id"`
	Status        string                     `json:"status"`
	VisitType     string                     `json:"visit_type"`
	Title         *string                    `json:"title"`
	ScheduledAt   *string                    `json:"scheduled_at"`
	CompletedAt   *string                    `json:"completed_at"`
	ReasonText    *string                    `json:"reason_text"`
	ResultText    *string                    `json:"result_text"`
	ClinicName    *string                    `json:"clinic_name"`
	VetName       *string                    `json:"vet_name"`
	RelatedLogs   []RelatedLogResponse       `json:"related_logs"`
	Attachments   []HealthAttachmentResponse `json:"attachments"`
	RowVersion    int                        `json:"row_version"`
	CreatedAt     string                     `json:"created_at"`
	CreatedByUser string                     `json:"created_by_user_id"`
	UpdatedAt     string                     `json:"updated_at"`
	UpdatedByUser string                     `json:"updated_by_user_id"`
	CanEdit       bool                       `json:"can_edit"`
	CanDelete     bool                       `json:"can_delete"`
}

type VaccinationListItemResponse struct {
	ID                  string                         `json:"id"`
	PetID               string                         `json:"pet_id"`
	GeneratedFromID     *string                        `json:"generated_from_id"`
	Status              string                         `json:"status"`
	VaccineName         string                         `json:"vaccine_name"`
	CatalogMedicationID *string                        `json:"catalog_medication_id"`
	Targets             []HealthDictionaryItemResponse `json:"targets"`
	ScheduledAt         *string                        `json:"scheduled_at"`
	AdministeredAt      *string                        `json:"administered_at"`
	NextDueAt           *string                        `json:"next_due_at"`
	VetVisitID          *string                        `json:"vet_visit_id"`
	ClinicName          *string                        `json:"clinic_name"`
	VetName             *string                        `json:"vet_name"`
	NotesPreview        *string                        `json:"notes_preview"`
	AttachmentsCount    int                            `json:"attachments_count"`
	RowVersion          int                            `json:"row_version"`
	CreatedAt           string                         `json:"created_at"`
	CreatedByUserID     string                         `json:"created_by_user_id"`
	UpdatedAt           string                         `json:"updated_at"`
	UpdatedByUserID     string                         `json:"updated_by_user_id"`
}

type VaccinationResponse struct {
	ID                  string                         `json:"id"`
	PetID               string                         `json:"pet_id"`
	GeneratedFromID     *string                        `json:"generated_from_id"`
	Status              string                         `json:"status"`
	VaccineName         string                         `json:"vaccine_name"`
	CatalogMedicationID *string                        `json:"catalog_medication_id"`
	Targets             []HealthDictionaryItemResponse `json:"targets"`
	ScheduledAt         *string                        `json:"scheduled_at"`
	AdministeredAt      *string                        `json:"administered_at"`
	NextDueAt           *string                        `json:"next_due_at"`
	VetVisitID          *string                        `json:"vet_visit_id"`
	ClinicName          *string                        `json:"clinic_name"`
	VetName             *string                        `json:"vet_name"`
	Notes               *string                        `json:"notes"`
	Attachments         []HealthAttachmentResponse     `json:"attachments"`
	RowVersion          int                            `json:"row_version"`
	CreatedAt           string                         `json:"created_at"`
	CreatedByUserID     string                         `json:"created_by_user_id"`
	UpdatedAt           string                         `json:"updated_at"`
	UpdatedByUserID     string                         `json:"updated_by_user_id"`
	CanEdit             bool                           `json:"can_edit"`
	CanDelete           bool                           `json:"can_delete"`
}

type ProcedureListItemResponse struct {
	ID                  string                        `json:"id"`
	PetID               string                        `json:"pet_id"`
	GeneratedFromID     *string                       `json:"generated_from_id"`
	Status              string                        `json:"status"`
	ProcedureTypeItem   *HealthDictionaryItemResponse `json:"procedure_type_item"`
	Title               string                        `json:"title"`
	DescriptionPreview  *string                       `json:"description_preview"`
	CatalogMedicationID *string                       `json:"catalog_medication_id"`
	ProductName         *string                       `json:"product_name"`
	ScheduledAt         *string                       `json:"scheduled_at"`
	PerformedAt         *string                       `json:"performed_at"`
	NextDueAt           *string                       `json:"next_due_at"`
	VetVisitID          *string                       `json:"vet_visit_id"`
	NotesPreview        *string                       `json:"notes_preview"`
	AttachmentsCount    int                           `json:"attachments_count"`
	RowVersion          int                           `json:"row_version"`
	CreatedAt           string                        `json:"created_at"`
	CreatedByUserID     string                        `json:"created_by_user_id"`
	UpdatedAt           string                        `json:"updated_at"`
	UpdatedByUserID     string                        `json:"updated_by_user_id"`
}

type ProcedureResponse struct {
	ID                  string                        `json:"id"`
	PetID               string                        `json:"pet_id"`
	GeneratedFromID     *string                       `json:"generated_from_id"`
	Status              string                        `json:"status"`
	ProcedureTypeItem   *HealthDictionaryItemResponse `json:"procedure_type_item"`
	Title               string                        `json:"title"`
	Description         *string                       `json:"description"`
	CatalogMedicationID *string                       `json:"catalog_medication_id"`
	ProductName         *string                       `json:"product_name"`
	ScheduledAt         *string                       `json:"scheduled_at"`
	PerformedAt         *string                       `json:"performed_at"`
	NextDueAt           *string                       `json:"next_due_at"`
	VetVisitID          *string                       `json:"vet_visit_id"`
	Notes               *string                       `json:"notes"`
	Attachments         []HealthAttachmentResponse    `json:"attachments"`
	RowVersion          int                           `json:"row_version"`
	CreatedAt           string                        `json:"created_at"`
	CreatedByUserID     string                        `json:"created_by_user_id"`
	UpdatedAt           string                        `json:"updated_at"`
	UpdatedByUserID     string                        `json:"updated_by_user_id"`
	CanEdit             bool                          `json:"can_edit"`
	CanDelete           bool                          `json:"can_delete"`
}

type MedicalRecordListItemResponse struct {
	ID                 string                        `json:"id"`
	PetID              string                        `json:"pet_id"`
	RecordTypeItem     *HealthDictionaryItemResponse `json:"record_type_item"`
	Status             string                        `json:"status"`
	Title              string                        `json:"title"`
	DescriptionPreview *string                       `json:"description_preview"`
	StartedAt          *string                       `json:"started_at"`
	ResolvedAt         *string                       `json:"resolved_at"`
	AttachmentsCount   int                           `json:"attachments_count"`
	RowVersion         int                           `json:"row_version"`
	CreatedAt          string                        `json:"created_at"`
	CreatedByUserID    string                        `json:"created_by_user_id"`
	UpdatedAt          string                        `json:"updated_at"`
	UpdatedByUserID    string                        `json:"updated_by_user_id"`
}

type MedicalRecordResponse struct {
	ID             string                        `json:"id"`
	PetID          string                        `json:"pet_id"`
	RecordTypeItem *HealthDictionaryItemResponse `json:"record_type_item"`
	Status         string                        `json:"status"`
	Title          string                        `json:"title"`
	Description    *string                       `json:"description"`
	StartedAt      *string                       `json:"started_at"`
	ResolvedAt     *string                       `json:"resolved_at"`
	Attachments    []HealthAttachmentResponse    `json:"attachments"`
	RowVersion     int                           `json:"row_version"`
	CreatedAt      string                        `json:"created_at"`
	CreatedByUser  string                        `json:"created_by_user_id"`
	UpdatedAt      string                        `json:"updated_at"`
	UpdatedByUser  string                        `json:"updated_by_user_id"`
	CanEdit        bool                          `json:"can_edit"`
	CanDelete      bool                          `json:"can_delete"`
}

type PetDocumentResponse struct {
	ID            string  `json:"id"`
	FileID        string  `json:"file_id"`
	FileName      *string `json:"file_name"`
	FileType      string  `json:"file_type"`
	DownloadURL   *string `json:"download_url"`
	PreviewURL    *string `json:"preview_url"`
	AddedAt       string  `json:"added_at"`
	AddedByUserID string  `json:"added_by_user_id"`
	EntityType    string  `json:"entity_type"`
	EntityID      string  `json:"entity_id"`
}

type ScheduledItemListItemResponse struct {
	ID                  string              `json:"id"`
	PetID               string              `json:"pet_id"`
	SourceType          string              `json:"source_type"`
	SourceID            *string             `json:"source_id"`
	Title               string              `json:"title"`
	NotePreview         *string             `json:"note_preview"`
	StartsAt            string              `json:"starts_at"`
	PushEnabled         *bool               `json:"push_enabled"`
	RemindOffsetMinutes *int                `json:"remind_offset_minutes"`
	Recurrence          *RecurrenceResponse `json:"recurrence"`
	RowVersion          int                 `json:"row_version"`
	CreatedAt           string              `json:"created_at"`
	CreatedByUserID     string              `json:"created_by_user_id"`
	UpdatedAt           string              `json:"updated_at"`
	UpdatedByUserID     string              `json:"updated_by_user_id"`
}

type ScheduledItemResponse struct {
	ID                  string              `json:"id"`
	PetID               string              `json:"pet_id"`
	SourceType          string              `json:"source_type"`
	SourceID            *string             `json:"source_id"`
	Title               string              `json:"title"`
	Note                *string             `json:"note"`
	StartsAt            string              `json:"starts_at"`
	PushEnabled         *bool               `json:"push_enabled"`
	RemindOffsetMinutes *int                `json:"remind_offset_minutes"`
	Recurrence          *RecurrenceResponse `json:"recurrence"`
	RowVersion          int                 `json:"row_version"`
	CreatedAt           string              `json:"created_at"`
	CreatedByUserID     string              `json:"created_by_user_id"`
	UpdatedAt           string              `json:"updated_at"`
	UpdatedByUserID     string              `json:"updated_by_user_id"`
}

type ScheduledOccurrenceResponse struct {
	ID              string                `json:"id"`
	ScheduledItemID string                `json:"scheduled_item_id"`
	PetID           string                `json:"pet_id"`
	ScheduledFor    string                `json:"scheduled_for"`
	CreatedAt       string                `json:"created_at"`
	Rule            ScheduledItemResponse `json:"rule"`
}

type CalendarDayItemResponse struct {
	ItemType              string  `json:"item_type"`
	EntityID              string  `json:"entity_id"`
	PetID                 string  `json:"pet_id"`
	Title                 string  `json:"title"`
	Subtitle              *string `json:"subtitle"`
	ScheduledFor          string  `json:"scheduled_for"`
	Status                string  `json:"status"`
	VisitID               *string `json:"visit_id"`
	VaccinationID         *string `json:"vaccination_id"`
	ProcedureID           *string `json:"procedure_id"`
	ScheduledItemID       *string `json:"scheduled_item_id"`
	ScheduledOccurrenceID *string `json:"scheduled_occurrence_id"`
}

type LogListItemResponse struct {
	ID                  string                   `json:"id"`
	PetID               string                   `json:"pet_id"`
	OccurredAt          string                   `json:"occurred_at"`
	LogTypeID           *string                  `json:"log_type_id"`
	LogTypeName         *string                  `json:"log_type_name"`
	LogTypeScope        *string                  `json:"log_type_scope"`
	DescriptionPreview  *string                  `json:"description_preview"`
	RelatedEntityType   *string                  `json:"related_entity_type"`
	RelatedEntityID     *string                  `json:"related_entity_id"`
	MetricValuesPreview []LogMetricValueResponse `json:"metric_values_preview"`
	AttachmentsCount    int                      `json:"attachments_count"`
	HasAttachments      bool                     `json:"has_attachments"`
	CreatedByUserID     string                   `json:"created_by_user_id"`
}

type LogMetricValueResponse struct {
	MetricID   string  `json:"metric_id"`
	MetricName string  `json:"metric_name"`
	InputKind  string  `json:"input_kind"`
	Unit       *string `json:"unit"`
	ValueNum   float64 `json:"value_num"`
}

type LogAttachmentResponse struct {
	ID            string  `json:"id"`
	FileID        string  `json:"file_id"`
	FileName      *string `json:"file_name"`
	FileType      string  `json:"file_type"`
	DownloadURL   *string `json:"download_url"`
	PreviewURL    *string `json:"preview_url"`
	AddedAt       string  `json:"added_at"`
	AddedByUserID string  `json:"added_by_user_id"`
}

type LogResponse struct {
	ID                string                   `json:"id"`
	PetID             string                   `json:"pet_id"`
	OccurredAt        string                   `json:"occurred_at"`
	LogTypeID         *string                  `json:"log_type_id"`
	LogTypeName       *string                  `json:"log_type_name"`
	LogTypeScope      *string                  `json:"log_type_scope"`
	Description       *string                  `json:"description"`
	RelatedEntityType *string                  `json:"related_entity_type"`
	RelatedEntityID   *string                  `json:"related_entity_id"`
	MetricValues      []LogMetricValueResponse `json:"metric_values"`
	Attachments       []LogAttachmentResponse  `json:"attachments"`
	RowVersion        int                      `json:"row_version"`
	CreatedAt         string                   `json:"created_at"`
	CreatedByUserID   string                   `json:"created_by_user_id"`
	UpdatedAt         string                   `json:"updated_at"`
	UpdatedByUserID   string                   `json:"updated_by_user_id"`
	CanEdit           bool                     `json:"can_edit"`
	CanDelete         bool                     `json:"can_delete"`
}

type LogTypeMetricRequirementResponse struct {
	MetricID    string   `json:"metric_id"`
	MetricName  *string  `json:"metric_name"`
	MetricScope *string  `json:"metric_scope"`
	InputKind   *string  `json:"input_kind"`
	Unit        *string  `json:"unit"`
	MinValue    *float64 `json:"min_value"`
	MaxValue    *float64 `json:"max_value"`
	IsRequired  bool     `json:"is_required"`
}

type LogTypeResponse struct {
	ID                 string                             `json:"id"`
	Scope              string                             `json:"scope"`
	PetID              *string                            `json:"pet_id"`
	Code               *string                            `json:"code"`
	Name               string                             `json:"name"`
	MetricRequirements []LogTypeMetricRequirementResponse `json:"metric_requirements"`
	CreatedAt          string                             `json:"created_at"`
	CreatedByUserID    *string                            `json:"created_by_user_id"`
	UpdatedAt          string                             `json:"updated_at"`
	UpdatedByUserID    *string                            `json:"updated_by_user_id"`
	RowVersion         int                                `json:"row_version"`
	IsArchived         bool                               `json:"is_archived"`
}

type MetricUsageResponse struct {
	LogTypesCount int `json:"log_types_count"`
	LogsCount     int `json:"logs_count"`
}

type MetricResponse struct {
	ID              string              `json:"id"`
	Scope           string              `json:"scope"`
	PetID           *string             `json:"pet_id"`
	Code            *string             `json:"code"`
	Name            string              `json:"name"`
	InputKind       string              `json:"input_kind"`
	Unit            *string             `json:"unit"`
	MinValue        *float64            `json:"min_value"`
	MaxValue        *float64            `json:"max_value"`
	CreatedAt       string              `json:"created_at"`
	CreatedByUserID *string             `json:"created_by_user_id"`
	UpdatedAt       string              `json:"updated_at"`
	UpdatedByUserID *string             `json:"updated_by_user_id"`
	RowVersion      int                 `json:"row_version"`
	IsArchived      bool                `json:"is_archived"`
	Usage           MetricUsageResponse `json:"usage"`
}

type AnalyticsUsedLogTypeResponse struct {
	LogTypeID   string `json:"log_type_id"`
	LogTypeName string `json:"log_type_name"`
}

type AnalyticsMetricSummaryResponse struct {
	MetricID        string                         `json:"metric_id"`
	MetricName      string                         `json:"metric_name"`
	MetricScope     string                         `json:"metric_scope"`
	InputKind       string                         `json:"input_kind"`
	Unit            *string                        `json:"unit"`
	PointsCount     int                            `json:"points_count"`
	FirstOccurredAt string                         `json:"first_occurred_at"`
	LastOccurredAt  string                         `json:"last_occurred_at"`
	LastValueNum    float64                        `json:"last_value_num"`
	UsedInLogTypes  []AnalyticsUsedLogTypeResponse `json:"used_in_log_types"`
}

type MetricSeriesPointResponse struct {
	OccurredAt  string  `json:"occurred_at"`
	ValueNum    float64 `json:"value_num"`
	LogID       string  `json:"log_id"`
	LogTypeID   *string `json:"log_type_id"`
	LogTypeName *string `json:"log_type_name"`
}

type MetricSeriesSummaryResponse struct {
	PointsCount       int     `json:"points_count"`
	MinValueNum       float64 `json:"min_value_num"`
	MaxValueNum       float64 `json:"max_value_num"`
	LastValueNum      float64 `json:"last_value_num"`
	AvgValueNum       float64 `json:"avg_value_num"`
	DeltaFromFirstNum float64 `json:"delta_from_first_num"`
}

type PermissionsResponse struct {
	LogRead  bool `json:"log_read,omitempty"`
	LogWrite bool `json:"log_write,omitempty"`
}

type LogsBootstrapResponse struct {
	Permissions    PermissionsResponse `json:"permissions"`
	RecentLogTypes []LogTypeResponse   `json:"recent_log_types"`
	SystemLogTypes []LogTypeResponse   `json:"system_log_types"`
	CustomLogTypes []LogTypeResponse   `json:"custom_log_types"`
	SystemMetrics  []MetricResponse    `json:"system_metrics"`
	CustomMetrics  []MetricResponse    `json:"custom_metrics"`
}

type LogFacetsResponse struct {
	Types               []any `json:"types"`
	HasAttachmentsCount int   `json:"has_attachments_count"`
	HasMetricsCount     int   `json:"has_metrics_count"`
}

type LogsListResponse struct {
	Items      []LogListItemResponse `json:"items"`
	NextCursor *string               `json:"next_cursor"`
	Facets     *LogFacetsResponse    `json:"facets,omitempty"`
}

type LogTypesListResponse struct {
	Items []LogTypeResponse `json:"items"`
}

type MetricsListResponse struct {
	Items []MetricResponse `json:"items"`
}

type AnalyticsMetricsListResponse struct {
	Items []AnalyticsMetricSummaryResponse `json:"items"`
}

type MetricSeriesResponse struct {
	Metric  MetricResponse               `json:"metric"`
	Summary *MetricSeriesSummaryResponse `json:"summary"`
	Points  []MetricSeriesPointResponse  `json:"points"`
}

type DayResponse struct {
	Date  string                    `json:"date"`
	Items []CalendarDayItemResponse `json:"items"`
}

type CalendarMarkerResponse struct {
	Date           string `json:"date"`
	PlannedCount   int    `json:"planned_count"`
	CompletedCount int    `json:"completed_count"`
	TotalCount     int    `json:"total_count"`
}

type CalendarRangeResponse struct {
	DateFrom string                   `json:"date_from"`
	DateTo   string                   `json:"date_to"`
	Items    []CalendarMarkerResponse `json:"items"`
}

type ScheduledItemsListResponse struct {
	Items      []ScheduledItemListItemResponse `json:"items"`
	NextCursor *string                         `json:"next_cursor"`
}

type ScheduledOccurrencesListResponse struct {
	Items      []ScheduledOccurrenceResponse `json:"items"`
	NextCursor *string                       `json:"next_cursor"`
}

type UploadResponse struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

type InitAttachmentUploadResponse struct {
	FileID string         `json:"file_id"`
	Upload UploadResponse `json:"upload"`
}

type UploadedFileResponse struct {
	ID               string  `json:"id"`
	MimeType         string  `json:"mime_type"`
	SizeBytes        int64   `json:"size_bytes"`
	OriginalFilename *string `json:"original_filename"`
}

type ConfirmAttachmentUploadResponse struct {
	File UploadedFileResponse `json:"file"`
}

type PetDocumentsListResponse struct {
	Items      []PetDocumentResponse `json:"items"`
	NextCursor *string               `json:"next_cursor"`
}

type VetVisitsListResponse struct {
	Items      []VetVisitListItemResponse `json:"items"`
	NextCursor *string                    `json:"next_cursor"`
}

type VaccinationsListResponse struct {
	Items      []VaccinationListItemResponse `json:"items"`
	NextCursor *string                       `json:"next_cursor"`
}

type ProceduresListResponse struct {
	Items      []ProcedureListItemResponse `json:"items"`
	NextCursor *string                     `json:"next_cursor"`
}

type MedicalRecordsListResponse struct {
	Items      []MedicalRecordListItemResponse `json:"items"`
	NextCursor *string                         `json:"next_cursor"`
}
