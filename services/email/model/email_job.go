package model

type EmailJob struct {
	Email    string         `json:"email"`
	Template string         `json:"template"`
	Locale   string         `json:"locale"`
	Subject  string         `json:"subject"`
	Data     map[string]any `json:"data"`
	Meta     map[string]any `json:"meta"`
}
