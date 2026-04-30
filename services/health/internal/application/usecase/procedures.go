package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Procedures struct {
	repo       ports.ProcedureRepository
	dictionary ports.HealthDictionaryRepository
	vetVisits  ports.VetVisitRepository
	acl        ports.HealthAccessChecker
	file       ports.HealthFileClient
	scheduled  ports.ScheduledRepository
}

func NewProcedures(repo ports.ProcedureRepository, dictionary ports.HealthDictionaryRepository, vetVisits ports.VetVisitRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient, scheduled ports.ScheduledRepository) *Procedures {
	return &Procedures{repo: repo, dictionary: dictionary, vetVisits: vetVisits, acl: acl, file: file, scheduled: scheduled}
}

type ListProceduresParams struct {
	UserID          uuid.UUID
	PetID           uuid.UUID
	Cursor          *ports.TimeCursor
	Limit           int
	Q               string
	Status          string
	Bucket          string
	ProcedureTypeID *uuid.UUID
	DateFrom        *time.Time
	DateTo          *time.Time
	Sort            string
}

type ListProceduresResult struct {
	Items      []model.ProcedureListItem
	NextCursor *ports.TimeCursor
}

type CreateProcedureParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	Status              string
	ProcedureTypeID     *uuid.UUID
	ProcedureTypeName   *string
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	Reminder            *MedicalEntityReminderParams
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	Notes               *string
	Attachments         []AttachmentParam
}

type UpdateProcedureParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	ProcedureID         uuid.UUID
	RowVersion          int
	Status              string
	ProcedureTypeID     *uuid.UUID
	ProcedureTypeName   *string
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	Reminder            *MedicalEntityReminderParams
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	Notes               *string
	Attachments         []AttachmentParam
}

type DeleteProcedureParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	ProcedureID uuid.UUID
	RowVersion  int
}

func (u *Procedures) ListProcedures(ctx context.Context, p ListProceduresParams) (ListProceduresResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListProceduresResult{}, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return ListProceduresResult{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "COMPLETED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return ListProceduresResult{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "planned", "history", "all"})
	if bucket == "INVALID" {
		return ListProceduresResult{}, ErrInvalidInput
	}
	out, err := u.repo.ListProcedures(ctx, ports.ListProceduresQuery{
		PetID:               p.PetID,
		Cursor:              p.Cursor,
		Limit:               p.Limit,
		Q:                   strings.TrimSpace(p.Q),
		Status:              status,
		Bucket:              bucket,
		ProcedureTypeItemID: p.ProcedureTypeID,
		DateFrom:            p.DateFrom,
		DateTo:              p.DateTo,
		Sort:                p.Sort,
	})
	if err != nil {
		return ListProceduresResult{}, mapRepoErr(err)
	}
	return ListProceduresResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Procedures) GetProcedure(ctx context.Context, userID, petID, procedureID uuid.UUID) (*model.Procedure, error) {
	if userID == uuid.Nil || petID == uuid.Nil || procedureID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, petID, userID); err != nil {
		return nil, err
	}
	item, err := u.repo.GetProcedure(ctx, petID, procedureID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Procedures) CreateProcedure(ctx context.Context, p CreateProcedureParams) (*model.Procedure, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "COMPLETED"})
	if err != nil {
		return nil, err
	}
	if err := validateMedicalEntityReminderParams(p.Reminder); err != nil {
		return nil, err
	}
	if status == "PLANNED" && p.Reminder != nil && p.ScheduledAt == nil {
		return nil, ErrInvalidInput
	}
	if status == "COMPLETED" && p.Reminder != nil && p.NextDueAt == nil {
		return nil, ErrInvalidInput
	}
	procedureTypeItem, err := resolveDictionaryItem(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindProcedureType, p.ProcedureTypeID, p.ProcedureTypeName)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if err := ensureMedicalVisitLinkValid(ctx, u.vetVisits, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.CreateProcedure(ctx, ports.CreateProcedureInput{
		ID:                  uuid.New(),
		PetID:               p.PetID,
		Status:              status,
		ProcedureTypeItemID: dictionaryItemIDPtr(procedureTypeItem),
		Title:               title,
		Description:         trimStringOrNil(p.Description),
		CatalogMedicationID: p.CatalogMedicationID,
		ProductName:         trimStringOrNil(p.ProductName),
		ScheduledAt:         p.ScheduledAt,
		PerformedAt:         p.PerformedAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		Notes:               trimStringOrNil(p.Notes),
		CreatedBy:           p.UserID,
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "PROCEDURE", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeProcedure, item.ID, item.Title, procedureScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if item.Status == "PLANNED" && p.Reminder != nil {
		if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeProcedure, item.ID, item.ScheduledAt != nil, p.UserID); err != nil {
			return nil, err
		}
	}
	if err := u.syncGeneratedNextProcedure(ctx, p.PetID, item, p.UserID, p.Reminder); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Procedures) UpdateProcedure(ctx context.Context, p UpdateProcedureParams) (*model.Procedure, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ProcedureID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(p.Status)), []string{"PLANNED", "COMPLETED"})
	if err != nil {
		return nil, err
	}
	if err := validateMedicalEntityReminderParams(p.Reminder); err != nil {
		return nil, err
	}
	if status == "PLANNED" && p.Reminder != nil && p.ScheduledAt == nil {
		return nil, ErrInvalidInput
	}
	if status == "COMPLETED" && p.Reminder != nil && p.NextDueAt == nil {
		return nil, ErrInvalidInput
	}
	procedureTypeItem, err := resolveDictionaryItem(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindProcedureType, p.ProcedureTypeID, p.ProcedureTypeName)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if err := ensureMedicalVisitLinkValid(ctx, u.vetVisits, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	inheritedReminder, err := getMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeProcedure, p.ProcedureID)
	if err != nil {
		return nil, err
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.UpdateProcedure(ctx, ports.UpdateProcedureInput{
		ID:                  p.ProcedureID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		Status:              status,
		ProcedureTypeItemID: dictionaryItemIDPtr(procedureTypeItem),
		Title:               title,
		Description:         trimStringOrNil(p.Description),
		CatalogMedicationID: p.CatalogMedicationID,
		ProductName:         trimStringOrNil(p.ProductName),
		ScheduledAt:         p.ScheduledAt,
		PerformedAt:         p.PerformedAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		Notes:               trimStringOrNil(p.Notes),
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "PROCEDURE", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeProcedure, item.ID, item.Title, procedureScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if item.Status == "PLANNED" && p.Reminder != nil {
		if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeProcedure, item.ID, item.ScheduledAt != nil, p.UserID); err != nil {
			return nil, err
		}
	}
	if err := u.syncGeneratedNextProcedure(ctx, p.PetID, item, p.UserID, generatedPlanReminder(p.Reminder, inheritedReminder)); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Procedures) DeleteProcedure(ctx context.Context, p DeleteProcedureParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ProcedureID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := u.repo.GetProcedure(ctx, p.PetID, p.ProcedureID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := u.repo.DeleteProcedure(ctx, ports.DeleteProcedureInput{
		ID:         p.ProcedureID,
		PetID:      p.PetID,
		RowVersion: p.RowVersion,
		DeletedBy:  p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.scheduled.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
		PetID:           p.PetID,
		SourceType:      model.ScheduledItemSourceTypeProcedure,
		SourceID:        p.ProcedureID,
		DeletedByUserID: p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.deleteGeneratedProcedurePlan(ctx, p.PetID, p.ProcedureID, p.UserID); err != nil {
		return err
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "PROCEDURE", p.ProcedureID, fileIDs); err != nil {
			return err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (u *Procedures) syncGeneratedNextProcedure(ctx context.Context, petID uuid.UUID, current *model.Procedure, userID uuid.UUID, reminder *MedicalEntityReminderParams) error {
	if current == nil {
		return ErrInvalidInput
	}
	child, err := u.repo.GetGeneratedProcedure(ctx, petID, current.ID)
	if err != nil && err != ports.ErrNotFound {
		return mapRepoErr(err)
	}
	if current.Status != "COMPLETED" || current.NextDueAt == nil {
		if child != nil && child.Status == "PLANNED" {
			return u.deleteProcedurePlan(ctx, petID, child, userID)
		}
		return nil
	}
	if child != nil && child.Status != "PLANNED" {
		return nil
	}

	var planned *model.Procedure
	plannedStatus := "PLANNED"
	if child == nil {
		generatedFromID := current.ID
		planned, _, err = u.repo.CreateProcedure(ctx, ports.CreateProcedureInput{
			ID:                  uuid.New(),
			PetID:               petID,
			GeneratedFromID:     &generatedFromID,
			Status:              plannedStatus,
			ProcedureTypeItemID: dictionaryItemIDPtr(current.ProcedureTypeItem),
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
			Attachments:         []ports.AttachmentInput{},
		})
	} else {
		planned, err = u.repo.UpdateGeneratedProcedurePlan(ctx, ports.UpdateGeneratedProcedurePlanInput{
			ID:                  child.ID,
			PetID:               petID,
			ProcedureTypeItemID: dictionaryItemIDPtr(current.ProcedureTypeItem),
			Title:               current.Title,
			Description:         current.Description,
			CatalogMedicationID: current.CatalogMedicationID,
			ProductName:         current.ProductName,
			ScheduledAt:         current.NextDueAt,
			UpdatedBy:           userID,
		})
	}
	if err != nil {
		return mapRepoErr(err)
	}
	_, err = syncSystemOneShotScheduledItem(ctx, u.scheduled, petID, model.ScheduledItemSourceTypeProcedure, planned.ID, planned.Title, procedureScheduledNote(planned), planned.ScheduledAt, planned.Status == "PLANNED", userID)
	if err != nil {
		return err
	}
	if reminder == nil {
		return nil
	}
	return applyMedicalEntityReminderSettings(ctx, u.scheduled, petID, reminder, model.ScheduledItemSourceTypeProcedure, planned.ID, true, userID)
}

func (u *Procedures) deleteGeneratedProcedurePlan(ctx context.Context, petID, generatedFromID, userID uuid.UUID) error {
	child, err := u.repo.GetGeneratedProcedure(ctx, petID, generatedFromID)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil
		}
		return mapRepoErr(err)
	}
	if child.Status != "PLANNED" {
		return nil
	}
	return u.deleteProcedurePlan(ctx, petID, child, userID)
}

func (u *Procedures) deleteProcedurePlan(ctx context.Context, petID uuid.UUID, item *model.Procedure, userID uuid.UUID) error {
	if item == nil {
		return nil
	}
	if err := u.repo.DeleteProcedure(ctx, ports.DeleteProcedureInput{
		ID:         item.ID,
		PetID:      petID,
		RowVersion: item.RowVersion,
		DeletedBy:  userID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.scheduled.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
		PetID:           petID,
		SourceType:      model.ScheduledItemSourceTypeProcedure,
		SourceID:        item.ID,
		DeletedByUserID: userID,
	}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(item.Attachments)
	if len(fileIDs) == 0 {
		return nil
	}
	if err := u.file.UnlinkAttachments(ctx, "PROCEDURE", item.ID, fileIDs); err != nil {
		return err
	}
	return u.file.DeleteFilesIfUnlinked(ctx, fileIDs)
}

func procedureScheduledNote(item *model.Procedure) *string {
	if item == nil {
		return nil
	}
	if item.ProductName != nil && strings.TrimSpace(*item.ProductName) != "" {
		product := strings.TrimSpace(*item.ProductName)
		return &product
	}
	return nil
}

func ensureMedicalVisitLinkValid(ctx context.Context, visits ports.VetVisitRepository, petID uuid.UUID, visitID *uuid.UUID) error {
	if visitID == nil {
		return nil
	}
	if *visitID == uuid.Nil {
		return ErrInvalidInput
	}
	_, err := visits.GetVetVisit(ctx, petID, *visitID, false)
	return mapRepoErr(err)
}
