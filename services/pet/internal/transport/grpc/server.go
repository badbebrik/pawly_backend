package grpc

import (
	"context"
	"pet/internal/service"
	petpb "pet/proto/petpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	petpb.UnimplementedPetServiceServer
	svc *service.PetService
}

func NewServer(svc *service.PetService) *Server {
	return &Server{svc: svc}
}

func Register(srv *grpc.Server, s *Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)
	petpb.RegisterPetServiceServer(srv, s)
}

func (s *Server) BatchGetBrief(ctx context.Context, req *petpb.BatchGetBriefRequest) (*petpb.BatchGetBriefResponse, error) {
	if len(req.GetPetIds()) == 0 {
		return &petpb.BatchGetBriefResponse{
			Items:          []*petpb.PetBrief{},
			NotFoundPetIds: []string{},
		}, nil
	}

	petIDs := make([]uuid.UUID, 0, len(req.GetPetIds()))
	for i := range req.GetPetIds() {
		petID, err := uuid.Parse(req.GetPetIds()[i])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid pet_ids")
		}
		petIDs = append(petIDs, petID)
	}

	items, notFound, err := s.svc.BatchGetBrief(ctx, petIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	out := make([]*petpb.PetBrief, 0, len(items))
	for _, petID := range petIDs {
		item, ok := items[petID]
		if !ok {
			continue
		}

		avatarURL := ""
		if item.AvatarURL != nil {
			avatarURL = *item.AvatarURL
		}

		out = append(out, &petpb.PetBrief{
			PetId:     item.PetID.String(),
			Name:      item.Name,
			AvatarUrl: avatarURL,
		})
	}

	notFoundRaw := make([]string, 0, len(notFound))
	for i := range notFound {
		notFoundRaw = append(notFoundRaw, notFound[i].String())
	}

	return &petpb.BatchGetBriefResponse{
		Items:          out,
		NotFoundPetIds: notFoundRaw,
	}, nil
}
