package usecase

import (
	"context"
	"errors"
	"profile/internal/application/ports"
	"profile/internal/config"
	"profile/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

func baseProfileDeps() (*Set, *stubProfileRepo, *stubProfileFileClient) {
	repo := &stubProfileRepo{profiles: map[uuid.UUID]model.Profile{}}
	files := &stubProfileFileClient{
		initFileID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		upload: ports.UploadInfo{
			Method:    "PUT",
			URL:       "https://files.example/upload",
			Headers:   map[string]string{"Content-Type": "image/jpeg"},
			ExpiresAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		downloadURL: "https://files.example/avatar.jpg",
		batchURLs:   map[uuid.UUID]string{},
	}
	set := NewSet(Dependencies{
		Profiles:   repo,
		FileClient: files,
		Config: &config.Config{
			DefaultLocale:   "ru",
			DefaultTimezone: "UTC",
		},
	})
	return set, repo, files
}

func expectProfileErr(t *testing.T, got error, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("expected error %v, got %v", want, got)
	}
}

func profileStringPtr(v string) *string {
	return &v
}

func profileInt64Ptr(v int64) *int64 {
	return &v
}

func profileUUIDPtr(v uuid.UUID) *uuid.UUID {
	return &v
}

type stubProfileRepo struct {
	profiles map[uuid.UUID]model.Profile

	createErr  error
	updateErr  error
	deleteErr  error
	getErr     error
	getManyErr error

	created      []model.Profile
	updated      []model.Profile
	deleted      []uuid.UUID
	getManyInput []uuid.UUID
}

func (r *stubProfileRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*model.Profile, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	p, ok := r.profiles[userID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return cloneProfile(p), nil
}

func (r *stubProfileRepo) GetByUserIDs(_ context.Context, userIDs []uuid.UUID) ([]model.Profile, error) {
	if r.getManyErr != nil {
		return nil, r.getManyErr
	}
	r.getManyInput = append([]uuid.UUID(nil), userIDs...)
	out := make([]model.Profile, 0, len(userIDs))
	for _, id := range userIDs {
		if p, ok := r.profiles[id]; ok {
			out = append(out, *cloneProfile(p))
		}
	}
	return out, nil
}

func (r *stubProfileRepo) Create(_ context.Context, p *model.Profile) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.profiles[p.UserID] = *cloneProfile(*p)
	r.created = append(r.created, *cloneProfile(*p))
	return nil
}

func (r *stubProfileRepo) Update(_ context.Context, p *model.Profile) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.profiles[p.UserID] = *cloneProfile(*p)
	r.updated = append(r.updated, *cloneProfile(*p))
	return nil
}

func (r *stubProfileRepo) Delete(_ context.Context, userID uuid.UUID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.profiles, userID)
	r.deleted = append(r.deleted, userID)
	return nil
}

func cloneProfile(in model.Profile) *model.Profile {
	out := in
	if in.FirstName != nil {
		v := *in.FirstName
		out.FirstName = &v
	}
	if in.LastName != nil {
		v := *in.LastName
		out.LastName = &v
	}
	if in.AvatarFileID != nil {
		v := *in.AvatarFileID
		out.AvatarFileID = &v
	}
	return &out
}

type stubProfileFileClient struct {
	initFileID uuid.UUID
	upload     ports.UploadInfo
	initErr    error
	initCalls  []profileInitUploadCall

	confirmErr   error
	confirmCalls []profileConfirmUploadCall

	downloadURL string
	downloadErr error

	batchURLs   map[uuid.UUID]string
	batchErr    error
	batchInputs [][]uuid.UUID

	linkErr   error
	linkCalls []profileAvatarLinkCall
	unlinkErr error
	unlinks   []profileAvatarLinkCall
	deleteErr error
	deletions []uuid.UUID
}

type profileInitUploadCall struct {
	mimeType     string
	expectedSize int64
	userID       uuid.UUID
}

type profileConfirmUploadCall struct {
	fileID    uuid.UUID
	sizeBytes int64
}

type profileAvatarLinkCall struct {
	fileID uuid.UUID
	userID uuid.UUID
}

func (c *stubProfileFileClient) InitUpload(_ context.Context, mimeType string, expectedSize int64, userID uuid.UUID) (uuid.UUID, ports.UploadInfo, error) {
	c.initCalls = append(c.initCalls, profileInitUploadCall{mimeType: mimeType, expectedSize: expectedSize, userID: userID})
	if c.initErr != nil {
		return uuid.Nil, ports.UploadInfo{}, c.initErr
	}
	return c.initFileID, c.upload, nil
}

func (c *stubProfileFileClient) ConfirmUpload(_ context.Context, fileID uuid.UUID, sizeBytes int64) error {
	c.confirmCalls = append(c.confirmCalls, profileConfirmUploadCall{fileID: fileID, sizeBytes: sizeBytes})
	return c.confirmErr
}

func (c *stubProfileFileClient) GetDownloadURL(_ context.Context, _ uuid.UUID) (string, time.Time, error) {
	if c.downloadErr != nil {
		return "", time.Time{}, c.downloadErr
	}
	return c.downloadURL, time.Now().Add(time.Hour), nil
}

func (c *stubProfileFileClient) BatchGetDownloadURLs(_ context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	c.batchInputs = append(c.batchInputs, append([]uuid.UUID(nil), fileIDs...))
	if c.batchErr != nil {
		return nil, c.batchErr
	}
	out := make(map[uuid.UUID]string, len(c.batchURLs))
	for id, url := range c.batchURLs {
		out[id] = url
	}
	return out, nil
}

func (c *stubProfileFileClient) LinkAvatar(_ context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	c.linkCalls = append(c.linkCalls, profileAvatarLinkCall{fileID: fileID, userID: userID})
	return c.linkErr
}

func (c *stubProfileFileClient) UnlinkAvatar(_ context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	c.unlinks = append(c.unlinks, profileAvatarLinkCall{fileID: fileID, userID: userID})
	return c.unlinkErr
}

func (c *stubProfileFileClient) DeleteFileIfUnlinked(_ context.Context, fileID uuid.UUID) error {
	c.deletions = append(c.deletions, fileID)
	return c.deleteErr
}

var _ ports.ProfileRepository = (*stubProfileRepo)(nil)
var _ ports.FileClient = (*stubProfileFileClient)(nil)
