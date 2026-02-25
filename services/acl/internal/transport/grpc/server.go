package grpc

import (
	"acl/internal/model"
	"acl/internal/service"
	aclpb "acl/proto"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	aclpb.UnimplementedACLServiceServer
	svc *service.ACLService
}

func NewServer(svc *service.ACLService) *Server {
	return &Server{svc: svc}
}

func Register(srv *grpc.Server, s *Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)
	aclpb.RegisterACLServiceServer(srv, s)
}

func (s *Server) IsMember(ctx context.Context, req *aclpb.IsMemberRequest) (*aclpb.IsMemberResponse, error) {
	petID, userID, err := parsePetAndUser(req.GetPetId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	ok, err := s.svc.IsMember(ctx, petID, userID)
	if err != nil {
		return nil, mapSvcErr(err)
	}
	return &aclpb.IsMemberResponse{IsMember: ok}, nil
}

func (s *Server) GetPolicy(ctx context.Context, req *aclpb.GetPolicyRequest) (*aclpb.GetPolicyResponse, error) {
	petID, userID, err := parsePetAndUser(req.GetPetId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	res, err := s.svc.GetPolicy(ctx, petID, userID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &aclpb.GetPolicyResponse{
		MemberId:       res.MemberID.String(),
		Status:         mapMembershipStatus(res.Status),
		IsPrimaryOwner: res.IsPrimaryOwner,
		Policy:         toProtoPolicy(res.Policy),
	}, nil
}

func (s *Server) Check(ctx context.Context, req *aclpb.CheckRequest) (*aclpb.CheckResponse, error) {
	petID, userID, err := parsePetAndUser(req.GetPetId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	allowed, err := s.svc.Check(ctx, service.CheckParams{
		PetID:  petID,
		UserID: userID,
		Action: mapAction(req.GetAction()),
	})
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &aclpb.CheckResponse{Allowed: allowed}, nil
}

func (s *Server) ListPetsForUser(ctx context.Context, req *aclpb.ListPetsForUserRequest) (*aclpb.ListPetsForUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	petIDs, err := s.svc.ListPetsForUser(ctx, userID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	items := make([]string, 0, len(petIDs))
	for _, id := range petIDs {
		items = append(items, id.String())
	}
	return &aclpb.ListPetsForUserResponse{PetIds: items}, nil
}

func parsePetAndUser(petRaw, userRaw string) (uuid.UUID, uuid.UUID, error) {
	petID, err := uuid.Parse(petRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, status.Error(codes.InvalidArgument, "invalid pet_id")
	}

	userID, err := uuid.Parse(userRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	return petID, userID, nil
}

func mapSvcErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapMembershipStatus(s string) aclpb.MembershipStatus {
	switch s {
	case "ACTIVE":
		return aclpb.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE
	case "REMOVED":
		return aclpb.MembershipStatus_MEMBERSHIP_STATUS_REMOVED
	default:
		return aclpb.MembershipStatus_MEMBERSHIP_STATUS_UNSPECIFIED
	}
}

func mapAction(action aclpb.Action) string {
	switch action {
	case aclpb.Action_ACTION_PET_READ:
		return string(service.ActionPetRead)
	case aclpb.Action_ACTION_PET_EDIT:
		return string(service.ActionPetEdit)
	case aclpb.Action_ACTION_PET_STATUS_CHANGE:
		return string(service.ActionPetStatusChange)
	case aclpb.Action_ACTION_PET_DELETE:
		return string(service.ActionPetDelete)
	case aclpb.Action_ACTION_LOG_READ:
		return string(service.ActionLogRead)
	case aclpb.Action_ACTION_LOG_CREATE:
		return string(service.ActionLogCreate)
	case aclpb.Action_ACTION_LOG_EDIT:
		return string(service.ActionLogEdit)
	case aclpb.Action_ACTION_LOG_DELETE:
		return string(service.ActionLogDelete)
	case aclpb.Action_ACTION_LOG_ATTACHMENTS_READ:
		return string(service.ActionLogAttachmentsRead)
	case aclpb.Action_ACTION_HEALTH_READ:
		return string(service.ActionHealthRead)
	case aclpb.Action_ACTION_HEALTH_WRITE:
		return string(service.ActionHealthWrite)
	case aclpb.Action_ACTION_TASK_READ:
		return string(service.ActionTaskRead)
	case aclpb.Action_ACTION_TASK_WRITE:
		return string(service.ActionTaskWrite)
	case aclpb.Action_ACTION_MEMBERS_VIEW:
		return string(service.ActionMembersView)
	case aclpb.Action_ACTION_MEMBERS_INVITE:
		return string(service.ActionMembersInvite)
	case aclpb.Action_ACTION_MEMBERS_REMOVE:
		return string(service.ActionMembersRemove)
	case aclpb.Action_ACTION_MEMBERS_EDIT_PERMISSIONS:
		return string(service.ActionMembersEditPermissions)
	default:
		return ""
	}
}

func toProtoPolicy(p model.Policy) *aclpb.Policy {
	return &aclpb.Policy{
		PetRead:                p.PetRead,
		PetEdit:                p.PetEdit,
		PetStatusChange:        p.PetStatusChange,
		PetDelete:              p.PetDelete,
		LogRead:                p.LogRead,
		LogCreate:              p.LogCreate,
		LogEdit:                p.LogEdit,
		LogDelete:              p.LogDelete,
		LogAttachmentsRead:     p.LogAttachmentsRead,
		HealthRead:             p.HealthRead,
		HealthWrite:            p.HealthWrite,
		TaskRead:               p.TaskRead,
		TaskWrite:              p.TaskWrite,
		MembersView:            p.MembersView,
		MembersInvite:          p.MembersInvite,
		MembersRemove:          p.MembersRemove,
		MembersEditPermissions: p.MembersEditPermissions,
	}
}
