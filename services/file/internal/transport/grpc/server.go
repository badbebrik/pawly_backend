package grpc

import (
	"context"
	"errors"
	"file/internal/application/ports"
	"file/internal/application/usecase"
	"file/internal/domain/model"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	filepb "pawly/pkg/filepb"
)

type Server struct {
	filepb.UnimplementedFileServiceServer
	useCases *usecase.Set
}

func NewServer(useCases *usecase.Set) *Server {
	return &Server{useCases: useCases}
}

func (s *Server) InitUpload(ctx context.Context, req *filepb.InitUploadRequest) (*filepb.InitUploadResponse, error) {
	params := usecase.InitUploadParams{MimeType: req.GetMimeType()}
	if req.GetExpectedSizeBytes() > 0 {
		v := req.GetExpectedSizeBytes()
		params.ExpectedSizeBytes = &v
	}
	if req.GetOriginalFilename() != "" {
		v := req.GetOriginalFilename()
		params.OriginalFilename = &v
	}

	f, upload, err := s.useCases.InitUpload(ctx, params)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.InitUploadResponse{
		File: toProtoFile(f),
		Upload: &filepb.UploadInfo{
			Method:    upload.Method,
			Url:       upload.URL,
			Headers:   upload.Headers,
			ExpiresAt: timestamppb.New(upload.ExpiresAt),
		},
	}, nil
}

func (s *Server) ConfirmUpload(ctx context.Context, req *filepb.ConfirmUploadRequest) (*filepb.ConfirmUploadResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	f, err := s.useCases.ConfirmUpload(ctx, usecase.ConfirmUploadParams{
		FileID:    fileID,
		SizeBytes: req.GetSizeBytes(),
	})
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.ConfirmUploadResponse{File: toProtoFile(f)}, nil
}

func (s *Server) GetDownloadUrl(ctx context.Context, req *filepb.GetDownloadUrlRequest) (*filepb.GetDownloadUrlResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	url, expiresAt, err := s.useCases.GetDownloadURL(ctx, fileID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.GetDownloadUrlResponse{
		FileId:    fileID.String(),
		Url:       url,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

func (s *Server) BatchGetDownloadUrls(ctx context.Context, req *filepb.BatchGetDownloadUrlsRequest) (*filepb.BatchGetDownloadUrlsResponse, error) {
	items := make([]*filepb.GetDownloadUrlResponse, 0, len(req.GetFileIds()))
	for _, idStr := range req.GetFileIds() {
		fileID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		url, expiresAt, err := s.useCases.GetDownloadURL(ctx, fileID)
		if err != nil {
			continue
		}
		items = append(items, &filepb.GetDownloadUrlResponse{
			FileId:    fileID.String(),
			Url:       url,
			ExpiresAt: timestamppb.New(expiresAt),
		})
	}

	return &filepb.BatchGetDownloadUrlsResponse{Items: items}, nil
}

func (s *Server) LinkFile(ctx context.Context, req *filepb.LinkFileRequest) (*filepb.LinkFileResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	usageType, ok := mapUsageType(req.GetOwnerService(), req.GetOwnerType())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "unsupported owner mapping")
	}

	link, err := s.useCases.Link(ctx, usecase.LinkParams{
		FileID:    fileID,
		UsageType: usageType,
		OwnerID:   ownerID,
	})
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.LinkFileResponse{Link: toProtoLink(link)}, nil
}

func (s *Server) UnlinkFile(ctx context.Context, req *filepb.UnlinkFileRequest) (*filepb.UnlinkFileResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}
	usageType, ok := mapUsageType(req.GetOwnerService(), req.GetOwnerType())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "unsupported owner mapping")
	}

	deleted, err := s.useCases.Unlink(ctx, fileID, usageType, ownerID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.UnlinkFileResponse{Deleted: deleted}, nil
}

func (s *Server) GetFile(ctx context.Context, req *filepb.GetFileRequest) (*filepb.GetFileResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	f, err := s.useCases.GetFile(ctx, fileID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.GetFileResponse{File: toProtoFile(f)}, nil
}

func (s *Server) DeleteFileIfUnlinked(ctx context.Context, req *filepb.GetFileRequest) (*filepb.GetFileResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	f, err := s.useCases.DeleteFileIfUnlinked(ctx, fileID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.GetFileResponse{File: toProtoFile(f)}, nil
}

func toProtoFile(f *model.FileObject) *filepb.FileObject {
	var size int64
	if f.SizeBytes != nil {
		size = *f.SizeBytes
	}
	return &filepb.FileObject{
		Id:               f.ID.String(),
		Status:           mapFileStatus(f.Status),
		MimeType:         f.MimeType,
		SizeBytes:        size,
		Bucket:           f.StorageBucket,
		ObjectKey:        f.StorageKey,
		CreatedByUserId:  uuid.Nil.String(),
		CreatedAt:        timestamppb.New(f.CreatedAt),
		UpdatedAt:        timestamppb.New(f.UpdatedAt),
		UploadExpiresAt:  timestamppb.New(f.UploadExpiresAt),
		OriginalFilename: strOrEmpty(f.OriginalFilename),
	}
}

func strOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func toProtoLink(l *model.FileLink) *filepb.FileLink {
	ownerService, ownerType := mapUsageTypeToProto(l.UsageType)
	return &filepb.FileLink{
		Id:              l.ID.String(),
		FileId:          l.FileID.String(),
		OwnerService:    ownerService,
		OwnerType:       ownerType,
		OwnerId:         l.OwnerID.String(),
		PetId:           "",
		CreatedByUserId: uuid.Nil.String(),
		CreatedAt:       timestamppb.New(l.CreatedAt),
	}
}

func mapSvcErr(err error) error {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, usecase.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, usecase.ErrInvalidState):
		return status.Error(codes.FailedPrecondition, "invalid state")
	case errors.Is(err, usecase.ErrUploadExpired):
		return status.Error(codes.FailedPrecondition, "upload expired")
	case errors.Is(err, usecase.ErrNotReady):
		return status.Error(codes.FailedPrecondition, "file not ready")
	case errors.Is(err, usecase.ErrHasLinks):
		return status.Error(codes.FailedPrecondition, "file still has links")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapFileStatus(s model.FileStatus) filepb.FileStatus {
	switch s {
	case model.FileStatusUploading:
		return filepb.FileStatus_FILE_STATUS_UPLOADING
	case model.FileStatusReady:
		return filepb.FileStatus_FILE_STATUS_READY
	case model.FileStatusPendingDelete:
		return filepb.FileStatus_FILE_STATUS_PENDING_DELETE
	case model.FileStatusDeleted:
		return filepb.FileStatus_FILE_STATUS_DELETED
	default:
		return filepb.FileStatus_FILE_STATUS_UNSPECIFIED
	}
}

func mapUsageType(ownerService filepb.OwnerService, ownerType string) (model.FileUsageType, bool) {
	switch ownerService {
	case filepb.OwnerService_OWNER_SERVICE_PROFILE:
		if ownerType == "AVATAR" {
			return model.FileUsageTypeUserAvatar, true
		}
	case filepb.OwnerService_OWNER_SERVICE_PET:
		if ownerType == "PET_AVATAR" {
			return model.FileUsageTypePetAvatar, true
		}
	case filepb.OwnerService_OWNER_SERVICE_LOG:
		if ownerType == "LOG_ATTACHMENT" {
			return model.FileUsageTypeLogAttachment, true
		}
	case filepb.OwnerService_OWNER_SERVICE_HEALTH:
		switch ownerType {
		case "VET_VISIT", "VACCINATION", "MEDICAL_RECORD", "PROCEDURE", "HEALTH_ATTACHMENT":
			return model.FileUsageTypeHealthAttach, true
		}
	case filepb.OwnerService_OWNER_SERVICE_CHAT:
		if ownerType == "MESSAGE_ATTACHMENT" {
			return model.FileUsageTypeChatAttachment, true
		}
	}
	return "", false
}

func mapUsageTypeToProto(usageType model.FileUsageType) (filepb.OwnerService, string) {
	switch usageType {
	case model.FileUsageTypeUserAvatar:
		return filepb.OwnerService_OWNER_SERVICE_PROFILE, "AVATAR"
	case model.FileUsageTypePetAvatar:
		return filepb.OwnerService_OWNER_SERVICE_PET, "PET_AVATAR"
	case model.FileUsageTypeLogAttachment:
		return filepb.OwnerService_OWNER_SERVICE_LOG, "LOG_ATTACHMENT"
	case model.FileUsageTypeHealthAttach:
		return filepb.OwnerService_OWNER_SERVICE_HEALTH, "HEALTH_ATTACHMENT"
	case model.FileUsageTypeChatAttachment:
		return filepb.OwnerService_OWNER_SERVICE_CHAT, "MESSAGE_ATTACHMENT"
	default:
		return filepb.OwnerService_OWNER_SERVICE_UNSPECIFIED, ""
	}
}

var _ filepb.FileServiceServer = (*Server)(nil)
