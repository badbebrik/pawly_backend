package model

import "errors"

var (
	ErrConversationUserPairInvalid = errors.New("conversation_user_pair_invalid")
	ErrMessageTextRequired         = errors.New("message_text_required")
	ErrMessageClientIDRequired     = errors.New("message_client_id_required")
)
