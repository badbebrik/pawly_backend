package petclient

import (
	"chat/internal/application/ports"
	"context"
	"fmt"

	petpb "pet/proto/petpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client petpb.PetServiceClient
}

func New(addr string) (*Client, error) {
	if addr == "" {
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

func (c *Client) BatchGetBrief(ctx context.Context, petIDs []uuid.UUID) (map[uuid.UUID]ports.PetBrief, error) {
	if len(petIDs) == 0 {
		return map[uuid.UUID]ports.PetBrief{}, nil
	}

	rawIDs := make([]string, 0, len(petIDs))
	for i := range petIDs {
		rawIDs = append(rawIDs, petIDs[i].String())
	}

	resp, err := c.client.BatchGetBrief(ctx, &petpb.BatchGetBriefRequest{
		PetIds: rawIDs,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]ports.PetBrief, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		petID, err := uuid.Parse(item.GetPetId())
		if err != nil {
			continue
		}

		var avatarURL *string
		if item.GetAvatarUrl() != "" {
			value := item.GetAvatarUrl()
			avatarURL = &value
		}

		result[petID] = ports.PetBrief{
			PetID:     petID,
			Name:      item.GetName(),
			AvatarURL: avatarURL,
		}
	}

	return result, nil
}

var _ ports.PetClient = (*Client)(nil)
