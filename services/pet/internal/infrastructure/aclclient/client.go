package aclclient

import (
	"context"
	"pet/internal/service"
	aclpb "pet/proto"

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

	out := make([]uuid.UUID, 0, len(resp.GetPetIds()))
	for _, raw := range resp.GetPetIds() {
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func (c *Client) CreateOwnerMembership(ctx context.Context, petID, userID uuid.UUID) (uuid.UUID, error) {
	resp, err := c.client.CreateOwnerMembership(ctx, &aclpb.CreateOwnerMembershipRequest{
		PetId:  petID.String(),
		UserId: userID.String(),
	})
	if err != nil {
		return uuid.Nil, mapErr(err)
	}

	memberID, err := uuid.Parse(resp.GetMemberId())
	if err != nil {
		return uuid.Nil, service.ErrConflict
	}
	return memberID, nil
}

func mapAction(action string) aclpb.Action {
	switch action {
	case "pet_read":
		return aclpb.Action_ACTION_PET_READ
	case "pet_edit":
		return aclpb.Action_ACTION_PET_EDIT
	case "pet_status_change":
		return aclpb.Action_ACTION_PET_STATUS_CHANGE
	case "pet_delete":
		return aclpb.Action_ACTION_PET_DELETE
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
