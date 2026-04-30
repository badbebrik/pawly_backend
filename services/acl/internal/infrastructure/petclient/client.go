package petclient

import (
	"acl/internal/application/ports"
	"context"
	"fmt"
	"strings"

	petpb "pawly/pkg/petpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client petpb.PetServiceClient
}

func New(addr string) (*Client, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("pet service grpc addr is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: petpb.NewPetServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) BatchGetBrief(ctx context.Context, petIDs []uuid.UUID) ([]ports.PetBrief, []uuid.UUID, error) {
	if c == nil || c.client == nil {
		return nil, nil, fmt.Errorf("pet client is not configured")
	}
	if len(petIDs) == 0 {
		return []ports.PetBrief{}, []uuid.UUID{}, nil
	}

	rawPetIDs := make([]string, 0, len(petIDs))
	for i := range petIDs {
		rawPetIDs = append(rawPetIDs, petIDs[i].String())
	}

	resp, err := c.client.BatchGetBrief(ctx, &petpb.BatchGetBriefRequest{
		PetIds: rawPetIDs,
	})
	if err != nil {
		return nil, nil, err
	}

	items := make([]ports.PetBrief, 0, len(resp.GetItems()))
	for i := range resp.GetItems() {
		item := resp.GetItems()[i]
		petID, err := uuid.Parse(item.GetPetId())
		if err != nil {
			continue
		}
		items = append(items, ports.PetBrief{
			PetID:            petID,
			Name:             item.GetName(),
			PhotoDownloadURL: optionalString(item.GetAvatarUrl()),
		})
	}

	notFound := make([]uuid.UUID, 0, len(resp.GetNotFoundPetIds()))
	for i := range resp.GetNotFoundPetIds() {
		id, err := uuid.Parse(resp.GetNotFoundPetIds()[i])
		if err != nil {
			continue
		}
		notFound = append(notFound, id)
	}

	return items, notFound, nil
}

var _ ports.PetClient = (*Client)(nil)

func optionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
