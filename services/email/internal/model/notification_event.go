package model

type NotificationEvent struct {
	Event    string         `json:"event"`
	UserID   string         `json:"user_id"`
	Locale   string         `json:"locale"`
	Channels []string       `json:"channels"`
	Data     map[string]any `json:"data"`
}
