package aclclient

import (
	"chat/internal/application/ports"
	"context"
	"fmt"

	aclpb "acl/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client aclpb.ACLServiceClient
}

func New(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("acl service grpc addr is empty")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: aclpb.NewACLServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) IsActiveMember(ctx context.Context, petID, userID uuid.UUID) (bool, error) {
	resp, err := c.client.GetPolicy(ctx, &aclpb.GetPolicyRequest{
		PetId:  petID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		return false, mapErr(err)
	}

	return resp.GetStatus() == aclpb.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE, nil
}

func mapErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return ports.ErrNotFound
	case codes.AlreadyExists, codes.FailedPrecondition:
		return ports.ErrConflict
	default:
		return err
	}
}

var _ ports.ACLClient = (*Client)(nil)
