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

func (c *Client) ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]service.ACLMembership, error) {
	resp, err := c.client.ListPetsForUser(ctx, &aclpb.ListPetsForUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	out := make([]service.ACLMembership, 0, len(resp.GetMemberships()))
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
		role := service.ACLRole{
			Kind:  mapRoleKind(roleMsg.GetKind()),
			Title: roleMsg.GetTitle(),
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

		out = append(out, service.ACLMembership{
			PetID:          petID,
			MemberID:       memberID,
			Status:         mapMembershipStatus(member.GetStatus()),
			IsPrimaryOwner: member.GetIsPrimaryOwner(),
			Role:           role,
			Policy:         mapPolicy(member.GetPolicy()),
		})
	}

	// Fallback for older ACL service versions that return only pet_ids.
	if len(out) == 0 && len(resp.GetPetIds()) > 0 {
		for _, raw := range resp.GetPetIds() {
			id, err := uuid.Parse(raw)
			if err != nil {
				continue
			}
			out = append(out, service.ACLMembership{
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

func mapPolicy(p *aclpb.Policy) service.ACLPolicy {
	if p == nil {
		return service.ACLPolicy{}
	}
	return service.ACLPolicy{
		PetRead:                p.GetPetRead(),
		PetEdit:                p.GetPetEdit(),
		PetStatusChange:        p.GetPetStatusChange(),
		PetDelete:              p.GetPetDelete(),
		LogRead:                p.GetLogRead(),
		LogCreate:              p.GetLogCreate(),
		LogEdit:                p.GetLogEdit(),
		LogDelete:              p.GetLogDelete(),
		LogAttachmentsRead:     p.GetLogAttachmentsRead(),
		HealthRead:             p.GetHealthRead(),
		HealthWrite:            p.GetHealthWrite(),
		TaskRead:               p.GetTaskRead(),
		TaskWrite:              p.GetTaskWrite(),
		MembersView:            p.GetMembersView(),
		MembersInvite:          p.GetMembersInvite(),
		MembersRemove:          p.GetMembersRemove(),
		MembersEditPermissions: p.GetMembersEditPermissions(),
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
