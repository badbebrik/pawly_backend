package dto

type DeleteRowVersionRequest struct {
	RowVersion int `json:"row_version"`
}

type MedicalEntityReminderRequest struct {
	PushEnabled         bool `json:"push_enabled"`
	RemindOffsetMinutes *int `json:"remind_offset_minutes"`
}

type AttachmentRequest struct {
	FileID   string  `json:"file_id"`
	FileName *string `json:"file_name"`
}

type HealthDictionaryItemRefRequest struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
}

type CreateOrUpdateVetVisitRequest struct {
	Status      string                        `json:"status"`
	VisitType   string                        `json:"visit_type"`
	Title       *string                       `json:"title"`
	ScheduledAt *string                       `json:"scheduled_at"`
	Reminder    *MedicalEntityReminderRequest `json:"reminder"`
	CompletedAt *string                       `json:"completed_at"`
	ReasonText  *string                       `json:"reason_text"`
	ResultText  *string                       `json:"result_text"`
	ClinicName  *string                       `json:"clinic_name"`
	VetName     *string                       `json:"vet_name"`
	Attachments []AttachmentRequest           `json:"attachments"`
	RowVersion  int                           `json:"row_version"`
}

type LinkVetVisitLogRequest struct {
	LogID string `json:"log_id"`
}

type CreateOrUpdateVaccinationRequest struct {
	Status              string                           `json:"status"`
	VaccineName         string                           `json:"vaccine_name"`
	CatalogMedicationID *string                          `json:"catalog_medication_id"`
	Targets             []HealthDictionaryItemRefRequest `json:"targets"`
	ScheduledAt         *string                          `json:"scheduled_at"`
	Reminder            *MedicalEntityReminderRequest    `json:"reminder"`
	AdministeredAt      *string                          `json:"administered_at"`
	NextDueAt           *string                          `json:"next_due_at"`
	VetVisitID          *string                          `json:"vet_visit_id"`
	ClinicName          *string                          `json:"clinic_name"`
	VetName             *string                          `json:"vet_name"`
	Notes               *string                          `json:"notes"`
	Attachments         []AttachmentRequest              `json:"attachments"`
	RowVersion          int                              `json:"row_version"`
}

type CreateOrUpdateProcedureRequest struct {
	Status              string                        `json:"status"`
	ProcedureTypeID     *string                       `json:"procedure_type_id"`
	ProcedureTypeName   *string                       `json:"procedure_type_name"`
	Title               string                        `json:"title"`
	Description         *string                       `json:"description"`
	CatalogMedicationID *string                       `json:"catalog_medication_id"`
	ProductName         *string                       `json:"product_name"`
	ScheduledAt         *string                       `json:"scheduled_at"`
	Reminder            *MedicalEntityReminderRequest `json:"reminder"`
	PerformedAt         *string                       `json:"performed_at"`
	NextDueAt           *string                       `json:"next_due_at"`
	VetVisitID          *string                       `json:"vet_visit_id"`
	Notes               *string                       `json:"notes"`
	Attachments         []AttachmentRequest           `json:"attachments"`
	RowVersion          int                           `json:"row_version"`
}

type CreateOrUpdateMedicalRecordRequest struct {
	RecordTypeID   *string             `json:"record_type_id"`
	RecordTypeName *string             `json:"record_type_name"`
	Status         string              `json:"status"`
	Title          string              `json:"title"`
	Description    *string             `json:"description"`
	StartedAt      *string             `json:"started_at"`
	ResolvedAt     *string             `json:"resolved_at"`
	Attachments    []AttachmentRequest `json:"attachments"`
	RowVersion     int                 `json:"row_version"`
}

type InitAttachmentUploadRequest struct {
	MimeType          string  `json:"mime_type"`
	OriginalFilename  string  `json:"original_filename"`
	ExpectedSizeBytes int64   `json:"expected_size_bytes"`
	EntityType        *string `json:"entity_type"`
}

type ConfirmAttachmentUploadRequest struct {
	FileID     string  `json:"file_id"`
	SizeBytes  int64   `json:"size_bytes"`
	EntityType *string `json:"entity_type"`
}

type RenameDocumentRequest struct {
	FileName string `json:"file_name"`
}

type RecurrenceRequest struct {
	Rule     *string `json:"rule"`
	Interval *int    `json:"interval"`
	Until    *string `json:"until"`
}

type CreateOrUpdateScheduledItemRequest struct {
	SourceType          string             `json:"source_type"`
	SourceID            *string            `json:"source_id"`
	Title               string             `json:"title"`
	Note                *string            `json:"note"`
	StartsAt            *string            `json:"starts_at"`
	PushEnabled         *bool              `json:"push_enabled"`
	RemindOffsetMinutes *int               `json:"remind_offset_minutes"`
	Recurrence          *RecurrenceRequest `json:"recurrence"`
	RowVersion          int                `json:"row_version"`
}

type UpdateScheduledItemReminderSettingsRequest struct {
	PushEnabled         bool `json:"push_enabled"`
	RemindOffsetMinutes *int `json:"remind_offset_minutes"`
	RowVersion          int  `json:"row_version"`
}
