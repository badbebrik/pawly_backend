package usecase

import (
	"auth/internal/application/ports"
	"time"
)

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

type Dependencies struct {
	Users        ports.UserRepository
	Sessions     ports.SessionRepository
	OAuth        ports.OAuthIdentityRepository
	ResetTokens  ports.ResetTokenRepository
	Verification ports.VerificationStore
	Notifier     ports.Notifier
	Tokens       ports.TokenManager
	Profiles     ports.ProfileProvisioner
	OAuthVerify  ports.OAuthVerifier
	Clock        ports.Clock
}

type dependencies struct {
	users        ports.UserRepository
	sessions     ports.SessionRepository
	oauth        ports.OAuthIdentityRepository
	resetTokens  ports.ResetTokenRepository
	verification ports.VerificationStore
	notifier     ports.Notifier
	tokens       ports.TokenManager
	profiles     ports.ProfileProvisioner
	oauthVerify  ports.OAuthVerifier
	clock        ports.Clock
}

func newDependencies(in Dependencies) *dependencies {
	clk := in.Clock
	if clk == nil {
		clk = systemClock{}
	}

	return &dependencies{
		users:        in.Users,
		sessions:     in.Sessions,
		oauth:        in.OAuth,
		resetTokens:  in.ResetTokens,
		verification: in.Verification,
		notifier:     in.Notifier,
		tokens:       in.Tokens,
		profiles:     in.Profiles,
		oauthVerify:  in.OAuthVerify,
		clock:        clk,
	}
}
