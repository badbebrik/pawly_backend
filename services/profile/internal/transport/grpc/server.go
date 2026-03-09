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

func (s *Server) BatchProfilesBrief(ctx context.Context, req *profilepb.BatchProfilesBriefRequest) (*profilepb.BatchProfilesBriefResponse, error) {
	if len(req.GetUserIds()) == 0 {
		return &profilepb.BatchProfilesBriefResponse{
			Items:           []*profilepb.ProfileBrief{},
			NotFoundUserIds: []string{},
		}, nil
	}

	userIDs := make([]uuid.UUID, 0, len(req.GetUserIds()))
	for i := range req.GetUserIds() {
		userID, err := uuid.Parse(req.GetUserIds()[i])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid user_ids")
		}
		userIDs = append(userIDs, userID)
	}

	items, notFound, err := s.svc.BatchGetProfilesBrief(ctx, userIDs)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	out := make([]*profilepb.ProfileBrief, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &profilepb.ProfileBrief{
			UserId:            item.UserID.String(),
			FirstName:         valueOrEmpty(item.FirstName),
			LastName:          valueOrEmpty(item.LastName),
			DisplayName:       buildDisplayName(item.FirstName, item.LastName),
			AvatarDownloadUrl: valueOrEmpty(item.AvatarDownloadURL),
		})
	}

	notFoundRaw := make([]string, 0, len(notFound))
	for i := range notFound {
		notFoundRaw = append(notFoundRaw, notFound[i].String())
	}

	return &profilepb.BatchProfilesBriefResponse{
		Items:           out,
		NotFoundUserIds: notFoundRaw,
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

func valueOrEmpty(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func buildDisplayName(firstName, lastName *string) string {
	parts := []string{valueOrEmpty(firstName), valueOrEmpty(lastName)}
	return strings.TrimSpace(strings.Join(parts, " "))
}
