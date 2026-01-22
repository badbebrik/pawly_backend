package handlers

import (
	"encoding/json"
	"gateway/internal/grpc"
	"gateway/internal/transport/http/middleware"
	"net/http"
	"strings"
	"time"

	filepb "file/proto"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileHandlers struct {
	client *grpc.FileClient
}

func NewFileHandlers(client *grpc.FileClient) *FileHandlers {
	return &FileHandlers{client: client}
}

type initUploadRequest struct {
	MimeType          string `json:"mime_type"`
	ExpectedSizeBytes *int64 `json:"expected_size_bytes"`
}

type uploadDTO struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

type fileObjectDTO struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
	CreatedBy string `json:"created_by_user_id"`
	CreatedAt string `json:"created_at"`
}

type initUploadResponse struct {
	File   fileObjectDTO `json:"file"`
	Upload uploadDTO     `json:"upload"`
}

type confirmUploadRequest struct {
	SizeBytes int64 `json:"size_bytes"`
}

type confirmUploadResponse struct {
	File fileObjectDTO `json:"file"`
}

type downloadURLResponse struct {
	FileID    string `json:"file_id"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

type batchDownloadRequest struct {
	FileIDs []string `json:"file_ids"`
}

type batchDownloadResponse struct {
	Items []downloadURLResponse `json:"items"`
}

type linkFileRequest struct {
	FileID       string `json:"file_id"`
	OwnerService string `json:"owner_service"`
	OwnerType    string `json:"owner_type"`
	OwnerID      string `json:"owner_id"`
	PetID        string `json:"pet_id"`
}

type fileLinkDTO struct {
	ID           string `json:"id"`
	FileID       string `json:"file_id"`
	OwnerService string `json:"owner_service"`
	OwnerType    string `json:"owner_type"`
	OwnerID      string `json:"owner_id"`
	PetID        string `json:"pet_id"`
	CreatedBy    string `json:"created_by_user_id"`
	CreatedAt    string `json:"created_at"`
}

type linkFileResponse struct {
	Link fileLinkDTO `json:"link"`
}

type unlinkFileRequest struct {
	FileID       string `json:"file_id"`
	OwnerService string `json:"owner_service"`
	OwnerType    string `json:"owner_type"`
	OwnerID      string `json:"owner_id"`
}

type unlinkFileResponse struct {
	Deleted bool `json:"deleted"`
}

type getFileResponse struct {
	File fileObjectDTO `json:"file"`
}

type listLinksResponse struct {
	Items []fileLinkDTO `json:"items"`
}

func (h *FileHandlers) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req initUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().InitUpload(ctx, &filepb.InitUploadRequest{
		MimeType:          req.MimeType,
		ExpectedSizeBytes: valueOrZero(req.ExpectedSizeBytes),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, initUploadResponse{
		File:   toFileDTO(resp.GetFile()),
		Upload: uploadDTO{
			Method:    resp.GetUpload().GetMethod(),
			URL:       resp.GetUpload().GetUrl(),
			Headers:   resp.GetUpload().GetHeaders(),
			ExpiresAt: tsToString(resp.GetUpload().GetExpiresAt()),
		},
	})
}

func (h *FileHandlers) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "file_id")

	var req confirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().ConfirmUpload(ctx, &filepb.ConfirmUploadRequest{
		FileId:    fileID,
		SizeBytes: req.SizeBytes,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, confirmUploadResponse{File: toFileDTO(resp.GetFile())})
}

func (h *FileHandlers) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "file_id")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().GetDownloadUrl(ctx, &filepb.GetDownloadUrlRequest{FileId: fileID})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, downloadURLResponse{
		FileID:    resp.GetFileId(),
		URL:       resp.GetUrl(),
		ExpiresAt: tsToString(resp.GetExpiresAt()),
	})
}

func (h *FileHandlers) BatchDownloadURLs(w http.ResponseWriter, r *http.Request) {
	var req batchDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().BatchGetDownloadUrls(ctx, &filepb.BatchGetDownloadUrlsRequest{FileIds: req.FileIDs})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	items := make([]downloadURLResponse, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, downloadURLResponse{
			FileID:    it.GetFileId(),
			URL:       it.GetUrl(),
			ExpiresAt: tsToString(it.GetExpiresAt()),
		})
	}

	writeJSON(w, http.StatusOK, batchDownloadResponse{Items: items})
}

func (h *FileHandlers) LinkFile(w http.ResponseWriter, r *http.Request) {
	var req linkFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().LinkFile(ctx, &filepb.LinkFileRequest{
		FileId:       req.FileID,
		OwnerService: mapOwnerService(req.OwnerService),
		OwnerType:    req.OwnerType,
		OwnerId:      req.OwnerID,
		PetId:        req.PetID,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, linkFileResponse{Link: toLinkDTO(resp.GetLink())})
}

func (h *FileHandlers) UnlinkFile(w http.ResponseWriter, r *http.Request) {
	var req unlinkFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().UnlinkFile(ctx, &filepb.UnlinkFileRequest{
		FileId:       req.FileID,
		OwnerService: mapOwnerService(req.OwnerService),
		OwnerType:    req.OwnerType,
		OwnerId:      req.OwnerID,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, unlinkFileResponse{Deleted: resp.GetDeleted()})
}

func (h *FileHandlers) GetFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "file_id")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().GetFile(ctx, &filepb.GetFileRequest{FileId: fileID})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, getFileResponse{File: toFileDTO(resp.GetFile())})
}

func (h *FileHandlers) ListFileLinks(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "file_id")

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing_user")
		return
	}

	ctx := grpc.WithUserID(r.Context(), userID)
	ctx, cancel := grpc.Timeout(ctx)
	defer cancel()

	resp, err := h.client.Client().ListFileLinks(ctx, &filepb.ListFileLinksRequest{FileId: fileID})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	items := make([]fileLinkDTO, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, toLinkDTO(it))
	}

	writeJSON(w, http.StatusOK, listLinksResponse{Items: items})
}

func toFileDTO(f *filepb.FileObject) fileObjectDTO {
	return fileObjectDTO{
		ID:        f.GetId(),
		Status:    f.GetStatus().String(),
		MimeType:  f.GetMimeType(),
		SizeBytes: f.GetSizeBytes(),
		Bucket:    f.GetBucket(),
		ObjectKey: f.GetObjectKey(),
		CreatedBy: f.GetCreatedByUserId(),
		CreatedAt: tsToString(f.GetCreatedAt()),
	}
}

func toLinkDTO(l *filepb.FileLink) fileLinkDTO {
	return fileLinkDTO{
		ID:           l.GetId(),
		FileID:       l.GetFileId(),
		OwnerService: l.GetOwnerService().String(),
		OwnerType:    l.GetOwnerType(),
		OwnerID:      l.GetOwnerId(),
		PetID:        l.GetPetId(),
		CreatedBy:    l.GetCreatedByUserId(),
		CreatedAt:    tsToString(l.GetCreatedAt()),
	}
}

func tsToString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return time.Unix(ts.GetSeconds(), int64(ts.GetNanos())).UTC().Format(time.RFC3339)
}

func mapOwnerService(s string) filepb.OwnerService {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "OWNER_SERVICE_")

	switch s {
	case "PROFILE":
		return filepb.OwnerService_OWNER_SERVICE_PROFILE
	case "PET":
		return filepb.OwnerService_OWNER_SERVICE_PET
	case "GUIDE":
		return filepb.OwnerService_OWNER_SERVICE_GUIDE
	case "LOG":
		return filepb.OwnerService_OWNER_SERVICE_LOG
	case "HEALTH":
		return filepb.OwnerService_OWNER_SERVICE_HEALTH
	case "CHAT":
		return filepb.OwnerService_OWNER_SERVICE_CHAT
	case "CATALOG":
		return filepb.OwnerService_OWNER_SERVICE_CATALOG
	default:
		return filepb.OwnerService_OWNER_SERVICE_UNSPECIFIED
	}
}

func valueOrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func writeErr(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
