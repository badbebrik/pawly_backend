package usecase

type Set struct {
	RegisterEmail           *RegisterEmail
	ResendEmailVerification *ResendEmailVerification
	VerifyEmail             *VerifyEmail
	LoginEmail              *LoginEmail
	LoginOAuth              *LoginOAuth
	Logout                  *Logout
	LogoutAll               *LogoutAll
	ChangePassword          *ChangePassword
	Refresh                 *Refresh
	PasswordResetRequest    *PasswordResetRequest
	PasswordResetVerify     *PasswordResetVerify
	PasswordResetConfirm    *PasswordResetConfirm
}

func NewSet(in Dependencies) *Set {
	deps := newDependencies(in)

	return &Set{
		RegisterEmail:           &RegisterEmail{deps: deps},
		ResendEmailVerification: &ResendEmailVerification{deps: deps},
		VerifyEmail:             &VerifyEmail{deps: deps},
		LoginEmail:              &LoginEmail{deps: deps},
		LoginOAuth:              &LoginOAuth{deps: deps},
		Logout:                  &Logout{deps: deps},
		LogoutAll:               &LogoutAll{deps: deps},
		ChangePassword:          &ChangePassword{deps: deps},
		Refresh:                 &Refresh{deps: deps},
		PasswordResetRequest:    &PasswordResetRequest{deps: deps},
		PasswordResetVerify:     &PasswordResetVerify{deps: deps},
		PasswordResetConfirm:    &PasswordResetConfirm{deps: deps},
	}
}
