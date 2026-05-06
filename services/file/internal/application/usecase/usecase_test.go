package usecase

import (
	"context"
	"errors"
	"file/internal/application/ports"
	"file/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubFileObjectRepo struct {
	createFn               func(context.Context, *model.FileObject) error
	getByIDFn              func(context.Context, uuid.UUID) (*model.FileObject, error)
	listExpiredUploadingFn func(context.Context, time.Time, int) ([]model.FileObject, error)
	listPendingDeleteFn    func(context.Context, int) ([]model.FileObject, error)
	updateStatusFn         func(context.Context, uuid.UUID, model.FileStatus) error
	confirmUploadFn        func(context.Context, uuid.UUID, int64) error
	markPendingDeleteFn    func(context.Context, uuid.UUID) error
	markDeletedFn          func(context.Context, uuid.UUID) error
	deleteByIDFn           func(context.Context, uuid.UUID) error
}

func (s *stubFileObjectRepo) Create(ctx context.Context, f *model.FileObject) error {
	if s.createFn != nil {
		return s.createFn(ctx, f)
	}
	return nil
}

func (s *stubFileObjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FileObject, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, ports.ErrNotFound
}

func (s *stubFileObjectRepo) ListExpiredUploading(ctx context.Context, now time.Time, limit int) ([]model.FileObject, error) {
	if s.listExpiredUploadingFn != nil {
		return s.listExpiredUploadingFn(ctx, now, limit)
	}
	return []model.FileObject{}, nil
}

func (s *stubFileObjectRepo) ListPendingDelete(ctx context.Context, limit int) ([]model.FileObject, error) {
	if s.listPendingDeleteFn != nil {
		return s.listPendingDeleteFn(ctx, limit)
	}
	return []model.FileObject{}, nil
}

func (s *stubFileObjectRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus) error {
	if s.updateStatusFn != nil {
		return s.updateStatusFn(ctx, id, status)
	}
	return nil
}

func (s *stubFileObjectRepo) ConfirmUpload(ctx context.Context, id uuid.UUID, sizeBytes int64) error {
	if s.confirmUploadFn != nil {
		return s.confirmUploadFn(ctx, id, sizeBytes)
	}
	return nil
}

func (s *stubFileObjectRepo) MarkPendingDelete(ctx context.Context, id uuid.UUID) error {
	if s.markPendingDeleteFn != nil {
		return s.markPendingDeleteFn(ctx, id)
	}
	return nil
}

func (s *stubFileObjectRepo) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	if s.markDeletedFn != nil {
		return s.markDeletedFn(ctx, id)
	}
	return nil
}

func (s *stubFileObjectRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if s.deleteByIDFn != nil {
		return s.deleteByIDFn(ctx, id)
	}
	return nil
}

type stubFileLinkRepo struct {
	createFn         func(context.Context, *model.FileLink) error
	listByFileIDFn   func(context.Context, uuid.UUID) ([]model.FileLink, error)
	countByFileIDFn  func(context.Context, uuid.UUID) (int64, error)
	deleteFn         func(context.Context, uuid.UUID, model.FileUsageType, uuid.UUID) (bool, error)
	deleteByFileIDFn func(context.Context, uuid.UUID) (int64, error)
}

func (s *stubFileLinkRepo) Create(ctx context.Context, l *model.FileLink) error {
	if s.createFn != nil {
		return s.createFn(ctx, l)
	}
	return nil
}

func (s *stubFileLinkRepo) ListByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error) {
	if s.listByFileIDFn != nil {
		return s.listByFileIDFn(ctx, fileID)
	}
	return []model.FileLink{}, nil
}

func (s *stubFileLinkRepo) CountByFileID(ctx context.Context, fileID uuid.UUID) (int64, error) {
	if s.countByFileIDFn != nil {
		return s.countByFileIDFn(ctx, fileID)
	}
	return 0, nil
}

func (s *stubFileLinkRepo) Delete(ctx context.Context, fileID uuid.UUID, usageType model.FileUsageType, ownerID uuid.UUID) (bool, error) {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, fileID, usageType, ownerID)
	}
	return true, nil
}

func (s *stubFileLinkRepo) DeleteByFileID(ctx context.Context, fileID uuid.UUID) (int64, error) {
	if s.deleteByFileIDFn != nil {
		return s.deleteByFileIDFn(ctx, fileID)
	}
	return 0, nil
}

type stubStorage struct {
	presignPutFn   func(context.Context, string, string, string, time.Duration) (string, error)
	presignGetFn   func(context.Context, string, string, time.Duration) (string, error)
	statObjectFn   func(context.Context, string, string) (model.ObjectInfo, error)
	deleteObjectFn func(context.Context, string, string) error
}

func (s *stubStorage) Bucket() string {
	return "files"
}

func (s *stubStorage) PresignPut(ctx context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error) {
	if s.presignPutFn != nil {
		return s.presignPutFn(ctx, bucket, objectKey, contentType, expires)
	}
	return "put-url", nil
}

func (s *stubStorage) PresignGet(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error) {
	if s.presignGetFn != nil {
		return s.presignGetFn(ctx, bucket, objectKey, expires)
	}
	return "get-url", nil
}

func (s *stubStorage) StatObject(ctx context.Context, bucket, objectKey string) (model.ObjectInfo, error) {
	if s.statObjectFn != nil {
		return s.statObjectFn(ctx, bucket, objectKey)
	}
	return model.ObjectInfo{Size: 100, ContentType: "image/png"}, nil
}

func (s *stubStorage) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	if s.deleteObjectFn != nil {
		return s.deleteObjectFn(ctx, bucket, objectKey)
	}
	return nil
}

func (s *stubStorage) UploadTTL() time.Duration {
	return time.Hour
}

func (s *stubStorage) DownloadTTL() time.Duration {
	return 10 * time.Minute
}

func newTestSet(objects ports.FileObjectRepository, links ports.FileLinkRepository, storage ports.Storage) *Set {
	return NewSet(Dependencies{Objects: objects, Links: links, Storage: storage})
}

func expectFileErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v want %v", got, want)
	}
}

func readyFile(id uuid.UUID) *model.FileObject {
	return &model.FileObject{
		ID:              id,
		StorageBucket:   "files",
		StorageKey:      id.String(),
		Status:          model.FileStatusReady,
		MimeType:        "image/png",
		UploadExpiresAt: time.Now().Add(time.Hour),
	}
}

func uploadingFile(id uuid.UUID, size *int64) *model.FileObject {
	return &model.FileObject{
		ID:              id,
		StorageBucket:   "files",
		StorageKey:      id.String(),
		Status:          model.FileStatusUploading,
		MimeType:        "image/png",
		SizeBytes:       size,
		UploadExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestInitUpload_CreatesUploadingFileAndPresign(t *testing.T) {
	var created *model.FileObject
	var presignContentType string
	fileName := " file.png "
	size := int64(100)

	set := newTestSet(&stubFileObjectRepo{
		createFn: func(_ context.Context, f *model.FileObject) error {
			created = f
			return nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{
		presignPutFn: func(_ context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error) {
			if bucket != "files" || objectKey == "" || expires != time.Hour {
				t.Fatalf("unexpected presign put args: bucket=%s key=%s expires=%s", bucket, objectKey, expires)
			}
			presignContentType = contentType
			return "put-url", nil
		},
	})

	out, upload, err := set.InitUpload(context.Background(), InitUploadParams{
		MimeType:          "image/png",
		ExpectedSizeBytes: &size,
		OriginalFilename:  &fileName,
	})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if out.ID == uuid.Nil || created == nil || created.Status != model.FileStatusUploading {
		t.Fatalf("unexpected created file: %+v", created)
	}
	if created.OriginalFilename == nil || *created.OriginalFilename != "file.png" {
		t.Fatalf("unexpected file name: %v", created.OriginalFilename)
	}
	if presignContentType != "image/png" || upload.Method != "PUT" || upload.URL != "put-url" {
		t.Fatalf("unexpected upload info: %+v", upload)
	}
}

func TestInitUploadRejectsEmptyMimeType(t *testing.T) {
	set := newTestSet(&stubFileObjectRepo{}, &stubFileLinkRepo{}, &stubStorage{})

	_, _, err := set.InitUpload(context.Background(), InitUploadParams{})
	expectFileErr(t, err, ErrInvalidInput)
}

func TestConfirmUpload_ConfirmsStorageObject(t *testing.T) {
	fileID := uuid.New()
	expectedSize := int64(100)
	var confirmedID uuid.UUID
	var confirmedSize int64

	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return uploadingFile(fileID, &expectedSize), nil
		},
		confirmUploadFn: func(_ context.Context, id uuid.UUID, sizeBytes int64) error {
			confirmedID = id
			confirmedSize = sizeBytes
			return nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{
		statObjectFn: func(context.Context, string, string) (model.ObjectInfo, error) {
			return model.ObjectInfo{Size: 100, ContentType: "image/jpeg"}, nil
		},
	})

	out, err := set.ConfirmUpload(context.Background(), ConfirmUploadParams{FileID: fileID, SizeBytes: 100})
	if err != nil {
		t.Fatalf("ConfirmUpload returned error: %v", err)
	}
	if confirmedID != fileID || confirmedSize != 100 {
		t.Fatalf("unexpected confirm args: id=%s size=%d", confirmedID, confirmedSize)
	}
	if out.Status != model.FileStatusReady || out.SizeBytes == nil || *out.SizeBytes != 100 || out.MimeType != "image/jpeg" {
		t.Fatalf("unexpected confirmed file: %+v", out)
	}
}

func TestConfirmUploadRejectsInvalidStatesAndSizes(t *testing.T) {
	fileID := uuid.New()

	t.Run("not uploading", func(t *testing.T) {
		set := newTestSet(&stubFileObjectRepo{
			getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
				return readyFile(fileID), nil
			},
		}, &stubFileLinkRepo{}, &stubStorage{})

		_, err := set.ConfirmUpload(context.Background(), ConfirmUploadParams{FileID: fileID, SizeBytes: 100})
		expectFileErr(t, err, ErrInvalidState)
	})

	t.Run("expected size mismatch", func(t *testing.T) {
		expectedSize := int64(100)
		set := newTestSet(&stubFileObjectRepo{
			getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
				return uploadingFile(fileID, &expectedSize), nil
			},
		}, &stubFileLinkRepo{}, &stubStorage{})

		_, err := set.ConfirmUpload(context.Background(), ConfirmUploadParams{FileID: fileID, SizeBytes: 90})
		expectFileErr(t, err, ErrInvalidInput)
	})

	t.Run("storage size mismatch", func(t *testing.T) {
		set := newTestSet(&stubFileObjectRepo{
			getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
				return uploadingFile(fileID, nil), nil
			},
		}, &stubFileLinkRepo{}, &stubStorage{
			statObjectFn: func(context.Context, string, string) (model.ObjectInfo, error) {
				return model.ObjectInfo{Size: 90, ContentType: "image/png"}, nil
			},
		})

		_, err := set.ConfirmUpload(context.Background(), ConfirmUploadParams{FileID: fileID, SizeBytes: 100})
		expectFileErr(t, err, ErrInvalidInput)
	})
}

func TestConfirmUploadRejectsExpiredUpload(t *testing.T) {
	fileID := uuid.New()
	file := uploadingFile(fileID, nil)
	file.UploadExpiresAt = time.Now().Add(-time.Minute)

	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return file, nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{})

	_, err := set.ConfirmUpload(context.Background(), ConfirmUploadParams{FileID: fileID, SizeBytes: 100})
	expectFileErr(t, err, ErrUploadExpired)
}

func TestGetDownloadURL_RequiresReadyFile(t *testing.T) {
	fileID := uuid.New()
	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return readyFile(fileID), nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{
		presignGetFn: func(_ context.Context, bucket, objectKey string, expires time.Duration) (string, error) {
			if bucket != "files" || objectKey != fileID.String() || expires != 10*time.Minute {
				t.Fatalf("unexpected presign get args: bucket=%s key=%s expires=%s", bucket, objectKey, expires)
			}
			return "get-url", nil
		},
	})

	url, expiresAt, err := set.GetDownloadURL(context.Background(), fileID)
	if err != nil {
		t.Fatalf("GetDownloadURL returned error: %v", err)
	}
	if url != "get-url" || expiresAt.IsZero() {
		t.Fatalf("unexpected download result: url=%s expires=%s", url, expiresAt)
	}

	notReadySet := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return uploadingFile(fileID, nil), nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{})
	_, _, err = notReadySet.GetDownloadURL(context.Background(), fileID)
	expectFileErr(t, err, ErrNotReady)
}

func TestLinkRequiresReadyFile(t *testing.T) {
	fileID := uuid.New()
	ownerID := uuid.New()
	var createdLink *model.FileLink

	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return readyFile(fileID), nil
		},
	}, &stubFileLinkRepo{
		createFn: func(_ context.Context, l *model.FileLink) error {
			createdLink = l
			return nil
		},
	}, &stubStorage{})

	out, err := set.Link(context.Background(), LinkParams{
		FileID: fileID, UsageType: model.FileUsageTypePetAvatar, OwnerID: ownerID,
	})
	if err != nil {
		t.Fatalf("Link returned error: %v", err)
	}
	if out.ID == uuid.Nil || createdLink == nil || createdLink.FileID != fileID || createdLink.OwnerID != ownerID {
		t.Fatalf("unexpected link: %+v", createdLink)
	}

	notReadySet := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return uploadingFile(fileID, nil), nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{})
	_, err = notReadySet.Link(context.Background(), LinkParams{FileID: fileID, UsageType: model.FileUsageTypePetAvatar, OwnerID: ownerID})
	expectFileErr(t, err, ErrNotReady)
}

func TestUnlinkDelegatesToLinks(t *testing.T) {
	fileID := uuid.New()
	ownerID := uuid.New()
	set := newTestSet(&stubFileObjectRepo{}, &stubFileLinkRepo{
		deleteFn: func(_ context.Context, gotFileID uuid.UUID, usageType model.FileUsageType, gotOwnerID uuid.UUID) (bool, error) {
			if gotFileID != fileID || usageType != model.FileUsageTypePetAvatar || gotOwnerID != ownerID {
				t.Fatalf("unexpected unlink args: file=%s usage=%s owner=%s", gotFileID, usageType, gotOwnerID)
			}
			return true, nil
		},
	}, &stubStorage{})

	deleted, err := set.Unlink(context.Background(), fileID, model.FileUsageTypePetAvatar, ownerID)
	if err != nil {
		t.Fatalf("Unlink returned error: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted=true")
	}
}

func TestDeleteFileIfUnlinkedDeletesReadyFile(t *testing.T) {
	fileID := uuid.New()
	var deletedObjectKey string
	var markedDeletedID uuid.UUID

	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return readyFile(fileID), nil
		},
		markDeletedFn: func(_ context.Context, id uuid.UUID) error {
			markedDeletedID = id
			return nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{
		deleteObjectFn: func(_ context.Context, bucket, objectKey string) error {
			if bucket != "files" {
				t.Fatalf("unexpected bucket: %s", bucket)
			}
			deletedObjectKey = objectKey
			return nil
		},
	})

	out, err := set.DeleteFileIfUnlinked(context.Background(), fileID)
	if err != nil {
		t.Fatalf("DeleteFileIfUnlinked returned error: %v", err)
	}
	if deletedObjectKey != fileID.String() || markedDeletedID != fileID {
		t.Fatalf("unexpected delete state: key=%s id=%s", deletedObjectKey, markedDeletedID)
	}
	if out.Status != model.FileStatusDeleted || out.DeletedAt == nil {
		t.Fatalf("unexpected deleted file: %+v", out)
	}
}

func TestDeleteFileIfUnlinkedRejectsLinkedOrInvalidState(t *testing.T) {
	fileID := uuid.New()

	t.Run("has links", func(t *testing.T) {
		set := newTestSet(&stubFileObjectRepo{
			getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
				return readyFile(fileID), nil
			},
		}, &stubFileLinkRepo{
			countByFileIDFn: func(context.Context, uuid.UUID) (int64, error) {
				return 1, nil
			},
		}, &stubStorage{})

		_, err := set.DeleteFileIfUnlinked(context.Background(), fileID)
		expectFileErr(t, err, ErrHasLinks)
	})

	t.Run("uploading", func(t *testing.T) {
		set := newTestSet(&stubFileObjectRepo{
			getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
				return uploadingFile(fileID, nil), nil
			},
		}, &stubFileLinkRepo{}, &stubStorage{})

		_, err := set.DeleteFileIfUnlinked(context.Background(), fileID)
		expectFileErr(t, err, ErrInvalidState)
	})
}

func TestDeleteFileIfUnlinkedMarksPendingDeleteOnStorageFailure(t *testing.T) {
	fileID := uuid.New()
	var markedPendingID uuid.UUID

	set := newTestSet(&stubFileObjectRepo{
		getByIDFn: func(context.Context, uuid.UUID) (*model.FileObject, error) {
			return readyFile(fileID), nil
		},
		markPendingDeleteFn: func(_ context.Context, id uuid.UUID) error {
			markedPendingID = id
			return nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{
		deleteObjectFn: func(context.Context, string, string) error {
			return errors.New("storage error")
		},
	})

	out, err := set.DeleteFileIfUnlinked(context.Background(), fileID)
	if err != nil {
		t.Fatalf("DeleteFileIfUnlinked returned error: %v", err)
	}
	if markedPendingID != fileID || out.Status != model.FileStatusPendingDelete {
		t.Fatalf("unexpected pending delete result: id=%s file=%+v", markedPendingID, out)
	}
}

func TestRunCleanupBatchDeletesExpiredAndPendingFiles(t *testing.T) {
	expiredID := uuid.New()
	pendingID := uuid.New()
	var deletedIDs []uuid.UUID
	var markedDeletedIDs []uuid.UUID

	set := newTestSet(&stubFileObjectRepo{
		listExpiredUploadingFn: func(_ context.Context, _ time.Time, limit int) ([]model.FileObject, error) {
			if limit != 100 {
				t.Fatalf("unexpected default limit: %d", limit)
			}
			return []model.FileObject{*uploadingFile(expiredID, nil)}, nil
		},
		listPendingDeleteFn: func(_ context.Context, limit int) ([]model.FileObject, error) {
			if limit != 100 {
				t.Fatalf("unexpected default limit: %d", limit)
			}
			file := readyFile(pendingID)
			file.Status = model.FileStatusPendingDelete
			return []model.FileObject{*file}, nil
		},
		deleteByIDFn: func(_ context.Context, id uuid.UUID) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
		markDeletedFn: func(_ context.Context, id uuid.UUID) error {
			markedDeletedIDs = append(markedDeletedIDs, id)
			return nil
		},
	}, &stubFileLinkRepo{}, &stubStorage{})

	out, err := set.RunCleanupBatch(context.Background(), 0)
	if err != nil {
		t.Fatalf("RunCleanupBatch returned error: %v", err)
	}
	if out.DeletedExpired != 1 || out.DeletedPending != 1 || out.MarkedPendingDelete != 0 {
		t.Fatalf("unexpected cleanup result: %+v", out)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != expiredID {
		t.Fatalf("unexpected expired deletes: %v", deletedIDs)
	}
	if len(markedDeletedIDs) != 1 || markedDeletedIDs[0] != pendingID {
		t.Fatalf("unexpected pending deletes: %v", markedDeletedIDs)
	}
}

func TestRunCleanupBatchSkipsLinkedPendingAndCountsDeleteFailures(t *testing.T) {
	pendingLinkedID := uuid.New()
	pendingFailID := uuid.New()

	set := newTestSet(&stubFileObjectRepo{
		listPendingDeleteFn: func(context.Context, int) ([]model.FileObject, error) {
			linked := readyFile(pendingLinkedID)
			linked.Status = model.FileStatusPendingDelete
			failing := readyFile(pendingFailID)
			failing.Status = model.FileStatusPendingDelete
			return []model.FileObject{*linked, *failing}, nil
		},
	}, &stubFileLinkRepo{
		countByFileIDFn: func(_ context.Context, id uuid.UUID) (int64, error) {
			if id == pendingLinkedID {
				return 1, nil
			}
			return 0, nil
		},
	}, &stubStorage{
		deleteObjectFn: func(_ context.Context, _ string, objectKey string) error {
			if objectKey == pendingFailID.String() {
				return errors.New("storage error")
			}
			return nil
		},
	})

	out, err := set.RunCleanupBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunCleanupBatch returned error: %v", err)
	}
	if out.DeletedPending != 0 || out.MarkedPendingDelete != 1 {
		t.Fatalf("unexpected cleanup result: %+v", out)
	}
}

var (
	_ ports.FileObjectRepository = (*stubFileObjectRepo)(nil)
	_ ports.FileLinkRepository   = (*stubFileLinkRepo)(nil)
	_ ports.Storage              = (*stubStorage)(nil)
)
