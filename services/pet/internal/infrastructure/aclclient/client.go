package aclclient

import (
	"context"
	aclpb "pawly/pkg/aclpb"
	"pet/internal/application/ports"

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

func (c *Client) ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]ports.ACLMembership, error) {
	resp, err := c.client.ListPetsForUser(ctx, &aclpb.ListPetsForUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	out := make([]ports.ACLMembership, 0, len(resp.GetMemberships()))
	for _, member := range resp.GetMemberships() {
		petID, err := uuid.Parse(member.GetPetId())
		if err != nil {
			continue
		}

		memberID, err := uuid.Parse(member.GetMemberId())
		if err != nil {
			continue
		}

		roleMsg := member.GetRole()
		role := ports.ACLRole{
			Kind:   mapRoleKind(roleMsg.GetKind()),
			Title:  roleMsg.GetTitle(),
			Policy: mapPolicy(roleMsg.GetPolicy()),
		}
		if roleID, err := uuid.Parse(roleMsg.GetId()); err == nil {
			role.ID = roleID
		}
		if rolePetID, err := uuid.Parse(roleMsg.GetPetId()); err == nil {
			role.PetID = &rolePetID
		}
		if rawCode := roleMsg.GetCode(); rawCode != "" {
			role.Code = &rawCode
		}
		if createdBy, err := uuid.Parse(roleMsg.GetCreatedByUserId()); err == nil {
			role.CreatedByUserID = &createdBy
		}

		out = append(out, ports.ACLMembership{
			PetID:          petID,
			MemberID:       memberID,
			Status:         mapMembershipStatus(member.GetStatus()),
			IsPrimaryOwner: member.GetIsPrimaryOwner(),
			Role:           role,
			Policy:         mapPolicy(member.GetPolicy()),
		})
	}

		if len(out) == 0 && len(resp.GetPetIds()) > 0 {
		for _, raw := range resp.GetPetIds() {
			id, err := uuid.Parse(raw)
			if err != nil {
				continue
			}
			out = append(out, ports.ACLMembership{
				PetID: id,
			})
		}
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
		return uuid.Nil, ports.ErrConflict
	}
	return memberID, nil
}

func (c *Client) TransferOwnership(ctx context.Context, petID, requesterUserID, targetMemberID uuid.UUID) (ports.ACLTransferOwnershipResult, error) {
	resp, err := c.client.TransferOwnership(ctx, &aclpb.TransferOwnershipRequest{
		PetId:           petID.String(),
		RequesterUserId: requesterUserID.String(),
		TargetMemberId:  targetMemberID.String(),
	})
	if err != nil {
		return ports.ACLTransferOwnershipResult{}, mapErr(err)
	}

	result := ports.ACLTransferOwnershipResult{}

	if result.PreviousOwnerMemberID, err = uuid.Parse(resp.GetPreviousOwnerMemberId()); err != nil {
		return ports.ACLTransferOwnershipResult{}, ports.ErrConflict
	}
	if result.PreviousOwnerUserID, err = uuid.Parse(resp.GetPreviousOwnerUserId()); err != nil {
		return ports.ACLTransferOwnershipResult{}, ports.ErrConflict
	}
	if result.CurrentOwnerMemberID, err = uuid.Parse(resp.GetCurrentOwnerMemberId()); err != nil {
		return ports.ACLTransferOwnershipResult{}, ports.ErrConflict
	}
	if result.CurrentOwnerUserID, err = uuid.Parse(resp.GetCurrentOwnerUserId()); err != nil {
		return ports.ACLTransferOwnershipResult{}, ports.ErrConflict
	}

	return result, nil
}

func mapAction(action string) aclpb.Action {
	switch action {
	case "pet_read":
		return aclpb.Action_ACTION_PET_READ
	case "pet_write":
		return aclpb.Action_ACTION_PET_WRITE
	default:
		return aclpb.Action_ACTION_UNSPECIFIED
	}
}

func mapMembershipStatus(status aclpb.MembershipStatus) string {
	switch status {
	case aclpb.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE:
		return "ACTIVE"
	case aclpb.MembershipStatus_MEMBERSHIP_STATUS_REMOVED:
		return "REMOVED"
	default:
		return ""
	}
}

func mapRoleKind(kind aclpb.RoleKind) string {
	switch kind {
	case aclpb.RoleKind_ROLE_KIND_SYSTEM:
		return "SYSTEM"
	case aclpb.RoleKind_ROLE_KIND_CUSTOM:
		return "CUSTOM"
	default:
		return ""
	}
}

func mapPolicy(p *aclpb.Policy) ports.ACLPolicy {
	if p == nil {
		return ports.ACLPolicy{}
	}
	return ports.ACLPolicy{
		PetRead:      p.GetPetRead(),
		PetWrite:     p.GetPetWrite(),
		LogRead:      p.GetLogRead(),
		LogWrite:     p.GetLogWrite(),
		HealthRead:   p.GetHealthRead(),
		HealthWrite:  p.GetHealthWrite(),
		MembersRead:  p.GetMembersRead(),
		MembersWrite: p.GetMembersWrite(),
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

var _ ports.ACLClient = (*Client)(nil)
