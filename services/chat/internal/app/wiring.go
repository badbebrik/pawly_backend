package app

import "chat/internal/application/usecase"

func (a *App) wire() error {
	a.useCases = &UseCases{
		OpenDirectConversation: usecase.NewOpenDirectConversation(nil, nil, nil, nil),
		ListConversations:      usecase.NewListConversations(nil, nil, nil),
		GetConversation:        usecase.NewGetConversation(nil, nil, nil, nil),
		GetUnreadSummary:       usecase.NewGetUnreadSummary(nil),
		GetMessageHistory:      usecase.NewGetMessageHistory(nil, nil, nil),
		SendMessage:            usecase.NewSendMessage(nil, nil, nil, nil, nil, nil),
		MarkRead:               usecase.NewMarkRead(nil, nil, nil),
	}

	return nil
}
