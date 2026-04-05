package fileclient

import (
	"context"
	"fmt"
	"health/internal/model"
	"health/internal/service"
	"strings"
	"time"

	"github.com/google/uuid"
	filepb "pawly/pkg/filepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client filepb.FileServiceClient
}

func New(addr string) (*Client, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("file grpc addr is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: filepb.NewFileServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) InitUpload(ctx context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, service.UploadInfo, error) {
	resp, err := c.client.InitUpload(ctx, &filepb.InitUploadRequest{
		MimeType:          mimeType,
		ExpectedSizeBytes: expectedSize,
		OriginalFilename:  strings.TrimSpace(originalFilename),
	})
	if err != nil {
		return uuid.Nil, service.UploadInfo{}, mapErr(err)
	}

	fileID, err := uuid.Parse(resp.GetFile().GetId())
	if err != nil {
		return uuid.Nil, service.UploadInfo{}, err
	}

	expiresAt := time.Time{}
	if resp.GetUpload().GetExpiresAt() != nil {
		expiresAt = resp.GetUpload().GetExpiresAt().AsTime()
	}

	return fileID, service.UploadInfo{
		Method:    resp.GetUpload().GetMethod(),
		URL:       resp.GetUpload().GetUrl(),
		Headers:   resp.GetUpload().GetHeaders(),
		ExpiresAt: expiresAt,
	}, nil
}

func (c *Client) ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) (*service.UploadedFile, error) {
	resp, err := c.client.ConfirmUpload(ctx, &filepb.ConfirmUploadRequest{
		FileId:    fileID.String(),
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	file := resp.GetFile()
	if file == nil {
		return nil, service.ErrConflict
	}

	return &service.UploadedFile{
		ID:               fileID,
		MimeType:         file.GetMimeType(),
		SizeBytes:        file.GetSizeBytes(),
		OriginalFilename: optionalString(file.GetOriginalFilename()),
	}, nil
}

func (c *Client) EnsureFilesExist(ctx context.Context, fileIDs []uuid.UUID) error {
	for i := range fileIDs {
		_, err := c.client.GetFile(ctx, &filepb.GetFileRequest{
			FileId: fileIDs[i].String(),
		})
		if err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func (c *Client) BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(fileIDs) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	in := make([]string, 0, len(fileIDs))
	for i := range fileIDs {
		in = append(in, fileIDs[i].String())
	}
	resp, err := c.client.BatchGetDownloadUrls(ctx, &filepb.BatchGetDownloadUrlsRequest{
		FileIds: in,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make(map[uuid.UUID]string, len(resp.GetItems()))
	for i := range resp.GetItems() {
		item := resp.GetItems()[i]
		id, err := uuid.Parse(item.GetFileId())
		if err != nil {
			continue
		}
		if strings.TrimSpace(item.GetUrl()) == "" {
			continue
		}
		out[id] = item.GetUrl()
	}
	return out, nil
}

func (c *Client) GetFiles(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]model.HealthFile, error) {
	out := make(map[uuid.UUID]model.HealthFile, len(fileIDs))
	for i := range fileIDs {
		resp, err := c.client.GetFile(ctx, &filepb.GetFileRequest{FileId: fileIDs[i].String()})
		if err != nil {
			return nil, mapErr(err)
		}
		file := resp.GetFile()
		out[fileIDs[i]] = model.HealthFile{
			ID:       fileIDs[i],
			MimeType: file.GetMimeType(),
			FileName: optionalString(file.GetOriginalFilename()),
		}
	}
	return out, nil
}

func (c *Client) LinkAttachments(ctx context.Context, petID uuid.UUID, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error {
	ownerService := filepb.OwnerService_OWNER_SERVICE_HEALTH
	ownerType := "HEALTH_ATTACHMENT"
	if entityType == "LOG" {
		ownerService = filepb.OwnerService_OWNER_SERVICE_LOG
		ownerType = "LOG_ATTACHMENT"
	}
	for i := range fileIDs {
		_, err := c.client.LinkFile(ctx, &filepb.LinkFileRequest{
			FileId:       fileIDs[i].String(),
			OwnerService: ownerService,
			OwnerType:    ownerType,
			OwnerId:      entityID.String(),
			PetId:        petID.String(),
		})
		if err != nil {
			mapped := mapErr(err)
			if mapped == service.ErrConflict {
				continue
			}
			return mapped
		}
	}
	return nil
}

func (c *Client) UnlinkAttachments(ctx context.Context, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error {
	ownerService := filepb.OwnerService_OWNER_SERVICE_HEALTH
	ownerType := "HEALTH_ATTACHMENT"
	if entityType == "LOG" {
		ownerService = filepb.OwnerService_OWNER_SERVICE_LOG
		ownerType = "LOG_ATTACHMENT"
	}
	for i := range fileIDs {
		_, err := c.client.UnlinkFile(ctx, &filepb.UnlinkFileRequest{
			FileId:       fileIDs[i].String(),
			OwnerService: ownerService,
			OwnerType:    ownerType,
			OwnerId:      entityID.String(),
		})
		if err != nil {
			mapped := mapErr(err)
			if mapped == service.ErrNotFound {
				continue
			}
			return mapped
		}
	}
	return nil
}

func (c *Client) DeleteFilesIfUnlinked(ctx context.Context, fileIDs []uuid.UUID) error {
	for i := range fileIDs {
		_, err := c.client.DeleteFileIfUnlinked(ctx, &filepb.GetFileRequest{
			FileId: fileIDs[i].String(),
		})
		if err != nil {
			mapped := mapErr(err)
			if mapped == service.ErrConflict || mapped == service.ErrNotFound {
				continue
			}
			return mapped
		}
	}
	return nil
}

func optionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.NotFound:
		return service.ErrNotFound
	case codes.PermissionDenied:
		return service.ErrForbidden
	case codes.InvalidArgument:
		return service.ErrInvalidInput
	case codes.FailedPrecondition, codes.AlreadyExists:
		return service.ErrConflict
	default:
		return err
	}
}

var _ service.FileClient = (*Client)(nil)
