package usecase

type Set struct {
	RegisterEmail           *RegisterEmailUseCase
	ResendEmailVerification *ResendEmailVerificationUseCase
	VerifyEmail             *VerifyEmailUseCase
	LoginEmail              *LoginEmailUseCase
	LoginOAuth              *LoginOAuthUseCase
	Logout                  *LogoutUseCase
	LogoutAll               *LogoutAllUseCase
	ChangePassword          *ChangePasswordUseCase
	Refresh                 *RefreshUseCase
	PasswordResetRequest    *PasswordResetRequestUseCase
	PasswordResetVerify     *PasswordResetVerifyUseCase
	PasswordResetConfirm    *PasswordResetConfirmUseCase
}

func NewSet(in Dependencies) *Set {
	deps := newDependencies(in)

	return &Set{
		RegisterEmail:           &RegisterEmailUseCase{deps: deps},
		ResendEmailVerification: &ResendEmailVerificationUseCase{deps: deps},
		VerifyEmail:             &VerifyEmailUseCase{deps: deps},
		LoginEmail:              &LoginEmailUseCase{deps: deps},
		LoginOAuth:              &LoginOAuthUseCase{deps: deps},
		Logout:                  &LogoutUseCase{deps: deps},
		LogoutAll:               &LogoutAllUseCase{deps: deps},
		ChangePassword:          &ChangePasswordUseCase{deps: deps},
		Refresh:                 &RefreshUseCase{deps: deps},
		PasswordResetRequest:    &PasswordResetRequestUseCase{deps: deps},
		PasswordResetVerify:     &PasswordResetVerifyUseCase{deps: deps},
		PasswordResetConfirm:    &PasswordResetConfirmUseCase{deps: deps},
	}
}
