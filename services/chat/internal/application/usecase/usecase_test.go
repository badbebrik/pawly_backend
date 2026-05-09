package usecase

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestChatUsecasesRejectBasicInvalidInput(t *testing.T) {
	ctx := context.Background()
	set, _, _, _, _, _ := baseChatDeps()

	_, err := set.OpenDirectConversation.Execute(ctx, OpenDirectConversationParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.OpenDirectConversation.Execute(ctx, OpenDirectConversationParams{
		CurrentUserID: chatUserID,
		PetID:         chatPetID,
		OtherUserID:   chatUserID,
	})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.ListConversations.Execute(ctx, ListConversationsParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.GetConversation.Execute(ctx, GetConversationParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.GetUnreadSummary.Execute(ctx, GetUnreadSummaryParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.GetMessageHistory.Execute(ctx, GetMessageHistoryParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.SendMessage.Execute(ctx, SendMessageParams{})
	expectChatErr(t, err, ErrInvalidInput)

	_, err = set.MarkRead.Execute(ctx, MarkReadParams{})
	expectChatErr(t, err, ErrInvalidInput)
}

func TestOpenDirectConversationCreatesConversation(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, _, _, _ := baseChatDeps()

	result, err := set.OpenDirectConversation.Execute(ctx, OpenDirectConversationParams{
		CurrentUserID: chatUserID,
		PetID:         chatPetID,
		OtherUserID:   chatOtherUserID,
	})
	if err != nil {
		t.Fatalf("OpenDirectConversation.Execute error: %v", err)
	}

	if result.Conversation.Pet.Name != "Barsik" || result.Conversation.OtherUser.DisplayName == nil || *result.Conversation.OtherUser.DisplayName != "Ivan Ivanov" {
		t.Fatalf("unexpected conversation result: %+v", result.Conversation)
	}
	if len(conversations.created) != 1 {
		t.Fatalf("expected created conversation, got %d", len(conversations.created))
	}
	if len(participants.created) != 2 {
		t.Fatalf("expected two participants, got %d", len(participants.created))
	}
}

func TestGetConversationReturnsDetails(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, _, _, _ := baseChatDeps()
	seedChatConversation(conversations, participants)

	result, err := set.GetConversation.Execute(ctx, GetConversationParams{
		CurrentUserID:  chatUserID,
		ConversationID: chatConversationID,
	})
	if err != nil {
		t.Fatalf("GetConversation.Execute error: %v", err)
	}

	if result.Conversation.ConversationID != chatConversationID || result.Conversation.Pet.Name != "Barsik" {
		t.Fatalf("unexpected conversation: %+v", result.Conversation)
	}
	if !result.Conversation.CanSend {
		t.Fatalf("expected CanSend=true")
	}
}

func TestListConversationsMapsRows(t *testing.T) {
	ctx := context.Background()
	set, conversations, _, _, _, _ := baseChatDeps()
	nextCursor := "next"
	conversations.listResult = ports.ListConversationsResult{
		Items: []ports.ConversationListRow{
			{
				ConversationID: chatConversationID,
				PetID:          chatPetID,
				OtherUserID:    chatOtherUserID,
				UnreadCount:    2,
			},
		},
		NextCursor: &nextCursor,
	}

	result, err := set.ListConversations.Execute(ctx, ListConversationsParams{CurrentUserID: chatUserID, Limit: 20})
	if err != nil {
		t.Fatalf("ListConversations.Execute error: %v", err)
	}

	if len(result.Items) != 1 || result.Items[0].ConversationID != chatConversationID || result.Items[0].UnreadCount != 2 {
		t.Fatalf("unexpected list result: %+v", result)
	}
	if result.NextCursor == nil || *result.NextCursor != "next" {
		t.Fatalf("unexpected next cursor: %+v", result.NextCursor)
	}
}

func TestGetUnreadSummaryReturnsRepositoryData(t *testing.T) {
	ctx := context.Background()
	set, _, _, _, _, _ := baseChatDeps()

	result, err := set.GetUnreadSummary.Execute(ctx, GetUnreadSummaryParams{CurrentUserID: chatUserID})
	if err != nil {
		t.Fatalf("GetUnreadSummary.Execute error: %v", err)
	}

	if result.UnreadConversations != 1 || result.UnreadMessages != 2 {
		t.Fatalf("unexpected summary: %+v", result)
	}
}

func TestGetMessageHistoryRequiresParticipantAndReturnsMessages(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, messages, _, _ := baseChatDeps()
	seedChatConversation(conversations, participants)
	text := "hello"
	messages.history = ports.MessageHistoryPage{
		Messages: []model.Message{
			{
				ID:             chatMessageID,
				ConversationID: chatConversationID,
				SenderUserID:   chatUserID,
				ClientMsgID:    chatClientMsgID,
				Text:           &text,
				CreatedAt:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		HasMore: true,
	}

	result, err := set.GetMessageHistory.Execute(ctx, GetMessageHistoryParams{
		CurrentUserID:  chatUserID,
		ConversationID: chatConversationID,
		Limit:          20,
	})
	if err != nil {
		t.Fatalf("GetMessageHistory.Execute error: %v", err)
	}

	if len(result.Messages) != 1 || result.Messages[0].Text == nil || *result.Messages[0].Text != "hello" || !result.HasMore {
		t.Fatalf("unexpected history: %+v", result)
	}
}

func TestSendMessageCreatesMessageAndPublishesEvent(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, messages, _, realtime := baseChatDeps()
	seedChatConversation(conversations, participants)
	text := "  hello  "

	result, err := set.SendMessage.Execute(ctx, SendMessageParams{
		CurrentUserID:  chatUserID,
		OriginClientID: uuid.MustParse("77777777-7777-7777-7777-777777777777"),
		ConversationID: chatConversationID,
		ClientMsgID:    chatClientMsgID,
		Text:           &text,
	})
	if err != nil {
		t.Fatalf("SendMessage.Execute error: %v", err)
	}

	if result.Message.Text == nil || *result.Message.Text != "hello" {
		t.Fatalf("expected normalized text, got %+v", result.Message.Text)
	}
	if len(messages.created) != 1 {
		t.Fatalf("expected message create, got %d", len(messages.created))
	}
	if len(participants.increments) != 1 || participants.increments[0].userID != chatOtherUserID || participants.increments[0].delta != 1 {
		t.Fatalf("expected unread increment for other user, got %+v", participants.increments)
	}
	if len(conversations.updateLastCalls) != 1 {
		t.Fatalf("expected last message update, got %d", len(conversations.updateLastCalls))
	}
	if len(realtime.messageEvents) != 1 || realtime.messageEvents[0].RecipientUserID != chatOtherUserID {
		t.Fatalf("expected realtime message event, got %+v", realtime.messageEvents)
	}
}

func TestSendMessageReturnsExistingClientMessage(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, messages, _, realtime := baseChatDeps()
	seedChatConversation(conversations, participants)
	text := "hello"
	existing := model.Message{
		ID:             chatMessageID,
		ConversationID: chatConversationID,
		SenderUserID:   chatUserID,
		ClientMsgID:    chatClientMsgID,
		Text:           &text,
		CreatedAt:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	messages.byClientMsg[messageClientKey(chatConversationID, chatUserID, chatClientMsgID)] = existing

	result, err := set.SendMessage.Execute(ctx, SendMessageParams{
		CurrentUserID:  chatUserID,
		ConversationID: chatConversationID,
		ClientMsgID:    chatClientMsgID,
		Text:           &text,
	})
	if err != nil {
		t.Fatalf("SendMessage.Execute error: %v", err)
	}

	if result.Message.MessageID != chatMessageID {
		t.Fatalf("expected existing message, got %+v", result.Message)
	}
	if len(messages.created) != 0 || len(realtime.messageEvents) != 0 {
		t.Fatalf("did not expect create or publish, created=%d events=%d", len(messages.created), len(realtime.messageEvents))
	}
}

func TestMarkReadUpdatesParticipantAndPublishesEvent(t *testing.T) {
	ctx := context.Background()
	set, conversations, participants, messages, _, realtime := baseChatDeps()
	seedChatConversation(conversations, participants)
	messages.byID[chatMessageID] = model.Message{
		ID:             chatMessageID,
		ConversationID: chatConversationID,
		SenderUserID:   chatOtherUserID,
		ClientMsgID:    chatClientMsgID,
		CreatedAt:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	result, err := set.MarkRead.Execute(ctx, MarkReadParams{
		CurrentUserID:     chatUserID,
		ConversationID:    chatConversationID,
		LastReadMessageID: chatMessageID,
	})
	if err != nil {
		t.Fatalf("MarkRead.Execute error: %v", err)
	}

	if result.LastReadMessageID != chatMessageID {
		t.Fatalf("unexpected mark read result: %+v", result)
	}
	if len(participants.markReads) != 1 {
		t.Fatalf("expected mark read call, got %d", len(participants.markReads))
	}
	if len(realtime.readEvents) != 1 || realtime.readEvents[0].LastReadMessageID != chatMessageID {
		t.Fatalf("expected read realtime event, got %+v", realtime.readEvents)
	}
}

func TestChatUsecasesReturnForbiddenForNonParticipants(t *testing.T) {
	ctx := context.Background()
	set, conversations, _, _, _, _ := baseChatDeps()
	conversations.byID[chatConversationID] = model.Conversation{
		ID:         chatConversationID,
		PetID:      chatPetID,
		UserLowID:  chatOtherUserID,
		UserHighID: uuid.MustParse("88888888-8888-8888-8888-888888888888"),
	}

	_, err := set.GetConversation.Execute(ctx, GetConversationParams{
		CurrentUserID:  chatUserID,
		ConversationID: chatConversationID,
	})
	expectChatErr(t, err, ErrForbidden)

	_, err = set.SendMessage.Execute(ctx, SendMessageParams{
		CurrentUserID:  chatUserID,
		ConversationID: chatConversationID,
		ClientMsgID:    chatClientMsgID,
		Text:           chatStringPtr("hello"),
	})
	expectChatErr(t, err, ErrForbidden)
}
