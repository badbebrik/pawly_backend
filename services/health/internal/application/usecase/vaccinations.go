package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Vaccinations struct {
	repo       ports.VaccinationRepository
	dictionary ports.HealthDictionaryRepository
	vetVisits  ports.VetVisitRepository
	acl        ports.HealthAccessChecker
	file       ports.HealthFileClient
	scheduled  ports.ScheduledRepository
}

func NewVaccinations(repo ports.VaccinationRepository, dictionary ports.HealthDictionaryRepository, vetVisits ports.VetVisitRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient, scheduled ports.ScheduledRepository) *Vaccinations {
	return &Vaccinations{repo: repo, dictionary: dictionary, vetVisits: vetVisits, acl: acl, file: file, scheduled: scheduled}
}

type ListVaccinationsParams struct {
	UserID   uuid.UUID
	PetID    uuid.UUID
	Cursor   *ports.TimeCursor
	Limit    int
	Q        string
	Status   string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type ListVaccinationsResult struct {
	Items      []model.VaccinationListItem
	NextCursor *ports.TimeCursor
}

type CreateVaccinationParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	Targets             []HealthDictionaryItemRefParam
	ScheduledAt         *time.Time
	Reminder            *MedicalEntityReminderParams
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	Notes               *string
	Attachments         []AttachmentParam
}

type UpdateVaccinationParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	VaccinationID       uuid.UUID
	RowVersion          int
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	Targets             []HealthDictionaryItemRefParam
	ScheduledAt         *time.Time
	Reminder            *MedicalEntityReminderParams
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	Notes               *string
	Attachments         []AttachmentParam
}

type DeleteVaccinationParams struct {
	UserID        uuid.UUID
	PetID         uuid.UUID
	VaccinationID uuid.UUID
	RowVersion    int
}

func (u *Vaccinations) ListVaccinations(ctx context.Context, p ListVaccinationsParams) (ListVaccinationsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListVaccinationsResult{}, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return ListVaccinationsResult{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "COMPLETED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return ListVaccinationsResult{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "upcoming", "history", "all"})
	if bucket == "INVALID" {
		return ListVaccinationsResult{}, ErrInvalidInput
	}
	out, err := u.repo.ListVaccinations(ctx, ports.ListVaccinationsQuery{
		PetID:    p.PetID,
		Cursor:   p.Cursor,
		Limit:    p.Limit,
		Q:        strings.TrimSpace(p.Q),
		Status:   status,
		Bucket:   bucket,
		DateFrom: p.DateFrom,
		DateTo:   p.DateTo,
		Sort:     p.Sort,
	})
	if err != nil {
		return ListVaccinationsResult{}, mapRepoErr(err)
	}
	return ListVaccinationsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Vaccinations) GetVaccination(ctx context.Context, userID, petID, vaccinationID uuid.UUID) (*model.Vaccination, error) {
	if userID == uuid.Nil || petID == uuid.Nil || vaccinationID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, petID, userID); err != nil {
		return nil, err
	}
	item, err := u.repo.GetVaccination(ctx, petID, vaccinationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Vaccinations) CreateVaccination(ctx context.Context, p CreateVaccinationParams) (*model.Vaccination, error) {
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
	vaccineName := strings.TrimSpace(p.VaccineName)
	if vaccineName == "" {
		return nil, ErrInvalidInput
	}
	if err := u.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	targetIDs, err := resolveDictionaryItemRefs(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindVaccinationTarget, p.Targets)
	if err != nil {
		return nil, err
	}
	if targetIDs == nil {
		targetIDs = []uuid.UUID{}
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.CreateVaccination(ctx, ports.CreateVaccinationInput{
		ID:                  uuid.New(),
		PetID:               p.PetID,
		Status:              status,
		VaccineName:         vaccineName,
		CatalogMedicationID: p.CatalogMedicationID,
		TargetItemIDs:       targetIDs,
		ScheduledAt:         p.ScheduledAt,
		AdministeredAt:      p.AdministeredAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		ClinicName:          trimStringOrNil(p.ClinicName),
		VetName:             trimStringOrNil(p.VetName),
		Notes:               trimStringOrNil(p.Notes),
		CreatedBy:           p.UserID,
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "VACCINATION", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeVaccination, item.ID, item.VaccineName, vaccinationScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if item.Status == "PLANNED" && p.Reminder != nil {
		if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeVaccination, item.ID, item.ScheduledAt != nil, p.UserID); err != nil {
			return nil, err
		}
	}
	if err := u.syncGeneratedNextVaccination(ctx, p.PetID, item, p.UserID, p.Reminder); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Vaccinations) UpdateVaccination(ctx context.Context, p UpdateVaccinationParams) (*model.Vaccination, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VaccinationID == uuid.Nil || p.RowVersion <= 0 {
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
	vaccineName := strings.TrimSpace(p.VaccineName)
	if vaccineName == "" {
		return nil, ErrInvalidInput
	}
	if err := u.ensureVisitLinkValid(ctx, p.PetID, p.VetVisitID); err != nil {
		return nil, err
	}
	targetIDs, err := resolveDictionaryItemRefs(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindVaccinationTarget, p.Targets)
	if err != nil {
		return nil, err
	}
	inheritedReminder, err := getMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeVaccination, p.VaccinationID)
	if err != nil {
		return nil, err
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.UpdateVaccination(ctx, ports.UpdateVaccinationInput{
		ID:                  p.VaccinationID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		Status:              status,
		VaccineName:         vaccineName,
		CatalogMedicationID: p.CatalogMedicationID,
		TargetItemIDs:       targetIDs,
		ScheduledAt:         p.ScheduledAt,
		AdministeredAt:      p.AdministeredAt,
		NextDueAt:           p.NextDueAt,
		VetVisitID:          p.VetVisitID,
		ClinicName:          trimStringOrNil(p.ClinicName),
		VetName:             trimStringOrNil(p.VetName),
		Notes:               trimStringOrNil(p.Notes),
		UpdatedBy:           p.UserID,
		Attachments:         attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "VACCINATION", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeVaccination, item.ID, item.VaccineName, vaccinationScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if item.Status == "PLANNED" && p.Reminder != nil {
		if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeVaccination, item.ID, item.ScheduledAt != nil, p.UserID); err != nil {
			return nil, err
		}
	}
	if err := u.syncGeneratedNextVaccination(ctx, p.PetID, item, p.UserID, generatedPlanReminder(p.Reminder, inheritedReminder)); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *Vaccinations) DeleteVaccination(ctx context.Context, p DeleteVaccinationParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VaccinationID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := u.repo.GetVaccination(ctx, p.PetID, p.VaccinationID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := u.repo.DeleteVaccination(ctx, ports.DeleteVaccinationInput{
		ID:         p.VaccinationID,
		PetID:      p.PetID,
		RowVersion: p.RowVersion,
		DeletedBy:  p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.scheduled.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
		PetID:           p.PetID,
		SourceType:      model.ScheduledItemSourceTypeVaccination,
		SourceID:        p.VaccinationID,
		DeletedByUserID: p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.deleteGeneratedVaccinationPlan(ctx, p.PetID, p.VaccinationID, p.UserID); err != nil {
		return err
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "VACCINATION", p.VaccinationID, fileIDs); err != nil {
			return err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (u *Vaccinations) ensureVisitLinkValid(ctx context.Context, petID uuid.UUID, visitID *uuid.UUID) error {
	if visitID == nil {
		return nil
	}
	if *visitID == uuid.Nil {
		return ErrInvalidInput
	}
	_, err := u.vetVisits.GetVetVisit(ctx, petID, *visitID, false)
	return mapRepoErr(err)
}

func (u *Vaccinations) syncGeneratedNextVaccination(ctx context.Context, petID uuid.UUID, current *model.Vaccination, userID uuid.UUID, reminder *MedicalEntityReminderParams) error {
	if current == nil {
		return ErrInvalidInput
	}
	child, err := u.repo.GetGeneratedVaccination(ctx, petID, current.ID)
	if err != nil && err != ports.ErrNotFound {
		return mapRepoErr(err)
	}
	if current.Status != "COMPLETED" || current.NextDueAt == nil {
		if child != nil && child.Status == "PLANNED" {
			return u.deleteVaccinationPlan(ctx, petID, child, userID)
		}
		return nil
	}
	if child != nil && child.Status != "PLANNED" {
		return nil
	}

	var planned *model.Vaccination
	plannedStatus := "PLANNED"
	if child == nil {
		generatedFromID := current.ID
		planned, _, err = u.repo.CreateVaccination(ctx, ports.CreateVaccinationInput{
			ID:                  uuid.New(),
			PetID:               petID,
			GeneratedFromID:     &generatedFromID,
			Status:              plannedStatus,
			VaccineName:         current.VaccineName,
			CatalogMedicationID: current.CatalogMedicationID,
			TargetItemIDs:       vaccinationTargetIDs(current.Targets),
			ScheduledAt:         current.NextDueAt,
			AdministeredAt:      nil,
			NextDueAt:           nil,
			VetVisitID:          nil,
			ClinicName:          nil,
			VetName:             nil,
			Notes:               nil,
			CreatedBy:           userID,
			UpdatedBy:           userID,
			Attachments:         []ports.AttachmentInput{},
		})
	} else {
		planned, err = u.repo.UpdateGeneratedVaccinationPlan(ctx, ports.UpdateGeneratedVaccinationPlanInput{
			ID:                  child.ID,
			PetID:               petID,
			VaccineName:         current.VaccineName,
			CatalogMedicationID: current.CatalogMedicationID,
			TargetItemIDs:       vaccinationTargetIDs(current.Targets),
			ScheduledAt:         current.NextDueAt,
			UpdatedBy:           userID,
		})
	}
	if err != nil {
		return mapRepoErr(err)
	}
	_, err = syncSystemOneShotScheduledItem(ctx, u.scheduled, petID, model.ScheduledItemSourceTypeVaccination, planned.ID, planned.VaccineName, vaccinationScheduledNote(planned), planned.ScheduledAt, planned.Status == "PLANNED", userID)
	if err != nil {
		return err
	}
	if reminder == nil {
		return nil
	}
	return applyMedicalEntityReminderSettings(ctx, u.scheduled, petID, reminder, model.ScheduledItemSourceTypeVaccination, planned.ID, true, userID)
}

func (u *Vaccinations) deleteGeneratedVaccinationPlan(ctx context.Context, petID, generatedFromID, userID uuid.UUID) error {
	child, err := u.repo.GetGeneratedVaccination(ctx, petID, generatedFromID)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil
		}
		return mapRepoErr(err)
	}
	if child.Status != "PLANNED" {
		return nil
	}
	return u.deleteVaccinationPlan(ctx, petID, child, userID)
}

func (u *Vaccinations) deleteVaccinationPlan(ctx context.Context, petID uuid.UUID, item *model.Vaccination, userID uuid.UUID) error {
	if item == nil {
		return nil
	}
	if err := u.repo.DeleteVaccination(ctx, ports.DeleteVaccinationInput{
		ID:         item.ID,
		PetID:      petID,
		RowVersion: item.RowVersion,
		DeletedBy:  userID,
	}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.scheduled.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
		PetID:           petID,
		SourceType:      model.ScheduledItemSourceTypeVaccination,
		SourceID:        item.ID,
		DeletedByUserID: userID,
	}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(item.Attachments)
	if len(fileIDs) == 0 {
		return nil
	}
	if err := u.file.UnlinkAttachments(ctx, "VACCINATION", item.ID, fileIDs); err != nil {
		return err
	}
	return u.file.DeleteFilesIfUnlinked(ctx, fileIDs)
}

func vaccinationScheduledNote(item *model.Vaccination) *string {
	if item == nil {
		return nil
	}
	note := "Вакцинация"
	return &note
}

func vaccinationTargetIDs(items []model.HealthDictionaryItem) []uuid.UUID {
	if len(items) == 0 {
		return []uuid.UUID{}
	}
	out := make([]uuid.UUID, 0, len(items))
	for i := range items {
		out = append(out, items[i].ID)
	}
	return out
}
