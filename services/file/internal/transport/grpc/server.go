package grpc

import (
	"context"
	"errors"
	"file/internal/model"
	"file/internal/repository"
	"file/internal/service"
	filepb "file/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	filepb.UnimplementedFileServiceServer
	svc *service.FileService
}

func NewServer(svc *service.FileService) *Server {
	return &Server{svc: svc}
}

func (s *Server) InitUpload(ctx context.Context, req *filepb.InitUploadRequest) (*filepb.InitUploadResponse, error) {
	params := service.InitUploadParams{
		MimeType: req.GetMimeType(),
	}
	if req.GetExpectedSizeBytes() > 0 {
		v := req.GetExpectedSizeBytes()
		params.ExpectedSizeBytes = &v
	}

	params.CreatedByUserID = uuid.Nil

	f, upload, err := s.svc.InitUpload(ctx, params)
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

	f, err := s.svc.ConfirmUpload(ctx, service.ConfirmUploadParams{
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

	url, expiresAt, err := s.svc.GetDownloadURL(ctx, fileID)
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
		url, expiresAt, err := s.svc.GetDownloadURL(ctx, fileID)
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
	var petID *uuid.UUID
	if req.GetPetId() != "" {
		v, err := uuid.Parse(req.GetPetId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid pet_id")
		}
		petID = &v
	}

	link, err := s.svc.Link(ctx, service.LinkParams{
		FileID:        fileID,
		OwnerService:  mapOwnerService(req.GetOwnerService()),
		OwnerType:     req.GetOwnerType(),
		OwnerID:       ownerID,
		PetID:         petID,
		CreatedByUser: uuid.Nil,
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

	deleted, err := s.svc.Unlink(ctx, fileID, mapOwnerService(req.GetOwnerService()), req.GetOwnerType(), ownerID)
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

	f, err := s.svc.GetFile(ctx, fileID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &filepb.GetFileResponse{File: toProtoFile(f)}, nil
}

func (s *Server) ListFileLinks(ctx context.Context, req *filepb.ListFileLinksRequest) (*filepb.ListFileLinksResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	items, err := s.svc.ListLinksByFileID(ctx, fileID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	res := make([]*filepb.FileLink, 0, len(items))
	for i := range items {
		res = append(res, toProtoLink(&items[i]))
	}
	return &filepb.ListFileLinksResponse{Items: res}, nil
}

func toProtoFile(f *model.FileObject) *filepb.FileObject {
	var size int64
	if f.SizeBytes != nil {
		size = *f.SizeBytes
	}
	return &filepb.FileObject{
		Id:              f.ID.String(),
		Status:          mapFileStatus(f.Status),
		MimeType:        f.MimeType,
		SizeBytes:       size,
		Bucket:          f.Bucket,
		ObjectKey:       f.ObjectKey,
		CreatedByUserId: f.CreatedByUserID.String(),
		CreatedAt:       timestamppb.New(f.CreatedAt),
		UpdatedAt:       timestamppb.New(f.UpdatedAt),
		UploadExpiresAt: timestamppb.New(f.UploadExpiresAt),
	}
}

func toProtoLink(l *model.FileLink) *filepb.FileLink {
	petID := ""
	if l.PetID != nil {
		petID = l.PetID.String()
	}
	return &filepb.FileLink{
		Id:              l.ID.String(),
		FileId:          l.FileID.String(),
		OwnerService:    mapOwnerServiceToProto(l.OwnerService),
		OwnerType:       l.OwnerType,
		OwnerId:         l.OwnerID.String(),
		PetId:           petID,
		CreatedByUserId: l.CreatedByUserID.String(),
		CreatedAt:       timestamppb.New(l.CreatedAt),
	}
}

func mapSvcErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, service.ErrInvalidState):
		return status.Error(codes.FailedPrecondition, "invalid state")
	case errors.Is(err, service.ErrUploadExpired):
		return status.Error(codes.FailedPrecondition, "upload expired")
	case errors.Is(err, service.ErrNotReady):
		return status.Error(codes.FailedPrecondition, "file not ready")
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
	case model.FileStatusFailed:
		return filepb.FileStatus_FILE_STATUS_FAILED
	case model.FileStatusDeleted:
		return filepb.FileStatus_FILE_STATUS_DELETED
	default:
		return filepb.FileStatus_FILE_STATUS_UNSPECIFIED
	}
}

func mapOwnerService(s filepb.OwnerService) model.OwnerService {
	switch s {
	case filepb.OwnerService_OWNER_SERVICE_PROFILE:
		return model.OwnerServiceProfile
	case filepb.OwnerService_OWNER_SERVICE_PET:
		return model.OwnerServicePet
	case filepb.OwnerService_OWNER_SERVICE_GUIDE:
		return model.OwnerServiceGuide
	case filepb.OwnerService_OWNER_SERVICE_LOG:
		return model.OwnerServiceLog
	case filepb.OwnerService_OWNER_SERVICE_HEALTH:
		return model.OwnerServiceHealth
	case filepb.OwnerService_OWNER_SERVICE_CHAT:
		return model.OwnerServiceChat
	case filepb.OwnerService_OWNER_SERVICE_CATALOG:
		return model.OwnerServiceCatalog
	default:
		return ""
	}
}

func mapOwnerServiceToProto(s model.OwnerService) filepb.OwnerService {
	switch s {
	case model.OwnerServiceProfile:
		return filepb.OwnerService_OWNER_SERVICE_PROFILE
	case model.OwnerServicePet:
		return filepb.OwnerService_OWNER_SERVICE_PET
	case model.OwnerServiceGuide:
		return filepb.OwnerService_OWNER_SERVICE_GUIDE
	case model.OwnerServiceLog:
		return filepb.OwnerService_OWNER_SERVICE_LOG
	case model.OwnerServiceHealth:
		return filepb.OwnerService_OWNER_SERVICE_HEALTH
	case model.OwnerServiceChat:
		return filepb.OwnerService_OWNER_SERVICE_CHAT
	case model.OwnerServiceCatalog:
		return filepb.OwnerService_OWNER_SERVICE_CATALOG
	default:
		return filepb.OwnerService_OWNER_SERVICE_UNSPECIFIED
	}
}

var _ filepb.FileServiceServer = (*Server)(nil)
