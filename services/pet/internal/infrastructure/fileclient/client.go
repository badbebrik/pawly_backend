package fileclient

import (
	"context"
	"fmt"
	"pet/internal/service"
	filepb "pet/proto/filepb"
	"time"

	"github.com/google/uuid"
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
	if addr == "" {
		return nil, fmt.Errorf("file service addr is empty")
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

func (c *Client) InitUpload(ctx context.Context, mimeType string, expectedSize int64) (uuid.UUID, service.UploadInfo, error) {
	resp, err := c.client.InitUpload(ctx, &filepb.InitUploadRequest{
		MimeType:          mimeType,
		ExpectedSizeBytes: expectedSize,
	})
	if err != nil {
		return uuid.Nil, service.UploadInfo{}, mapErr(err)
	}

	fileID, err := uuid.Parse(resp.GetFile().GetId())
	if err != nil {
		return uuid.Nil, service.UploadInfo{}, service.ErrConflict
	}

	return fileID, service.UploadInfo{
		Method:    resp.GetUpload().GetMethod(),
		URL:       resp.GetUpload().GetUrl(),
		Headers:   resp.GetUpload().GetHeaders(),
		ExpiresAt: resp.GetUpload().GetExpiresAt().AsTime(),
	}, nil
}

func (c *Client) ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error {
	_, err := c.client.ConfirmUpload(ctx, &filepb.ConfirmUploadRequest{
		FileId:    fileID.String(),
		SizeBytes: sizeBytes,
	})
	return mapErr(err)
}

func (c *Client) GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error) {
	resp, err := c.client.GetDownloadUrl(ctx, &filepb.GetDownloadUrlRequest{
		FileId: fileID.String(),
	})
	if err != nil {
		return "", time.Time{}, mapErr(err)
	}
	return resp.GetUrl(), resp.GetExpiresAt().AsTime(), nil
}

func (c *Client) BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(fileIDs) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	in := make([]string, 0, len(fileIDs))
	for _, id := range fileIDs {
		in = append(in, id.String())
	}

	resp, err := c.client.BatchGetDownloadUrls(ctx, &filepb.BatchGetDownloadUrlsRequest{
		FileIds: in,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	out := make(map[uuid.UUID]string, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		fileID, err := uuid.Parse(item.GetFileId())
		if err != nil {
			continue
		}
		if item.GetUrl() == "" {
			continue
		}
		out[fileID] = item.GetUrl()
	}
	return out, nil
}

func (c *Client) LinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error {
	_, err := c.client.LinkFile(ctx, &filepb.LinkFileRequest{
		FileId:       fileID.String(),
		OwnerService: filepb.OwnerService_OWNER_SERVICE_PET,
		OwnerType:    "PET_AVATAR",
		OwnerId:      petID.String(),
		PetId:        petID.String(),
	})
	return mapErr(err)
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
