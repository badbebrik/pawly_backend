package usecase

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	chatUserID         = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chatOtherUserID    = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	chatPetID          = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	chatConversationID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	chatMessageID      = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	chatClientMsgID    = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

func baseChatDeps() (*Set, *stubChatConversations, *stubChatParticipants, *stubChatMessages, *stubChatACL, *stubChatRealtime) {
	conversations := &stubChatConversations{
		byID:    map[uuid.UUID]model.Conversation{},
		direct:  map[string]model.Conversation{},
		summary: ports.UnreadSummary{UnreadConversations: 1, UnreadMessages: 2},
	}
	participants := &stubChatParticipants{
		byConversationUser: map[string]model.ConversationParticipant{},
	}
	messages := &stubChatMessages{
		byID:        map[uuid.UUID]model.Message{},
		byClientMsg: map[string]model.Message{},
	}
	acl := &stubChatACL{active: map[uuid.UUID]map[uuid.UUID]bool{}}
	acl.setActive(chatPetID, chatUserID, true)
	acl.setActive(chatPetID, chatOtherUserID, true)
	profiles := &stubChatProfiles{briefs: map[uuid.UUID]ports.ProfileBrief{
		chatOtherUserID: {UserID: chatOtherUserID, DisplayName: chatStringPtr("Ivan Ivanov")},
	}}
	pets := &stubChatPets{briefs: map[uuid.UUID]ports.PetBrief{
		chatPetID: {PetID: chatPetID, Name: "Barsik"},
	}}
	realtime := &stubChatRealtime{}
	set := NewSet(Dependencies{
		Conversations: conversations,
		Participants:  participants,
		Messages:      messages,
		TxManager:     stubChatTx{},
		ACLClient:     acl,
		ProfileClient: profiles,
		PetClient:     pets,
		Presence:      &stubChatPresence{},
		Realtime:      realtime,
	})
	return set, conversations, participants, messages, acl, realtime
}

func seedChatConversation(conversations *stubChatConversations, participants *stubChatParticipants) model.Conversation {
	conversation := model.Conversation{
		ID:         chatConversationID,
		PetID:      chatPetID,
		UserLowID:  chatUserID,
		UserHighID: chatOtherUserID,
		CreatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	conversations.byID[conversation.ID] = conversation
	conversations.direct[directKey(conversation.PetID, conversation.UserLowID, conversation.UserHighID)] = conversation
	participants.byConversationUser[participantKey(conversation.ID, chatUserID)] = model.ConversationParticipant{
		ConversationID: conversation.ID,
		UserID:         chatUserID,
		UnreadCount:    1,
	}
	participants.byConversationUser[participantKey(conversation.ID, chatOtherUserID)] = model.ConversationParticipant{
		ConversationID: conversation.ID,
		UserID:         chatOtherUserID,
	}
	return conversation
}

func expectChatErr(t *testing.T, got error, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("expected error %v, got %v", want, got)
	}
}

func chatStringPtr(v string) *string {
	return &v
}

type stubChatConversations struct {
	byID   map[uuid.UUID]model.Conversation
	direct map[string]model.Conversation

	listResult ports.ListConversationsResult
	summary    ports.UnreadSummary

	createErr       error
	getErr          error
	getDirectErr    error
	listErr         error
	summaryErr      error
	updateLastErr   error
	created         []model.Conversation
	updateLastCalls []chatUpdateLastCall
}

type chatUpdateLastCall struct {
	conversationID uuid.UUID
	messageID      uuid.UUID
	senderUserID   uuid.UUID
	preview        *string
}

func (r *stubChatConversations) Create(_ context.Context, conversation *model.Conversation) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[conversation.ID] = *conversation
	r.direct[directKey(conversation.PetID, conversation.UserLowID, conversation.UserHighID)] = *conversation
	r.created = append(r.created, *conversation)
	return nil
}

func (r *stubChatConversations) GetByID(_ context.Context, conversationID uuid.UUID) (*model.Conversation, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	conversation, ok := r.byID[conversationID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return &conversation, nil
}

func (r *stubChatConversations) GetDirectByPetAndUsers(_ context.Context, petID, userLowID, userHighID uuid.UUID) (*model.Conversation, error) {
	if r.getDirectErr != nil {
		return nil, r.getDirectErr
	}
	conversation, ok := r.direct[directKey(petID, userLowID, userHighID)]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return &conversation, nil
}

func (r *stubChatConversations) ListConversations(_ context.Context, _ uuid.UUID, _ ports.ListConversationsParams) (ports.ListConversationsResult, error) {
	if r.listErr != nil {
		return ports.ListConversationsResult{}, r.listErr
	}
	return r.listResult, nil
}

func (r *stubChatConversations) GetUnreadSummary(_ context.Context, _ uuid.UUID) (ports.UnreadSummary, error) {
	if r.summaryErr != nil {
		return ports.UnreadSummary{}, r.summaryErr
	}
	return r.summary, nil
}

func (r *stubChatConversations) UpdateLastMessage(_ context.Context, conversationID, messageID, senderUserID uuid.UUID, preview *string, _ time.Time) error {
	if r.updateLastErr != nil {
		return r.updateLastErr
	}
	r.updateLastCalls = append(r.updateLastCalls, chatUpdateLastCall{
		conversationID: conversationID,
		messageID:      messageID,
		senderUserID:   senderUserID,
		preview:        preview,
	})
	return nil
}

type stubChatParticipants struct {
	byConversationUser map[string]model.ConversationParticipant
	createErr          error
	getErr             error
	markReadErr        error
	incrementErr       error
	created            []model.ConversationParticipant
	increments         []chatIncrementCall
	markReads          []chatMarkReadCall
}

type chatIncrementCall struct {
	conversationID uuid.UUID
	userID         uuid.UUID
	delta          int
}

type chatMarkReadCall struct {
	conversationID    uuid.UUID
	userID            uuid.UUID
	lastReadMessageID uuid.UUID
}

func (r *stubChatParticipants) CreateBatch(_ context.Context, participants []model.ConversationParticipant) error {
	if r.createErr != nil {
		return r.createErr
	}
	for _, participant := range participants {
		r.byConversationUser[participantKey(participant.ConversationID, participant.UserID)] = participant
		r.created = append(r.created, participant)
	}
	return nil
}

func (r *stubChatParticipants) GetByConversationAndUser(_ context.Context, conversationID, userID uuid.UUID) (*model.ConversationParticipant, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	participant, ok := r.byConversationUser[participantKey(conversationID, userID)]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return &participant, nil
}

func (r *stubChatParticipants) ListByConversation(_ context.Context, _ uuid.UUID) ([]model.ConversationParticipant, error) {
	return nil, nil
}

func (r *stubChatParticipants) IncrementUnread(_ context.Context, conversationID, userID uuid.UUID, delta int) error {
	if r.incrementErr != nil {
		return r.incrementErr
	}
	r.increments = append(r.increments, chatIncrementCall{conversationID: conversationID, userID: userID, delta: delta})
	return nil
}

func (r *stubChatParticipants) MarkRead(_ context.Context, conversationID, userID, lastReadMessageID uuid.UUID, _ time.Time) (*model.ConversationParticipant, error) {
	if r.markReadErr != nil {
		return nil, r.markReadErr
	}
	participant := r.byConversationUser[participantKey(conversationID, userID)]
	participant.LastReadMessageID = &lastReadMessageID
	participant.UnreadCount = 0
	r.byConversationUser[participantKey(conversationID, userID)] = participant
	r.markReads = append(r.markReads, chatMarkReadCall{conversationID: conversationID, userID: userID, lastReadMessageID: lastReadMessageID})
	return &participant, nil
}

type stubChatMessages struct {
	byID        map[uuid.UUID]model.Message
	byClientMsg map[string]model.Message
	history     ports.MessageHistoryPage
	createErr   error
	getErr      error
	findErr     error
	listErr     error
	created     []model.Message
}

func (r *stubChatMessages) Create(_ context.Context, message *model.Message) error {
	if r.createErr != nil {
		return r.createErr
	}
	if message.ID == uuid.Nil {
		message.ID = chatMessageID
	}
	r.byID[message.ID] = *message
	r.byClientMsg[messageClientKey(message.ConversationID, message.SenderUserID, message.ClientMsgID)] = *message
	r.created = append(r.created, *message)
	return nil
}

func (r *stubChatMessages) GetByID(_ context.Context, messageID uuid.UUID) (*model.Message, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	message, ok := r.byID[messageID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return &message, nil
}

func (r *stubChatMessages) FindByClientMsgID(_ context.Context, conversationID, senderUserID, clientMsgID uuid.UUID) (*model.Message, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	message, ok := r.byClientMsg[messageClientKey(conversationID, senderUserID, clientMsgID)]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return &message, nil
}

func (r *stubChatMessages) ListHistory(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ int) (ports.MessageHistoryPage, error) {
	if r.listErr != nil {
		return ports.MessageHistoryPage{}, r.listErr
	}
	return r.history, nil
}

type stubChatTx struct{}

func (stubChatTx) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type stubChatACL struct {
	active map[uuid.UUID]map[uuid.UUID]bool
	err    error
}

func (c *stubChatACL) setActive(petID, userID uuid.UUID, active bool) {
	if c.active[petID] == nil {
		c.active[petID] = map[uuid.UUID]bool{}
	}
	c.active[petID][userID] = active
}

func (c *stubChatACL) IsActiveMember(_ context.Context, petID, userID uuid.UUID) (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	return c.active[petID][userID], nil
}

type stubChatProfiles struct {
	briefs map[uuid.UUID]ports.ProfileBrief
	err    error
}

func (c *stubChatProfiles) BatchGetBrief(_ context.Context, userIDs []uuid.UUID) (map[uuid.UUID]ports.ProfileBrief, error) {
	if c.err != nil {
		return nil, c.err
	}
	out := make(map[uuid.UUID]ports.ProfileBrief, len(userIDs))
	for _, id := range userIDs {
		if brief, ok := c.briefs[id]; ok {
			out[id] = brief
		}
	}
	return out, nil
}

type stubChatPets struct {
	briefs map[uuid.UUID]ports.PetBrief
	err    error
}

func (c *stubChatPets) BatchGetBrief(_ context.Context, petIDs []uuid.UUID) (map[uuid.UUID]ports.PetBrief, error) {
	if c.err != nil {
		return nil, c.err
	}
	out := make(map[uuid.UUID]ports.PetBrief, len(petIDs))
	for _, id := range petIDs {
		if brief, ok := c.briefs[id]; ok {
			out[id] = brief
		}
	}
	return out, nil
}

type stubChatPresence struct {
	inChat bool
	err    error
}

func (p *stubChatPresence) IsUserInConversation(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return p.inChat, p.err
}

type stubChatRealtime struct {
	messageEvents []ports.MessageSentEvent
	readEvents    []ports.ReadUpdatedEvent
}

func (r *stubChatRealtime) PublishMessageSent(_ context.Context, event ports.MessageSentEvent) error {
	r.messageEvents = append(r.messageEvents, event)
	return nil
}

func (r *stubChatRealtime) PublishReadUpdated(_ context.Context, event ports.ReadUpdatedEvent) error {
	r.readEvents = append(r.readEvents, event)
	return nil
}

func directKey(petID, lowID, highID uuid.UUID) string {
	return petID.String() + ":" + lowID.String() + ":" + highID.String()
}

func participantKey(conversationID, userID uuid.UUID) string {
	return conversationID.String() + ":" + userID.String()
}

func messageClientKey(conversationID, userID, clientMsgID uuid.UUID) string {
	return conversationID.String() + ":" + userID.String() + ":" + clientMsgID.String()
}

var _ ports.ConversationRepository = (*stubChatConversations)(nil)
var _ ports.ParticipantRepository = (*stubChatParticipants)(nil)
var _ ports.MessageRepository = (*stubChatMessages)(nil)
var _ ports.TxManager = stubChatTx{}
var _ ports.ACLClient = (*stubChatACL)(nil)
var _ ports.ProfileClient = (*stubChatProfiles)(nil)
var _ ports.PetClient = (*stubChatPets)(nil)
var _ ports.RealtimePublisher = (*stubChatRealtime)(nil)
