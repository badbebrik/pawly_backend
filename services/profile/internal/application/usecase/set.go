package usecase

type Set struct {
	CreateProfile        *CreateProfile
	DeleteProfile        *DeleteProfile
	GetProfile           *GetProfile
	UpdateProfileInfo    *UpdateProfileInfo
	UpdatePreferences    *UpdatePreferences
	GetAvatarDownloadURL *GetAvatarDownloadURL
	InitAvatarUpload     *InitAvatarUpload
	ConfirmAvatarUpload  *ConfirmAvatarUpload
	DeleteAvatar         *DeleteAvatar
	GetPreferences       *GetPreferences
	BatchGetPreferences  *BatchGetPreferences
	BatchProfilesBrief   *BatchProfilesBrief
}

func NewSet(in Dependencies) *Set {
	deps := newDependencies(in)
	return &Set{
		CreateProfile:        &CreateProfile{deps: deps},
		DeleteProfile:        &DeleteProfile{deps: deps},
		GetProfile:           &GetProfile{deps: deps},
		UpdateProfileInfo:    &UpdateProfileInfo{deps: deps},
		UpdatePreferences:    &UpdatePreferences{deps: deps},
		GetAvatarDownloadURL: &GetAvatarDownloadURL{deps: deps},
		InitAvatarUpload:     &InitAvatarUpload{deps: deps},
		ConfirmAvatarUpload:  &ConfirmAvatarUpload{deps: deps},
		DeleteAvatar:         &DeleteAvatar{deps: deps},
		GetPreferences:       &GetPreferences{deps: deps},
		BatchGetPreferences:  &BatchGetPreferences{deps: deps},
		BatchProfilesBrief:   &BatchProfilesBrief{deps: deps},
	}
}
