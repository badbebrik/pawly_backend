package profileclient

import (
	"chat/internal/application/ports"
	"context"
	"fmt"

	profilepb "profile/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client profilepb.ProfileServiceClient
}

func New(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("profile service grpc addr is empty")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: profilepb.NewProfileServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) BatchGetBrief(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]ports.ProfileBrief, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]ports.ProfileBrief{}, nil
	}

	rawIDs := make([]string, 0, len(userIDs))
	for i := range userIDs {
		rawIDs = append(rawIDs, userIDs[i].String())
	}

	resp, err := c.client.BatchProfilesBrief(ctx, &profilepb.BatchProfilesBriefRequest{
		UserIds: rawIDs,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]ports.ProfileBrief, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		userID, err := uuid.Parse(item.GetUserId())
		if err != nil {
			continue
		}

		var displayName *string
		if item.GetDisplayName() != "" {
			value := item.GetDisplayName()
			displayName = &value
		}

		var avatarURL *string
		if item.GetAvatarDownloadUrl() != "" {
			value := item.GetAvatarDownloadUrl()
			avatarURL = &value
		}

		result[userID] = ports.ProfileBrief{
			UserID:      userID,
			DisplayName: displayName,
			AvatarURL:   avatarURL,
		}
	}

	return result, nil
}

var _ ports.ProfileClient = (*Client)(nil)
