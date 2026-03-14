package ports

import "context"

type VerificationStore interface {
	RequestCode(ctx context.Context, email, purpose string) (code string, ttlSeconds int, resendInSeconds int, err error)
	VerifyCode(ctx context.Context, email, purpose, code string) error
}
