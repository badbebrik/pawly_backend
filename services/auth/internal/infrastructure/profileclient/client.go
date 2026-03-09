package profileclient

import (
	"context"
	"fmt"

	profilepb "auth/proto/profilepb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, locale string, firstName string, lastName string) error
	Close()
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	client profilepb.ProfileServiceClient
}

func New(addr string) (*GRPCClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("profile service grpc addr is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		conn:   conn,
		client: profilepb.NewProfileServiceClient(conn),
	}, nil
}

func (c *GRPCClient) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *GRPCClient) CreateProfile(ctx context.Context, userID uuid.UUID, locale string, firstName string, lastName string) error {
	_, err := c.client.CreateProfile(ctx, &profilepb.CreateProfileRequest{
		UserId:    userID.String(),
		Locale:    locale,
		FirstName: firstName,
		LastName:  lastName,
	})
	return err
}
