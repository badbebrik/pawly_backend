package aclclient

import (
	"context"
	"fmt"
	"health/internal/service"
	aclpb "health/proto"
	"strings"

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
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("acl grpc addr is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (c *Client) Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error) {
	resp, err := c.client.Check(ctx, &aclpb.CheckRequest{
		PetId:  petID.String(),
		UserId: userID.String(),
		Action: mapAction(action),
	})
	if err != nil {
		return false, mapErr(err)
	}
	return resp.GetAllowed(), nil
}

func (c *Client) ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	resp, err := c.client.ListPetsForUser(ctx, &aclpb.ListPetsForUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	petIDs := make([]uuid.UUID, 0, len(resp.GetMemberships()))
	seen := make(map[uuid.UUID]struct{}, len(resp.GetMemberships()))
	for _, membership := range resp.GetMemberships() {
		if membership == nil || membership.GetPolicy() == nil || !membership.GetPolicy().GetHealthRead() {
			continue
		}
		petID, err := uuid.Parse(strings.TrimSpace(membership.GetPetId()))
		if err != nil || petID == uuid.Nil {
			continue
		}
		if _, ok := seen[petID]; ok {
			continue
		}
		seen[petID] = struct{}{}
		petIDs = append(petIDs, petID)
	}
	return petIDs, nil
}

func mapAction(action string) aclpb.Action {
	switch action {
	case service.ActionLogRead:
		return aclpb.Action_ACTION_LOG_READ
	case service.ActionLogWrite:
		return aclpb.Action_ACTION_LOG_WRITE
	case service.ActionHealthRead:
		return aclpb.Action_ACTION_HEALTH_READ
	case service.ActionHealthWrite:
		return aclpb.Action_ACTION_HEALTH_WRITE
	default:
		return aclpb.Action_ACTION_UNSPECIFIED
	}
}

func mapErr(err error) error {
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

var _ service.ACLClient = (*Client)(nil)
