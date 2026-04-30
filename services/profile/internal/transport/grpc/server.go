package grpc

import (
	"context"
	"errors"
	profilepb "pawly/pkg/profilepb"
	"profile/internal/application/ports"
	profileuc "profile/internal/application/usecase"
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
	useCases *profileuc.Set
}

func NewServer(useCases *profileuc.Set) *Server {
	return &Server{useCases: useCases}
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

	profile, err := s.useCases.CreateProfile.Execute(ctx, profileuc.CreateProfileParams{
		UserID:    userID,
		Locale:    optionalString(req.GetLocale()),
		Timezone:  optionalString(req.GetTimezone()),
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

func (s *Server) DeleteProfile(ctx context.Context, req *profilepb.DeleteProfileRequest) (*profilepb.DeleteProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.useCases.DeleteProfile.Execute(ctx, userID); err != nil {
		if errors.Is(err, profileuc.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &profilepb.DeleteProfileResponse{
		UserId: userID.String(),
	}, nil
}

func (s *Server) GetPreferences(ctx context.Context, req *profilepb.GetPreferencesRequest) (*profilepb.GetPreferencesResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	prefs, err := s.useCases.GetPreferences.Execute(ctx, userID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, mapSvcErr(err)
	}

	return &profilepb.GetPreferencesResponse{
		UserId:   prefs.UserID.String(),
		Locale:   prefs.Locale,
		Timezone: prefs.Timezone,
	}, nil
}

func (s *Server) BatchGetPreferences(ctx context.Context, req *profilepb.BatchGetPreferencesRequest) (*profilepb.BatchGetPreferencesResponse, error) {
	if len(req.GetUserIds()) == 0 {
		return &profilepb.BatchGetPreferencesResponse{
			Items:           []*profilepb.ProfilePreferences{},
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

	items, notFound, err := s.useCases.BatchGetPreferences.Execute(ctx, userIDs)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	out := make([]*profilepb.ProfilePreferences, 0, len(items))
	for i := range items {
		out = append(out, &profilepb.ProfilePreferences{
			UserId:   items[i].UserID.String(),
			Locale:   items[i].Locale,
			Timezone: items[i].Timezone,
		})
	}

	notFoundRaw := make([]string, 0, len(notFound))
	for i := range notFound {
		notFoundRaw = append(notFoundRaw, notFound[i].String())
	}

	return &profilepb.BatchGetPreferencesResponse{
		Items:           out,
		NotFoundUserIds: notFoundRaw,
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

	items, notFound, err := s.useCases.BatchProfilesBrief.Execute(ctx, userIDs)
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
	case errors.Is(err, profileuc.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, profileuc.ErrInvalidLocale):
		return status.Error(codes.InvalidArgument, "invalid locale")
	case errors.Is(err, profileuc.ErrInvalidTimezone):
		return status.Error(codes.InvalidArgument, "invalid timezone")
	case errors.Is(err, profileuc.ErrAvatarUpload):
		return status.Error(codes.FailedPrecondition, "avatar_upload_failed")
	case errors.Is(err, ports.ErrConflict):
		return status.Error(codes.AlreadyExists, "profile already exists")
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
