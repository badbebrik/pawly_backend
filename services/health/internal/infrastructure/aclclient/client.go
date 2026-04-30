package aclclient

import (
	"context"
	"fmt"
	"health/internal/application/ports"
	aclpb "pawly/pkg/aclpb"
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

func (c *Client) ListPetAccessForUser(ctx context.Context, userID uuid.UUID) ([]ports.PetAccess, error) {
	resp, err := c.client.ListPetsForUser(ctx, &aclpb.ListPetsForUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	items := make([]ports.PetAccess, 0, len(resp.GetMemberships()))
	seen := make(map[uuid.UUID]struct{}, len(resp.GetMemberships()))
	for _, membership := range resp.GetMemberships() {
		if membership == nil || membership.GetPolicy() == nil {
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
		policy := membership.GetPolicy()
		items = append(items, ports.PetAccess{
			PetID:       petID,
			PetRead:     policy.GetPetRead(),
			PetWrite:    policy.GetPetWrite(),
			LogRead:     policy.GetLogRead(),
			LogWrite:    policy.GetLogWrite(),
			HealthRead:  policy.GetHealthRead(),
			HealthWrite: policy.GetHealthWrite(),
		})
	}
	return items, nil
}

func mapAction(action string) aclpb.Action {
	switch action {
	case ports.ActionPetRead:
		return aclpb.Action_ACTION_PET_READ
	case ports.ActionPetWrite:
		return aclpb.Action_ACTION_PET_WRITE
	case ports.ActionLogRead:
		return aclpb.Action_ACTION_LOG_READ
	case ports.ActionLogWrite:
		return aclpb.Action_ACTION_LOG_WRITE
	case ports.ActionHealthRead:
		return aclpb.Action_ACTION_HEALTH_READ
	case ports.ActionHealthWrite:
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
		return ports.ErrNotFound
	case codes.PermissionDenied:
		return ports.ErrForbidden
	case codes.InvalidArgument:
		return ports.ErrInvalidInput
	case codes.FailedPrecondition, codes.AlreadyExists:
		return ports.ErrConflict
	default:
		return err
	}
}

var _ ports.HealthAccessChecker = (*Client)(nil)
var _ ports.HealthPetLister = (*Client)(nil)
