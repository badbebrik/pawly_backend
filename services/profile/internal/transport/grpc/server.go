package grpc

import (
	"context"
	"errors"
	"profile/internal/service"
	profilepb "profile/proto"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	profilepb.UnimplementedProfileServiceServer
	svc *service.ProfileService
}

func NewServer(svc *service.ProfileService) *Server {
	return &Server{svc: svc}
}

func Register(srv *grpc.Server, s *Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)
	profilepb.RegisterProfileServiceServer(srv, s)
}

func (s *Server) CreateProfile(ctx context.Context, req *profilepb.CreateProfileRequest) (*profilepb.CreateProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	profile, err := s.svc.CreateProfile(ctx, service.CreateProfileInput{
		UserID:    userID,
		Locale:    optionalString(req.GetLocale()),
		FirstName: optionalString(req.GetFirstName()),
		LastName:  optionalString(req.GetLastName()),
	})
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &profilepb.CreateProfileResponse{
		UserId: profile.UserID.String(),
	}, nil
}

func mapSvcErr(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func optionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
