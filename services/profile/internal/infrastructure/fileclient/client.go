package fileclient

import (
	"context"
	filepb "file/proto"
	"fmt"
	"profile/internal/application/ports"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) InitUpload(ctx context.Context, mimeType string, expectedSize int64, _ uuid.UUID) (uuid.UUID, ports.UploadInfo, error) {
	resp, err := c.client.InitUpload(ctx, &filepb.InitUploadRequest{
		MimeType:          mimeType,
		ExpectedSizeBytes: expectedSize,
	})
	if err != nil {
		return uuid.Nil, ports.UploadInfo{}, err
	}

	fileID, err := uuid.Parse(resp.GetFile().GetId())
	if err != nil {
		return uuid.Nil, ports.UploadInfo{}, err
	}

	upload := ports.UploadInfo{
		Method:    resp.GetUpload().GetMethod(),
		URL:       resp.GetUpload().GetUrl(),
		Headers:   resp.GetUpload().GetHeaders(),
		ExpiresAt: resp.GetUpload().GetExpiresAt().AsTime(),
	}

	return fileID, upload, nil
}

func (c *Client) ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error {
	_, err := c.client.ConfirmUpload(ctx, &filepb.ConfirmUploadRequest{
		FileId:    fileID.String(),
		SizeBytes: sizeBytes,
	})
	return err
}

func (c *Client) GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error) {
	resp, err := c.client.GetDownloadUrl(ctx, &filepb.GetDownloadUrlRequest{
		FileId: fileID.String(),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return resp.GetUrl(), resp.GetExpiresAt().AsTime(), nil
}

func (c *Client) BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(fileIDs) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	raw := make([]string, 0, len(fileIDs))
	for i := range fileIDs {
		raw = append(raw, fileIDs[i].String())
	}

	resp, err := c.client.BatchGetDownloadUrls(ctx, &filepb.BatchGetDownloadUrlsRequest{
		FileIds: raw,
	})
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]string, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		id, err := uuid.Parse(item.GetFileId())
		if err != nil {
			continue
		}
		if item.GetUrl() == "" {
			continue
		}
		out[id] = item.GetUrl()
	}
	return out, nil
}

func (c *Client) LinkAvatar(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	_, err := c.client.LinkFile(ctx, &filepb.LinkFileRequest{
		FileId:       fileID.String(),
		OwnerService: filepb.OwnerService_OWNER_SERVICE_PROFILE,
		OwnerType:    "AVATAR",
		OwnerId:      userID.String(),
	})
	return err
}

var _ ports.FileGateway = (*Client)(nil)
