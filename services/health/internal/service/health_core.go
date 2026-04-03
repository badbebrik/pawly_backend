package service

import (
	"context"
	"health/internal/model"
	"health/internal/repository"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HealthBootstrapParams struct {
	UserID uuid.UUID
	PetID  uuid.UUID
}

type ListVetVisitsParams struct {
	UserID   uuid.UUID
	PetID    uuid.UUID
	Cursor   *repository.TimeCursor
	Limit    int
	Status   string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type CreateVetVisitParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	Status            string
	VisitType         string
	ScheduledAt       *time.Time
	CompletedAt       *time.Time
	ReasonText        *string
	ResultText        *string
	ClinicName        *string
	VetName           *string
	AttachmentFileIDs []uuid.UUID
}

type UpdateVetVisitParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	VisitID           uuid.UUID
	RowVersion        int
	Status            string
	VisitType         string
	ScheduledAt       *time.Time
	CompletedAt       *time.Time
	ReasonText        *string
	ResultText        *string
	ClinicName        *string
	VetName           *string
	AttachmentFileIDs []uuid.UUID
}

type DeleteVetVisitParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	VisitID    uuid.UUID
	RowVersion int
}

type ListVaccinationsParams struct {
	UserID   uuid.UUID
	PetID    uuid.UUID
	Cursor   *repository.TimeCursor
	Limit    int
	Status   string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type CreateVaccinationParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	ScheduledAt         *time.Time
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	Notes               *string
	AttachmentFileIDs   []uuid.UUID
}

type UpdateVaccinationParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	VaccinationID       uuid.UUID
	RowVersion          int
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	ScheduledAt         *time.Time
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	Notes               *string
	AttachmentFileIDs   []uuid.UUID
}

type DeleteVaccinationParams struct {
	UserID        uuid.UUID
	PetID         uuid.UUID
	VaccinationID uuid.UUID
	RowVersion    int
}

type ListProceduresParams struct {
	UserID        uuid.UUID
	PetID         uuid.UUID
	Cursor        *repository.TimeCursor
	Limit         int
	Status        string
	Bucket        string
	ProcedureType string
	DateFrom      *time.Time
	DateTo        *time.Time
	Sort          string
}

type CreateProcedureParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	Status              string
	ProcedureType       string
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	Notes               *string
	AttachmentFileIDs   []uuid.UUID
}

type UpdateProcedureParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	ProcedureID         uuid.UUID
	RowVersion          int
	Status              string
	ProcedureType       string
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	Notes               *string
	AttachmentFileIDs   []uuid.UUID
}

type DeleteProcedureParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	ProcedureID uuid.UUID
	RowVersion  int
}

type ListMedicalRecordsParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	Cursor     *repository.TimeCursor
	Limit      int
	Status     string
	Bucket     string
	RecordType string
	Sort       string
}

type CreateMedicalRecordParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	RecordType        string
	Status            string
	Title             string
	Description       *string
	StartedAt         *time.Time
	ResolvedAt        *time.Time
	AttachmentFileIDs []uuid.UUID
}

type UpdateMedicalRecordParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	RecordID          uuid.UUID
	RowVersion        int
	RecordType        string
	Status            string
	Title             string
	Description       *string
	StartedAt         *time.Time
	ResolvedAt        *time.Time
	AttachmentFileIDs []uuid.UUID
}

type DeleteMedicalRecordParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	RecordID   uuid.UUID
	RowVersion int
}

func (s *Service) GetHealthBootstrap(ctx context.Context, p HealthBootstrapParams) (*model.HealthBootstrap, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	readAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthRead)
	if err != nil {
		return nil, err
	}
	if !readAllowed {
		return nil, ErrForbidden
	}
	writeAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthWrite)
	if err != nil {
		return nil, err
	}
	logAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	return &model.HealthBootstrap{
		Permissions: model.HealthPermissions{HealthRead: true, HealthWrite: writeAllowed, LogRead: logAllowed},
		Enums: model.HealthEnums{
			VetVisitStatuses:      []string{"PLANNED", "COMPLETED", "CANCELLED"},
			VetVisitTypes:         []string{"CHECKUP", "SYMPTOM", "FOLLOW_UP", "VACCINATION", "PROCEDURE", "OTHER"},
			VaccinationStatuses:   []string{"PLANNED", "DONE", "CANCELLED"},
			ProcedureStatuses:     []string{"PLANNED", "DONE", "CANCELLED"},
			ProcedureTypes:        []string{"PARASITE_TREATMENT", "DEWORMING", "HYGIENE", "WOUND_CARE", "GROOMING", "OTHER"},
			MedicalRecordTypes:    []string{"DIAGNOSIS", "ALLERGY", "CHRONIC_CONDITION", "INJURY", "SURGERY", "CLINICAL_NOTE"},
			MedicalRecordStatuses: []string{"ACTIVE", "RESOLVED"},
		},
	}, nil
}

func (s *Service) ListVetVisits(ctx context.Context, p ListVetVisitsParams) (repository.ListVetVisitsOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListVetVisitsOutput{}, ErrInvalidInput
	}
	canRead, err := s.requireHealthRead(ctx, p.PetID, p.UserID)
	if err != nil || !canRead {
		return repository.ListVetVisitsOutput{}, err
	}
	canLogRead, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return repository.ListVetVisitsOutput{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "COMPLETED", "CANCELLED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return repository.ListVetVisitsOutput{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "upcoming", "history", "all"})
	if bucket == "INVALID" {
		return repository.ListVetVisitsOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListVetVisits(ctx, repository.ListVetVisitsInput{
		PetID:    p.PetID,
		Cursor:   p.Cursor,
		Limit:    p.Limit,
		Status:   status,
		Bucket:   bucket,
		DateFrom: p.DateFrom,
		DateTo:   p.DateTo,
		Sort:     p.Sort,
	})
	if err != nil {
		return repository.ListVetVisitsOutput{}, mapRepoErr(err)
	}
	if !canLogRead {
		for i := range out.Items {
			out.Items[i].RelatedLogsCount = 0
		}
	}
	return out, nil
}

func (s *Service) GetVetVisit(ctx context.Context, userID, petID, visitID uuid.UUID) (*model.VetVisit, error) {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	canRead, err := s.requireHealthRead(ctx, petID, userID)
	if err != nil || !canRead {
		return nil, err
	}
	canLogRead, err := s.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetVetVisit(ctx, petID, visitID, canLogRead)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	if !canLogRead {
		item.RelatedLogs = []model.RelatedLog{}
	}
	return item, nil
}

func (s *Service) CreateVetVisit(ctx context.Context, p CreateVetVisitParams) (*model.VetVisit, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, visitType, err := validateVetVisitState(p.Status, p.VisitType)
	if err != nil {
		return nil, err
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.CreateVetVisit(ctx, repository.CreateVetVisitInput{
		ID:          uuid.New(),
		PetID:       p.PetID,
		Status:      status,
		VisitType:   visitType,
		ScheduledAt: p.ScheduledAt,
		CompletedAt: p.CompletedAt,
		ReasonText:  trimOrNil(p.ReasonText),
		ResultText:  trimOrNil(p.ResultText),
		ClinicName:  trimOrNil(p.ClinicName),
		VetName:     trimOrNil(p.VetName),
		CreatedBy:   p.UserID,
		UpdatedBy:   p.UserID,
		Attachments: attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "VET_VISIT", item.ID, sync); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) UpdateVetVisit(ctx context.Context, p UpdateVetVisitParams) (*model.VetVisit, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VisitID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, visitType, err := validateVetVisitState(p.Status, p.VisitType)
	if err != nil {
		return nil, err
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.UpdateVetVisit(ctx, repository.UpdateVetVisitInput{
		ID:          p.VisitID,
		PetID:       p.PetID,
		RowVersion:  p.RowVersion,
		Status:      status,
		VisitType:   visitType,
		ScheduledAt: p.ScheduledAt,
		CompletedAt: p.CompletedAt,
		ReasonText:  trimOrNil(p.ReasonText),
		ResultText:  trimOrNil(p.ResultText),
		ClinicName:  trimOrNil(p.ClinicName),
		VetName:     trimOrNil(p.VetName),
		UpdatedBy:   p.UserID,
		Attachments: attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "VET_VISIT", item.ID, sync); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) DeleteVetVisit(ctx context.Context, p DeleteVetVisitParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VisitID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := s.repo.GetVetVisit(ctx, p.PetID, p.VisitID, false)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.repo.DeleteVetVisit(ctx, repository.DeleteVetVisitInput{ID: p.VisitID, PetID: p.PetID, RowVersion: p.RowVersion, DeletedBy: p.UserID}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "VET_VISIT", p.VisitID, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) LinkVetVisitLog(ctx context.Context, userID, petID, visitID, logID uuid.UUID) (*model.RelatedLog, error) {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil || logID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, petID, userID); err != nil {
		return nil, err
	}
	allowed, err := s.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	item, err := s.repo.LinkVetVisitLog(ctx, petID, visitID, logID, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) UnlinkVetVisitLog(ctx context.Context, userID, petID, visitID, logID uuid.UUID) error {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil || logID == uuid.Nil {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, petID, userID); err != nil {
		return err
	}
	allowed, err := s.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return mapRepoErr(s.repo.UnlinkVetVisitLog(ctx, petID, visitID, logID))
}

func (s *Service) ListVaccinations(ctx context.Context, p ListVaccinationsParams) (repository.ListVaccinationsOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListVaccinationsOutput{}, ErrInvalidInput
	}
	canRead, err := s.requireHealthRead(ctx, p.PetID, p.UserID)
	if err != nil || !canRead {
		return repository.ListVaccinationsOutput{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "DONE", "CANCELLED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return repository.ListVaccinationsOutput{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "planned", "history", "all"})
	if bucket == "INVALID" {
		return repository.ListVaccinationsOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListVaccinations(ctx, repository.ListVaccinationsInput{PetID: p.PetID, Cursor: p.Cursor, Limit: p.Limit, Status: status, Bucket: bucket, DateFrom: p.DateFrom, DateTo: p.DateTo, Sort: p.Sort})
	if err != nil {
		return repository.ListVaccinationsOutput{}, mapRepoErr(err)
	}
	return out, nil
}

func (s *Service) GetVaccination(ctx context.Context, userID, petID, vaccinationID uuid.UUID) (*model.Vaccination, error) {
	if userID == uuid.Nil || petID == uuid.Nil || vaccinationID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetVaccination(ctx, petID, vaccinationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) CreateVaccination(ctx context.Context, p CreateVaccinationParams) (*model.Vaccination, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "DONE", "CANCELLED"})
	if err != nil {
		return nil, err
	}
	vaccineName := strings.TrimSpace(p.VaccineName)
	if vaccineName == "" {
		return nil, ErrInvalidInput
	}
	if err := s.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.CreateVaccination(ctx, repository.CreateVaccinationInput{
		ID:                  uuid.New(),
		PetID:               p.PetID,
		Status:              status,
		VaccineName:         vaccineName,
		CatalogMedicationID: p.CatalogMedicationID,
		ScheduledAt:         p.ScheduledAt,
		AdministeredAt:      p.AdministeredAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		ClinicName:          trimOrNil(p.ClinicName),
		VetName:             trimOrNil(p.VetName),
		Notes:               trimOrNil(p.Notes),
		CreatedBy:           p.UserID,
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "VACCINATION", item.ID, sync); err != nil {
		return nil, err
	}
	if err := s.syncVaccinationAutolog(ctx, item, p.UserID); err != nil {
		return nil, err
	}
	if err := s.ensureNextVaccination(ctx, p.PetID, "", item, p.UserID); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) UpdateVaccination(ctx context.Context, p UpdateVaccinationParams) (*model.Vaccination, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VaccinationID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "DONE", "CANCELLED"})
	if err != nil {
		return nil, err
	}
	vaccineName := strings.TrimSpace(p.VaccineName)
	if vaccineName == "" {
		return nil, ErrInvalidInput
	}
	if err := s.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetVaccination(ctx, p.PetID, p.VaccinationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.UpdateVaccination(ctx, repository.UpdateVaccinationInput{
		ID:                  p.VaccinationID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		Status:              status,
		VaccineName:         vaccineName,
		CatalogMedicationID: p.CatalogMedicationID,
		ScheduledAt:         p.ScheduledAt,
		AdministeredAt:      p.AdministeredAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		ClinicName:          trimOrNil(p.ClinicName),
		VetName:             trimOrNil(p.VetName),
		Notes:               trimOrNil(p.Notes),
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "VACCINATION", item.ID, sync); err != nil {
		return nil, err
	}
	if err := s.syncVaccinationAutologTransition(ctx, current, item, p.UserID); err != nil {
		return nil, err
	}
	if err := s.ensureNextVaccination(ctx, p.PetID, current.Status, item, p.UserID); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) DeleteVaccination(ctx context.Context, p DeleteVaccinationParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VaccinationID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := s.repo.GetVaccination(ctx, p.PetID, p.VaccinationID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.repo.DeleteVaccination(ctx, repository.DeleteVaccinationInput{ID: p.VaccinationID, PetID: p.PetID, RowVersion: p.RowVersion, DeletedBy: p.UserID}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "VACCINATION", p.VaccinationID, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListProcedures(ctx context.Context, p ListProceduresParams) (repository.ListProceduresOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListProceduresOutput{}, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, p.PetID, p.UserID); err != nil {
		return repository.ListProceduresOutput{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "DONE", "CANCELLED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return repository.ListProceduresOutput{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "planned", "history", "all"})
	if bucket == "INVALID" {
		return repository.ListProceduresOutput{}, ErrInvalidInput
	}
	procedureType := normalizeEnumPtr(p.ProcedureType, []string{"PARASITE_TREATMENT", "DEWORMING", "HYGIENE", "WOUND_CARE", "GROOMING", "OTHER"})
	if strings.TrimSpace(p.ProcedureType) != "" && procedureType == nil {
		return repository.ListProceduresOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListProcedures(ctx, repository.ListProceduresInput{PetID: p.PetID, Cursor: p.Cursor, Limit: p.Limit, Status: status, Bucket: bucket, ProcedureType: procedureType, DateFrom: p.DateFrom, DateTo: p.DateTo, Sort: p.Sort})
	if err != nil {
		return repository.ListProceduresOutput{}, mapRepoErr(err)
	}
	return out, nil
}

func (s *Service) GetProcedure(ctx context.Context, userID, petID, procedureID uuid.UUID) (*model.Procedure, error) {
	if userID == uuid.Nil || petID == uuid.Nil || procedureID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetProcedure(ctx, petID, procedureID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) CreateProcedure(ctx context.Context, p CreateProcedureParams) (*model.Procedure, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "DONE", "CANCELLED"})
	if err != nil {
		return nil, err
	}
	procedureType, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.ProcedureType)), []string{"PARASITE_TREATMENT", "DEWORMING", "HYGIENE", "WOUND_CARE", "GROOMING", "OTHER"})
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if err := s.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.CreateProcedure(ctx, repository.CreateProcedureInput{
		ID:                  uuid.New(),
		PetID:               p.PetID,
		Status:              status,
		ProcedureType:       procedureType,
		Title:               title,
		Description:         trimOrNil(p.Description),
		CatalogMedicationID: p.CatalogMedicationID,
		ProductName:         trimOrNil(p.ProductName),
		ScheduledAt:         p.ScheduledAt,
		PerformedAt:         p.PerformedAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		Notes:               trimOrNil(p.Notes),
		CreatedBy:           p.UserID,
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "PROCEDURE", item.ID, sync); err != nil {
		return nil, err
	}
	if err := s.syncProcedureAutolog(ctx, item, p.UserID); err != nil {
		return nil, err
	}
	if err := s.ensureNextProcedure(ctx, p.PetID, "", item, p.UserID); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) UpdateProcedure(ctx context.Context, p UpdateProcedureParams) (*model.Procedure, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ProcedureID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "DONE", "CANCELLED"})
	if err != nil {
		return nil, err
	}
	procedureType, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.ProcedureType)), []string{"PARASITE_TREATMENT", "DEWORMING", "HYGIENE", "WOUND_CARE", "GROOMING", "OTHER"})
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if err := s.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetProcedure(ctx, p.PetID, p.ProcedureID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.UpdateProcedure(ctx, repository.UpdateProcedureInput{
		ID:                  p.ProcedureID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		Status:              status,
		ProcedureType:       procedureType,
		Title:               title,
		Description:         trimOrNil(p.Description),
		CatalogMedicationID: p.CatalogMedicationID,
		ProductName:         trimOrNil(p.ProductName),
		ScheduledAt:         p.ScheduledAt,
		PerformedAt:         p.PerformedAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		Notes:               trimOrNil(p.Notes),
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "PROCEDURE", item.ID, sync); err != nil {
		return nil, err
	}
	if err := s.syncProcedureAutologTransition(ctx, current, item, p.UserID); err != nil {
		return nil, err
	}
	if err := s.ensureNextProcedure(ctx, p.PetID, current.Status, item, p.UserID); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) DeleteProcedure(ctx context.Context, p DeleteProcedureParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ProcedureID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := s.repo.GetProcedure(ctx, p.PetID, p.ProcedureID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.repo.DeleteProcedure(ctx, repository.DeleteProcedureInput{ID: p.ProcedureID, PetID: p.PetID, RowVersion: p.RowVersion, DeletedBy: p.UserID}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "PROCEDURE", p.ProcedureID, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListMedicalRecords(ctx context.Context, p ListMedicalRecordsParams) (repository.ListMedicalRecordsOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListMedicalRecordsOutput{}, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, p.PetID, p.UserID); err != nil {
		return repository.ListMedicalRecordsOutput{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"ACTIVE", "RESOLVED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return repository.ListMedicalRecordsOutput{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "active", "archive", "all"})
	if bucket == "INVALID" {
		return repository.ListMedicalRecordsOutput{}, ErrInvalidInput
	}
	recordType := normalizeEnumPtr(p.RecordType, []string{"DIAGNOSIS", "ALLERGY", "CHRONIC_CONDITION", "INJURY", "SURGERY", "CLINICAL_NOTE"})
	if strings.TrimSpace(p.RecordType) != "" && recordType == nil {
		return repository.ListMedicalRecordsOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListMedicalRecords(ctx, repository.ListMedicalRecordsInput{PetID: p.PetID, Cursor: p.Cursor, Limit: p.Limit, Status: status, Bucket: bucket, RecordType: recordType, Sort: p.Sort})
	if err != nil {
		return repository.ListMedicalRecordsOutput{}, mapRepoErr(err)
	}
	return out, nil
}

func (s *Service) GetMedicalRecord(ctx context.Context, userID, petID, recordID uuid.UUID) (*model.MedicalRecord, error) {
	if userID == uuid.Nil || petID == uuid.Nil || recordID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetMedicalRecord(ctx, petID, recordID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) CreateMedicalRecord(ctx context.Context, p CreateMedicalRecordParams) (*model.MedicalRecord, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	recordType, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.RecordType)), []string{"DIAGNOSIS", "ALLERGY", "CHRONIC_CONDITION", "INJURY", "SURGERY", "CLINICAL_NOTE"})
	if err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"ACTIVE", "RESOLVED"})
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.CreateMedicalRecord(ctx, repository.CreateMedicalRecordInput{ID: uuid.New(), PetID: p.PetID, RecordType: recordType, Status: status, Title: title, Description: trimOrNil(p.Description), StartedAt: p.StartedAt, ResolvedAt: p.ResolvedAt, CreatedBy: p.UserID, UpdatedBy: p.UserID, Attachments: attachments})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "MEDICAL_RECORD", item.ID, sync); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) UpdateMedicalRecord(ctx context.Context, p UpdateMedicalRecordParams) (*model.MedicalRecord, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RecordID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	recordType, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.RecordType)), []string{"DIAGNOSIS", "ALLERGY", "CHRONIC_CONDITION", "INJURY", "SURGERY", "CLINICAL_NOTE"})
	if err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"ACTIVE", "RESOLVED"})
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	attachments, err := s.prepareHealthAttachments(ctx, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	item, sync, err := s.repo.UpdateMedicalRecord(ctx, repository.UpdateMedicalRecordInput{ID: p.RecordID, PetID: p.PetID, RowVersion: p.RowVersion, RecordType: recordType, Status: status, Title: title, Description: trimOrNil(p.Description), StartedAt: p.StartedAt, ResolvedAt: p.ResolvedAt, UpdatedBy: p.UserID, Attachments: attachments})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.syncHealthAttachments(ctx, p.PetID, "MEDICAL_RECORD", item.ID, sync); err != nil {
		return nil, err
	}
	s.enrichHealthAttachmentURLs(ctx, item.Attachments)
	return item, nil
}

func (s *Service) DeleteMedicalRecord(ctx context.Context, p DeleteMedicalRecordParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RecordID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := s.repo.GetMedicalRecord(ctx, p.PetID, p.RecordID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.repo.DeleteMedicalRecord(ctx, repository.DeleteMedicalRecordInput{ID: p.RecordID, PetID: p.PetID, RowVersion: p.RowVersion, DeletedBy: p.UserID}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "MEDICAL_RECORD", p.RecordID, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetHealthDay(ctx context.Context, userID, petID uuid.UUID, day time.Time) ([]model.CalendarDayItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || day.IsZero() {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
	items, err := s.repo.ListCalendarDayItems(ctx, petID, dayStart, dayEnd)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].ScheduledFor.Equal(items[j].ScheduledFor) {
			return items[i].ScheduledFor.Before(items[j].ScheduledFor)
		}
		if items[i].ItemType != items[j].ItemType {
			return items[i].ItemType < items[j].ItemType
		}
		return items[i].EntityID.String() < items[j].EntityID.String()
	})
	return items, nil
}

func (s *Service) requireHealthRead(ctx context.Context, petID, userID uuid.UUID) (bool, error) {
	allowed, err := s.acl.Check(ctx, petID, userID, ActionHealthRead)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, ErrForbidden
	}
	return true, nil
}

func (s *Service) requireHealthWrite(ctx context.Context, petID, userID uuid.UUID) error {
	allowed, err := s.acl.Check(ctx, petID, userID, ActionHealthWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

func (s *Service) prepareHealthAttachments(ctx context.Context, fileIDs []uuid.UUID) ([]repository.AttachmentInput, error) {
	ids := uniqueUUIDs(fileIDs)
	if len(ids) == 0 {
		return []repository.AttachmentInput{}, nil
	}
	files, err := s.file.GetFiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(files) != len(ids) {
		return nil, ErrInvalidInput
	}
	attachments := make([]repository.AttachmentInput, 0, len(ids))
	for i := range ids {
		file, ok := files[ids[i]]
		if !ok {
			return nil, ErrInvalidInput
		}
		attachments = append(attachments, repository.AttachmentInput{
			FileID:   ids[i],
			FileName: file.FileName,
			FileType: detectAttachmentFileType(file.MimeType),
		})
	}
	return attachments, nil
}

func (s *Service) syncHealthAttachments(ctx context.Context, petID uuid.UUID, entityType string, entityID uuid.UUID, sync repository.AttachmentSync) error {
	if len(sync.Add) > 0 {
		if err := s.file.LinkAttachments(ctx, petID, entityType, entityID, sync.Add); err != nil {
			return err
		}
	}
	if len(sync.Remove) > 0 {
		if err := s.file.UnlinkAttachments(ctx, entityType, entityID, sync.Remove); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) enrichHealthAttachmentURLs(ctx context.Context, attachments []model.HealthAttachment) {
	if len(attachments) == 0 {
		return
	}
	fileIDs := make([]uuid.UUID, 0, len(attachments))
	for i := range attachments {
		fileIDs = append(fileIDs, attachments[i].FileID)
	}
	urls, err := s.file.BatchGetDownloadURLs(ctx, fileIDs)
	if err != nil {
		return
	}
	for i := range attachments {
		if url, ok := urls[attachments[i].FileID]; ok && strings.TrimSpace(url) != "" {
			urlCopy := url
			attachments[i].DownloadURL = &urlCopy
			if attachments[i].FileType == "image" {
				attachments[i].PreviewURL = &urlCopy
			}
		}
	}
}

func (s *Service) ensureVisitLinkValid(ctx context.Context, petID uuid.UUID, visitID *uuid.UUID) error {
	if visitID == nil {
		return nil
	}
	if *visitID == uuid.Nil {
		return ErrInvalidInput
	}
	_, err := s.repo.GetVetVisit(ctx, petID, *visitID, false)
	if err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Service) syncVaccinationAutolog(ctx context.Context, item *model.Vaccination, userID uuid.UUID) error {
	if item == nil || item.Status != "DONE" {
		return nil
	}
	description := vaccinationAutologDescription(item)
	return mapRepoErr(s.repo.UpsertHealthEntityLog(ctx, repository.UpsertHealthEntityLogInput{
		PetID:           item.PetID,
		EntityType:      "VACCINATION",
		EntityID:        item.ID,
		OccurredAt:      vaccinationAutologOccurredAt(item),
		Description:     &description,
		CreatedByUserID: userID,
		UpdatedByUserID: userID,
	}))
}

func (s *Service) syncVaccinationAutologTransition(ctx context.Context, previous *model.Vaccination, current *model.Vaccination, userID uuid.UUID) error {
	if current == nil {
		return nil
	}
	if current.Status == "DONE" {
		return s.syncVaccinationAutolog(ctx, current, userID)
	}
	if previous != nil && previous.Status == "DONE" {
		return mapRepoErr(s.repo.DeleteHealthEntityLog(ctx, repository.DeleteHealthEntityLogInput{
			PetID:           current.PetID,
			EntityType:      "VACCINATION",
			EntityID:        current.ID,
			DeletedByUserID: userID,
		}))
	}
	return nil
}

func (s *Service) syncProcedureAutolog(ctx context.Context, item *model.Procedure, userID uuid.UUID) error {
	if item == nil || item.Status != "DONE" {
		return nil
	}
	description := procedureAutologDescription(item)
	return mapRepoErr(s.repo.UpsertHealthEntityLog(ctx, repository.UpsertHealthEntityLogInput{
		PetID:           item.PetID,
		EntityType:      "PROCEDURE",
		EntityID:        item.ID,
		OccurredAt:      procedureAutologOccurredAt(item),
		Description:     &description,
		CreatedByUserID: userID,
		UpdatedByUserID: userID,
	}))
}

func (s *Service) syncProcedureAutologTransition(ctx context.Context, previous *model.Procedure, current *model.Procedure, userID uuid.UUID) error {
	if current == nil {
		return nil
	}
	if current.Status == "DONE" {
		return s.syncProcedureAutolog(ctx, current, userID)
	}
	if previous != nil && previous.Status == "DONE" {
		return mapRepoErr(s.repo.DeleteHealthEntityLog(ctx, repository.DeleteHealthEntityLogInput{
			PetID:           current.PetID,
			EntityType:      "PROCEDURE",
			EntityID:        current.ID,
			DeletedByUserID: userID,
		}))
	}
	return nil
}

func vaccinationAutologOccurredAt(item *model.Vaccination) time.Time {
	if item == nil {
		return time.Time{}
	}
	if item.AdministeredAt != nil {
		return *item.AdministeredAt
	}
	if item.ScheduledAt != nil {
		return *item.ScheduledAt
	}
	return item.UpdatedAt
}

func procedureAutologOccurredAt(item *model.Procedure) time.Time {
	if item == nil {
		return time.Time{}
	}
	if item.PerformedAt != nil {
		return *item.PerformedAt
	}
	if item.ScheduledAt != nil {
		return *item.ScheduledAt
	}
	return item.UpdatedAt
}

func vaccinationAutologDescription(item *model.Vaccination) string {
	name := ""
	if item != nil {
		name = strings.TrimSpace(item.VaccineName)
	}
	if name == "" {
		return "Проведена вакцинация"
	}
	return "Проведена вакцинация: " + name
}

func procedureAutologDescription(item *model.Procedure) string {
	title := ""
	if item != nil {
		title = strings.TrimSpace(item.Title)
	}
	if title == "" {
		return "Проведена процедура"
	}
	return "Проведена процедура: " + title
}

func (s *Service) ensureNextVaccination(ctx context.Context, petID uuid.UUID, previousStatus string, current *model.Vaccination, userID uuid.UUID) error {
	if current == nil || current.NextDueAt == nil || current.Status != "DONE" || previousStatus == "DONE" {
		return nil
	}
	plannedStatus := "PLANNED"
	_, _, err := s.repo.CreateVaccination(ctx, repository.CreateVaccinationInput{
		ID:                  uuid.New(),
		PetID:               petID,
		Status:              plannedStatus,
		VaccineName:         current.VaccineName,
		CatalogMedicationID: current.CatalogMedicationID,
		ScheduledAt:         current.NextDueAt,
		AdministeredAt:      nil,
		NextDueAt:           nil,
		VetVisitID:          nil,
		ClinicName:          nil,
		VetName:             nil,
		Notes:               nil,
		CreatedBy:           userID,
		UpdatedBy:           userID,
		Attachments:         []repository.AttachmentInput{},
	})
	return mapRepoErr(err)
}

func (s *Service) ensureNextProcedure(ctx context.Context, petID uuid.UUID, previousStatus string, current *model.Procedure, userID uuid.UUID) error {
	if current == nil || current.NextDueAt == nil || current.Status != "DONE" || previousStatus == "DONE" {
		return nil
	}
	plannedStatus := "PLANNED"
	_, _, err := s.repo.CreateProcedure(ctx, repository.CreateProcedureInput{
		ID:                  uuid.New(),
		PetID:               petID,
		Status:              plannedStatus,
		ProcedureType:       current.ProcedureType,
		Title:               current.Title,
		Description:         current.Description,
		CatalogMedicationID: current.CatalogMedicationID,
		ProductName:         current.ProductName,
		ScheduledAt:         current.NextDueAt,
		PerformedAt:         nil,
		NextDueAt:           nil,
		VetVisitID:          nil,
		Notes:               nil,
		CreatedBy:           userID,
		UpdatedBy:           userID,
		Attachments:         []repository.AttachmentInput{},
	})
	return mapRepoErr(err)
}

func validateVetVisitState(statusRaw, typeRaw string) (string, string, error) {
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(statusRaw)), []string{"PLANNED", "COMPLETED", "CANCELLED"})
	if err != nil {
		return "", "", err
	}
	visitType, err := validateEnum(strings.TrimSpace(strings.ToUpper(typeRaw)), []string{"CHECKUP", "SYMPTOM", "FOLLOW_UP", "VACCINATION", "PROCEDURE", "OTHER"})
	if err != nil {
		return "", "", err
	}
	return status, visitType, nil
}

func validateEnum(value string, allowed []string) (string, error) {
	if value == "" {
		return "", ErrInvalidInput
	}
	for i := range allowed {
		if value == allowed[i] {
			return value, nil
		}
	}
	return "", ErrInvalidInput
}

func normalizeEnumPtr(raw string, allowed []string) *string {
	trimmed := strings.TrimSpace(strings.ToUpper(raw))
	if trimmed == "" {
		return nil
	}
	for i := range allowed {
		if trimmed == allowed[i] {
			return &trimmed
		}
	}
	return nil
}

func normalizeBucket(raw string, allowed []string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	for i := range allowed {
		if trimmed == allowed[i] {
			return trimmed
		}
	}
	return "INVALID"
}

func detectAttachmentFileType(mimeType string) string {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case mimeType == "application/pdf":
		return "pdf"
	default:
		return "other"
	}
}
