package profileclient

import (
	"acl/internal/application/ports"
	"context"
	"fmt"
	"strings"

	profilepb "pawly/pkg/profilepb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client profilepb.ProfileServiceClient
}

func New(addr string) (*Client, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("profile service grpc addr is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (c *Client) BatchProfilesBrief(ctx context.Context, userIDs []uuid.UUID) ([]ports.ProfileBrief, []uuid.UUID, error) {
	if c == nil || c.client == nil {
		return nil, nil, fmt.Errorf("profile client is not configured")
	}
	if len(userIDs) == 0 {
		return []ports.ProfileBrief{}, []uuid.UUID{}, nil
	}

	rawUserIDs := make([]string, 0, len(userIDs))
	for i := range userIDs {
		rawUserIDs = append(rawUserIDs, userIDs[i].String())
	}

	resp, err := c.client.BatchProfilesBrief(ctx, &profilepb.BatchProfilesBriefRequest{
		UserIds: rawUserIDs,
	})
	if err != nil {
		return nil, nil, err
	}

	items := make([]ports.ProfileBrief, 0, len(resp.GetItems()))
	for i := range resp.GetItems() {
		item := resp.GetItems()[i]
		userID, err := uuid.Parse(item.GetUserId())
		if err != nil {
			continue
		}
		items = append(items, ports.ProfileBrief{
			UserID:            userID,
			FirstName:         optionalString(item.GetFirstName()),
			LastName:          optionalString(item.GetLastName()),
			DisplayName:       optionalString(item.GetDisplayName()),
			AvatarDownloadURL: optionalString(item.GetAvatarDownloadUrl()),
		})
	}

	notFound := make([]uuid.UUID, 0, len(resp.GetNotFoundUserIds()))
	for i := range resp.GetNotFoundUserIds() {
		id, err := uuid.Parse(resp.GetNotFoundUserIds()[i])
		if err != nil {
			continue
		}
		notFound = append(notFound, id)
	}

	return items, notFound, nil
}

var _ ports.ProfileClient = (*Client)(nil)

func optionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
