package ports

import "context"

type OAuthClaims struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

type OAuthVerifier interface {
	VerifyGoogleIDToken(ctx context.Context, idToken string) (*OAuthClaims, error)
}
