package verification

import "time"

type CodeRecord struct {
	Code              string    `json:"code"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	ResendAvailableAt time.Time `json:"resend_available_at"`
	Attempts          int       `json:"attempts"`
}
