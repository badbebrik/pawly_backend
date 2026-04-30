package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VetVisits struct {
	repo      ports.VetVisitRepository
	acl       ports.HealthAccessChecker
	file      ports.HealthFileClient
	scheduled ports.ScheduledRepository
}

func NewVetVisits(repo ports.VetVisitRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient, scheduled ports.ScheduledRepository) *VetVisits {
	return &VetVisits{repo: repo, acl: acl, file: file, scheduled: scheduled}
}

type ListVetVisitsParams struct {
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

type ListVetVisitsResult struct {
	Items      []model.VetVisitListItem
	NextCursor *ports.TimeCursor
}

type CreateVetVisitParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	Status      string
	VisitType   string
	Title       *string
	ScheduledAt *time.Time
	Reminder    *MedicalEntityReminderParams
	CompletedAt *time.Time
	ReasonText  *string
	ResultText  *string
	ClinicName  *string
	VetName     *string
	Attachments []AttachmentParam
}

type UpdateVetVisitParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	VisitID     uuid.UUID
	RowVersion  int
	Status      string
	VisitType   string
	Title       *string
	ScheduledAt *time.Time
	Reminder    *MedicalEntityReminderParams
	CompletedAt *time.Time
	ReasonText  *string
	ResultText  *string
	ClinicName  *string
	VetName     *string
	Attachments []AttachmentParam
}

type DeleteVetVisitParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	VisitID    uuid.UUID
	RowVersion int
}

func (u *VetVisits) ListVetVisits(ctx context.Context, p ListVetVisitsParams) (ListVetVisitsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListVetVisitsResult{}, ErrInvalidInput
	}
	canRead, err := requireHealthRead(ctx, u.acl, p.PetID, p.UserID)
	if err != nil || !canRead {
		return ListVetVisitsResult{}, err
	}
	canLogRead, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return ListVetVisitsResult{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"PLANNED", "COMPLETED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return ListVetVisitsResult{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "upcoming", "history", "all"})
	if bucket == "INVALID" {
		return ListVetVisitsResult{}, ErrInvalidInput
	}
	out, err := u.repo.ListVetVisits(ctx, ports.ListVetVisitsQuery{
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
		return ListVetVisitsResult{}, mapRepoErr(err)
	}
	if !canLogRead {
		for i := range out.Items {
			out.Items[i].RelatedLogsCount = 0
		}
	}
	return ListVetVisitsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *VetVisits) GetVetVisit(ctx context.Context, userID, petID, visitID uuid.UUID) (*model.VetVisit, error) {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	canRead, err := requireHealthRead(ctx, u.acl, petID, userID)
	if err != nil || !canRead {
		return nil, err
	}
	canLogRead, err := u.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	item, err := u.repo.GetVetVisit(ctx, petID, visitID, canLogRead)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	if !canLogRead {
		item.RelatedLogs = []model.RelatedLog{}
	}
	return item, nil
}

func (u *VetVisits) CreateVetVisit(ctx context.Context, p CreateVetVisitParams) (*model.VetVisit, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, visitType, err := validateVetVisitState(p.Status, p.VisitType)
	if err != nil {
		return nil, err
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.CreateVetVisit(ctx, ports.CreateVetVisitInput{
		ID:          uuid.New(),
		PetID:       p.PetID,
		Status:      status,
		VisitType:   visitType,
		Title:       trimStringOrNil(p.Title),
		ScheduledAt: p.ScheduledAt,
		CompletedAt: p.CompletedAt,
		ReasonText:  trimStringOrNil(p.ReasonText),
		ResultText:  trimStringOrNil(p.ResultText),
		ClinicName:  trimStringOrNil(p.ClinicName),
		VetName:     trimStringOrNil(p.VetName),
		CreatedBy:   p.UserID,
		UpdatedBy:   p.UserID,
		Attachments: attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "VET_VISIT", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeVetVisit, item.ID, vetVisitScheduledTitle(item), vetVisitScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeVetVisit, item.ID, item.Status == "PLANNED" && item.ScheduledAt != nil, p.UserID); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *VetVisits) UpdateVetVisit(ctx context.Context, p UpdateVetVisitParams) (*model.VetVisit, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VisitID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	status, visitType, err := validateVetVisitState(p.Status, p.VisitType)
	if err != nil {
		return nil, err
	}
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.UpdateVetVisit(ctx, ports.UpdateVetVisitInput{
		ID:          p.VisitID,
		PetID:       p.PetID,
		RowVersion:  p.RowVersion,
		Status:      status,
		VisitType:   visitType,
		Title:       trimStringOrNil(p.Title),
		ScheduledAt: p.ScheduledAt,
		CompletedAt: p.CompletedAt,
		ReasonText:  trimStringOrNil(p.ReasonText),
		ResultText:  trimStringOrNil(p.ResultText),
		ClinicName:  trimStringOrNil(p.ClinicName),
		VetName:     trimStringOrNil(p.VetName),
		UpdatedBy:   p.UserID,
		Attachments: attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "VET_VISIT", item.ID, sync); err != nil {
		return nil, err
	}
	if _, err := syncSystemOneShotScheduledItem(ctx, u.scheduled, p.PetID, model.ScheduledItemSourceTypeVetVisit, item.ID, vetVisitScheduledTitle(item), vetVisitScheduledNote(item), item.ScheduledAt, item.Status == "PLANNED", p.UserID); err != nil {
		return nil, err
	}
	if err := applyMedicalEntityReminderSettings(ctx, u.scheduled, p.PetID, p.Reminder, model.ScheduledItemSourceTypeVetVisit, item.ID, item.Status == "PLANNED" && item.ScheduledAt != nil, p.UserID); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *VetVisits) DeleteVetVisit(ctx context.Context, p DeleteVetVisitParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.VisitID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := u.repo.GetVetVisit(ctx, p.PetID, p.VisitID, false)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := u.repo.DeleteVetVisit(ctx, ports.DeleteVetVisitInput{ID: p.VisitID, PetID: p.PetID, RowVersion: p.RowVersion, DeletedBy: p.UserID}); err != nil {
		return mapRepoErr(err)
	}
	if err := u.scheduled.DeleteHealthScheduledItem(ctx, ports.DeleteHealthScheduledItemInput{
		PetID:           p.PetID,
		SourceType:      model.ScheduledItemSourceTypeVetVisit,
		SourceID:        p.VisitID,
		DeletedByUserID: p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "VET_VISIT", p.VisitID, fileIDs); err != nil {
			return err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (u *VetVisits) LinkVetVisitLog(ctx context.Context, userID, petID, visitID, logID uuid.UUID) (*model.RelatedLog, error) {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil || logID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, petID, userID); err != nil {
		return nil, err
	}
	allowed, err := u.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	item, err := u.repo.LinkVetVisitLog(ctx, petID, visitID, logID, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (u *VetVisits) UnlinkVetVisitLog(ctx context.Context, userID, petID, visitID, logID uuid.UUID) error {
	if userID == uuid.Nil || petID == uuid.Nil || visitID == uuid.Nil || logID == uuid.Nil {
		return ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, petID, userID); err != nil {
		return err
	}
	allowed, err := u.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return mapRepoErr(u.repo.UnlinkVetVisitLog(ctx, petID, visitID, logID))
}

func requireHealthRead(ctx context.Context, acl ports.HealthAccessChecker, petID, userID uuid.UUID) (bool, error) {
	allowed, err := acl.Check(ctx, petID, userID, ActionHealthRead)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, ErrForbidden
	}
	return true, nil
}

func requireHealthWrite(ctx context.Context, acl ports.HealthAccessChecker, petID, userID uuid.UUID) error {
	allowed, err := acl.Check(ctx, petID, userID, ActionHealthWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

func vetVisitScheduledNote(item *model.VetVisit) *string {
	if item == nil {
		return nil
	}
	parts := make([]string, 0, 2)
	if item.ClinicName != nil && strings.TrimSpace(*item.ClinicName) != "" {
		parts = append(parts, strings.TrimSpace(*item.ClinicName))
	}
	if item.VetName != nil && strings.TrimSpace(*item.VetName) != "" {
		parts = append(parts, strings.TrimSpace(*item.VetName))
	}
	if len(parts) == 0 {
		return nil
	}
	note := strings.Join(parts, ", ")
	return &note
}

func vetVisitScheduledTitle(item *model.VetVisit) string {
	if item != nil && item.Title != nil && strings.TrimSpace(*item.Title) != "" {
		return strings.TrimSpace(*item.Title)
	}
	return "Прием у ветеринара"
}

func validateVetVisitState(statusRaw, typeRaw string) (string, string, error) {
	status, err := validateEnum(strings.TrimSpace(strings.ToUpper(statusRaw)), []string{"PLANNED", "COMPLETED"})
	if err != nil {
		return "", "", err
	}
	visitType, err := validateEnum(strings.TrimSpace(strings.ToUpper(typeRaw)), []string{"CHECKUP", "SYMPTOM", "FOLLOW_UP", "VACCINATION", "PROCEDURE", "OTHER"})
	if err != nil {
		return "", "", err
	}
	return status, visitType, nil
}

func normalizeBucket(raw string, allowed []string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	for i := range allowed {
		if value == allowed[i] {
			return value
		}
	}
	return "INVALID"
}

func healthAttachmentFileIDs(items []model.HealthAttachment) []uuid.UUID {
	if len(items) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(items))
	for i := range items {
		out = append(out, items[i].FileID)
	}
	return out
}
