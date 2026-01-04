package tokens

import "errors"

var (
	ErrInvalidToken     = errors.New("invalid_token")
	ErrExpiredToken     = errors.New("token_expired")
	ErrInvalidTokenType = errors.New("invalid_token_type")
	ErrPayloadMalformed = errors.New("payload_malformed")
)
