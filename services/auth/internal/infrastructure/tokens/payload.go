package tokens

type Payload struct {
	Sub       string `json:"sub"`
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	Exp       int64  `json:"exp"`
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeReset   = "password_reset"
)
