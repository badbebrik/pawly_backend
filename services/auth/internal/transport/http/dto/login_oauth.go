package dto

type LoginOAuthRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
	Locale   string `json:"locale"`
	Timezone string `json:"time_zone"`
}
