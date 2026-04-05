package usecase

type Set struct {
	CreateProfile        *CreateProfileUseCase
	DeleteProfile        *DeleteProfileUseCase
	GetProfile           *GetProfileUseCase
	UpdateProfileInfo    *UpdateProfileInfoUseCase
	UpdatePreferences    *UpdatePreferencesUseCase
	GetAvatarDownloadURL *GetAvatarDownloadURLUseCase
	InitAvatarUpload     *InitAvatarUploadUseCase
	ConfirmAvatarUpload  *ConfirmAvatarUploadUseCase
	DeleteAvatar         *DeleteAvatarUseCase
	GetPreferences       *GetPreferencesUseCase
	BatchGetPreferences  *BatchGetPreferencesUseCase
	BatchProfilesBrief   *BatchProfilesBriefUseCase
}

func NewSet(in Dependencies) *Set {
	deps := newDependencies(in)
	return &Set{
		CreateProfile:        &CreateProfileUseCase{deps: deps},
		DeleteProfile:        &DeleteProfileUseCase{deps: deps},
		GetProfile:           &GetProfileUseCase{deps: deps},
		UpdateProfileInfo:    &UpdateProfileInfoUseCase{deps: deps},
		UpdatePreferences:    &UpdatePreferencesUseCase{deps: deps},
		GetAvatarDownloadURL: &GetAvatarDownloadURLUseCase{deps: deps},
		InitAvatarUpload:     &InitAvatarUploadUseCase{deps: deps},
		ConfirmAvatarUpload:  &ConfirmAvatarUploadUseCase{deps: deps},
		DeleteAvatar:         &DeleteAvatarUseCase{deps: deps},
		GetPreferences:       &GetPreferencesUseCase{deps: deps},
		BatchGetPreferences:  &BatchGetPreferencesUseCase{deps: deps},
		BatchProfilesBrief:   &BatchProfilesBriefUseCase{deps: deps},
	}
}
