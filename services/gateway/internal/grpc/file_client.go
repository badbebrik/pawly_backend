package grpc

import (
	"context"
	"fmt"
	filepb "file/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type FileClient struct {
	conn   *grpc.ClientConn
	client filepb.FileServiceClient
}

func NewFileClient(addr string) (*FileClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("file service address is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &FileClient{
		conn:   conn,
		client: filepb.NewFileServiceClient(conn),
	}, nil
}

func (c *FileClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *FileClient) Client() filepb.FileServiceClient {
	return c.client
}

func WithUserID(ctx context.Context, userID string) context.Context {
	md := metadata.New(map[string]string{"x-user-id": userID})
	return metadata.NewOutgoingContext(ctx, md)
}

func Timeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}
