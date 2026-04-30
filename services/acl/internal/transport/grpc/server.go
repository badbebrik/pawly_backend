package grpc

import (
	"acl/internal/application/ports"
	acluc "acl/internal/application/usecase"
	"acl/internal/domain/model"
	"context"
	"errors"
	aclpb "pawly/pkg/aclpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	aclpb.UnimplementedACLServiceServer
	useCases *acluc.Set
}

func NewServer(useCases *acluc.Set) *Server {
	return &Server{useCases: useCases}
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

	ok, err := s.useCases.IsMember(ctx, petID, userID)
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

	res, err := s.useCases.GetPolicy(ctx, petID, userID)
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

	allowed, err := s.useCases.Check(ctx, acluc.CheckParams{
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

	memberships, err := s.useCases.ListPetMembershipsForUser(ctx, userID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	petIDs := make([]string, 0, len(memberships))
	items := make([]*aclpb.PetMembership, 0, len(memberships))
	for i := range memberships {
		member := memberships[i]
		petIDs = append(petIDs, member.PetID.String())
		items = append(items, toProtoPetMembership(member))
	}
	return &aclpb.ListPetsForUserResponse{
		PetIds:      petIDs,
		Memberships: items,
	}, nil
}

func (s *Server) CreateOwnerMembership(ctx context.Context, req *aclpb.CreateOwnerMembershipRequest) (*aclpb.CreateOwnerMembershipResponse, error) {
	petID, userID, err := parsePetAndUser(req.GetPetId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	member, err := s.useCases.CreateOwnerMembership(ctx, petID, userID)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &aclpb.CreateOwnerMembershipResponse{MemberId: member.ID.String()}, nil
}

func (s *Server) TransferOwnership(ctx context.Context, req *aclpb.TransferOwnershipRequest) (*aclpb.TransferOwnershipResponse, error) {
	petID, requesterUserID, err := parsePetAndUser(req.GetPetId(), req.GetRequesterUserId())
	if err != nil {
		return nil, err
	}
	targetMemberID, err := uuid.Parse(req.GetTargetMemberId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid target_member_id")
	}

	res, err := s.useCases.TransferOwnership(ctx, acluc.TransferOwnershipParams{
		PetID:          petID,
		RequesterID:    requesterUserID,
		TargetMemberID: targetMemberID,
	})
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &aclpb.TransferOwnershipResponse{
		PreviousOwnerMemberId: res.PreviousOwner.ID.String(),
		PreviousOwnerUserId:   res.PreviousOwner.UserID.String(),
		CurrentOwnerMemberId:  res.CurrentOwner.ID.String(),
		CurrentOwnerUserId:    res.CurrentOwner.UserID.String(),
	}, nil
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
	case errors.Is(err, acluc.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, acluc.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	case errors.Is(err, acluc.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, acluc.ErrConflict):
		return status.Error(codes.FailedPrecondition, "conflict")
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
		return string(acluc.ActionPetRead)
	case aclpb.Action_ACTION_PET_WRITE:
		return string(acluc.ActionPetWrite)
	case aclpb.Action_ACTION_LOG_READ:
		return string(acluc.ActionLogRead)
	case aclpb.Action_ACTION_LOG_WRITE:
		return string(acluc.ActionLogWrite)
	case aclpb.Action_ACTION_HEALTH_READ:
		return string(acluc.ActionHealthRead)
	case aclpb.Action_ACTION_HEALTH_WRITE:
		return string(acluc.ActionHealthWrite)
	case aclpb.Action_ACTION_MEMBERS_READ:
		return string(acluc.ActionMembersRead)
	case aclpb.Action_ACTION_MEMBERS_WRITE:
		return string(acluc.ActionMembersWrite)
	default:
		return ""
	}
}

func toProtoPolicy(p model.Policy) *aclpb.Policy {
	return &aclpb.Policy{
		PetRead:      p.PetRead,
		PetWrite:     p.PetWrite,
		LogRead:      p.LogRead,
		LogWrite:     p.LogWrite,
		HealthRead:   p.HealthRead,
		HealthWrite:  p.HealthWrite,
		MembersRead:  p.MembersRead,
		MembersWrite: p.MembersWrite,
	}
}

func toProtoPetMembership(member ports.MemberView) *aclpb.PetMembership {
	return &aclpb.PetMembership{
		PetId:          member.PetID.String(),
		MemberId:       member.ID.String(),
		Status:         mapMembershipStatus(member.Status),
		IsPrimaryOwner: member.IsPrimaryOwner,
		Role:           toProtoRole(member.Role),
		Policy:         toProtoPolicy(member.Policy),
	}
}

func toProtoRole(role ports.RoleView) *aclpb.Role {
	var petID string
	if role.PetID != nil {
		petID = role.PetID.String()
	}
	var createdBy string
	if role.CreatedByUserID != nil {
		createdBy = role.CreatedByUserID.String()
	}

	return &aclpb.Role{
		Id:              role.ID.String(),
		Kind:            mapRoleKind(role.Kind),
		PetId:           petID,
		Code:            role.Code,
		Title:           role.Title,
		CreatedByUserId: createdBy,
		Policy:          toProtoPolicy(role.Policy),
	}
}

func mapRoleKind(kind string) aclpb.RoleKind {
	switch kind {
	case "SYSTEM":
		return aclpb.RoleKind_ROLE_KIND_SYSTEM
	case "CUSTOM":
		return aclpb.RoleKind_ROLE_KIND_CUSTOM
	default:
		return aclpb.RoleKind_ROLE_KIND_UNSPECIFIED
	}
}
