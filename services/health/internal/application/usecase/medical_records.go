package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MedicalRecords struct {
	repo       ports.MedicalRecordRepository
	dictionary ports.HealthDictionaryRepository
	acl        ports.HealthAccessChecker
	file       ports.HealthFileClient
}

func NewMedicalRecords(repo ports.MedicalRecordRepository, dictionary ports.HealthDictionaryRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient) *MedicalRecords {
	return &MedicalRecords{repo: repo, dictionary: dictionary, acl: acl, file: file}
}

type ListMedicalRecordsParams struct {
	UserID       uuid.UUID
	PetID        uuid.UUID
	Cursor       *ports.TimeCursor
	Limit        int
	Q            string
	Status       string
	Bucket       string
	RecordTypeID *uuid.UUID
	Sort         string
}

type ListMedicalRecordsResult struct {
	Items      []model.MedicalRecordListItem
	NextCursor *ports.TimeCursor
}

type CreateMedicalRecordParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	RecordTypeID   *uuid.UUID
	RecordTypeName *string
	Status         string
	Title          string
	Description    *string
	StartedAt      *time.Time
	ResolvedAt     *time.Time
	Attachments    []AttachmentParam
}

type UpdateMedicalRecordParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int
	RecordTypeID   *uuid.UUID
	RecordTypeName *string
	Status         string
	Title          string
	Description    *string
	StartedAt      *time.Time
	ResolvedAt     *time.Time
	Attachments    []AttachmentParam
}

type DeleteMedicalRecordParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	RecordID   uuid.UUID
	RowVersion int
}

func (u *MedicalRecords) ListMedicalRecords(ctx context.Context, p ListMedicalRecordsParams) (ListMedicalRecordsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListMedicalRecordsResult{}, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return ListMedicalRecordsResult{}, err
	}
	status := normalizeEnumPtr(p.Status, []string{"ACTIVE", "RESOLVED"})
	if strings.TrimSpace(p.Status) != "" && status == nil {
		return ListMedicalRecordsResult{}, ErrInvalidInput
	}
	bucket := normalizeBucket(p.Bucket, []string{"", "active", "archive", "all"})
	if bucket == "INVALID" {
		return ListMedicalRecordsResult{}, ErrInvalidInput
	}
	out, err := u.repo.ListMedicalRecords(ctx, ports.ListMedicalRecordsQuery{
		PetID:            p.PetID,
		Cursor:           p.Cursor,
		Limit:            p.Limit,
		Q:                strings.TrimSpace(p.Q),
		Status:           status,
		Bucket:           bucket,
		RecordTypeItemID: p.RecordTypeID,
		Sort:             p.Sort,
	})
	if err != nil {
		return ListMedicalRecordsResult{}, mapRepoErr(err)
	}
	return ListMedicalRecordsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *MedicalRecords) GetMedicalRecord(ctx context.Context, userID, petID, recordID uuid.UUID) (*model.MedicalRecord, error) {
	if userID == uuid.Nil || petID == uuid.Nil || recordID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, petID, userID); err != nil {
		return nil, err
	}
	item, err := u.repo.GetMedicalRecord(ctx, petID, recordID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *MedicalRecords) CreateMedicalRecord(ctx context.Context, p CreateMedicalRecordParams) (*model.MedicalRecord, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	recordTypeItem, err := resolveDictionaryItem(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindMedicalRecordType, p.RecordTypeID, p.RecordTypeName)
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
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.CreateMedicalRecord(ctx, ports.CreateMedicalRecordInput{
		ID:               uuid.New(),
		PetID:            p.PetID,
		RecordTypeItemID: dictionaryItemIDPtr(recordTypeItem),
		Status:           status,
		Title:            title,
		Description:      trimStringOrNil(p.Description),
		StartedAt:        p.StartedAt,
		ResolvedAt:       p.ResolvedAt,
		CreatedBy:        p.UserID,
		UpdatedBy:        p.UserID,
		Attachments:      attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "MEDICAL_RECORD", item.ID, sync); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *MedicalRecords) UpdateMedicalRecord(ctx context.Context, p UpdateMedicalRecordParams) (*model.MedicalRecord, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RecordID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	recordTypeItem, err := resolveDictionaryItem(ctx, u.dictionary, p.PetID, p.UserID, ports.HealthDictionaryKindMedicalRecordType, p.RecordTypeID, p.RecordTypeName)
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
	attachments, err := prepareHealthAttachments(ctx, u.file, p.Attachments)
	if err != nil {
		return nil, err
	}
	item, sync, err := u.repo.UpdateMedicalRecord(ctx, ports.UpdateMedicalRecordInput{
		ID:               p.RecordID,
		PetID:            p.PetID,
		RowVersion:       p.RowVersion,
		RecordTypeItemID: dictionaryItemIDPtr(recordTypeItem),
		Status:           status,
		Title:            title,
		Description:      trimStringOrNil(p.Description),
		StartedAt:        p.StartedAt,
		ResolvedAt:       p.ResolvedAt,
		UpdatedBy:        p.UserID,
		Attachments:      attachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := syncHealthAttachments(ctx, u.file, p.PetID, "MEDICAL_RECORD", item.ID, sync); err != nil {
		return nil, err
	}
	enrichHealthAttachmentURLs(ctx, u.file, item.Attachments)
	return item, nil
}

func (u *MedicalRecords) DeleteMedicalRecord(ctx context.Context, p DeleteMedicalRecordParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RecordID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := u.repo.GetMedicalRecord(ctx, p.PetID, p.RecordID)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := u.repo.DeleteMedicalRecord(ctx, ports.DeleteMedicalRecordInput{
		ID:         p.RecordID,
		PetID:      p.PetID,
		RowVersion: p.RowVersion,
		DeletedBy:  p.UserID,
	}); err != nil {
		return mapRepoErr(err)
	}
	fileIDs := healthAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "MEDICAL_RECORD", p.RecordID, fileIDs); err != nil {
			return err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, fileIDs); err != nil {
			return err
		}
	}
	return nil
}
