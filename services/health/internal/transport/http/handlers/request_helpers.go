package handlers

import (
	"encoding/base64"
	healthports "health/internal/application/ports"
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	"net/http"
	"pawly/pkg/httpjson"
	"strings"
	"time"

	"github.com/google/uuid"
)

func decodeBody(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := httpjson.Decode(r, out); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return false
	}
	return true
}

func parseOptionalRFC3339(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func encodeTimeCursor(cur *healthports.TimeCursor) *string {
	if cur == nil {
		return nil
	}
	raw := cur.SortAt.UTC().Format(time.RFC3339Nano) + "|" + cur.ID.String()
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	return &encoded
}

func decodeTimeCursor(raw string) (*healthports.TimeCursor, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, healthuc.ErrInvalidInput
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}
	return &healthports.TimeCursor{SortAt: ts, ID: id}, nil
}

func parseBoolOrDefault(raw string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
}

func parseScheduledItemRequest(w http.ResponseWriter, userID, petID uuid.UUID, req dto.CreateOrUpdateScheduledItemRequest) (healthuc.CreateScheduledItemParams, bool) {
	startsAt, err := parseOptionalRFC3339(req.StartsAt)
	if err != nil || startsAt == nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid starts_at")
		return healthuc.CreateScheduledItemParams{}, false
	}
	var sourceID *uuid.UUID
	if req.SourceID != nil && strings.TrimSpace(*req.SourceID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.SourceID))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid source_id")
			return healthuc.CreateScheduledItemParams{}, false
		}
		sourceID = &parsed
	}
	var recurrenceRule *string
	var recurrenceInterval *int
	var recurrenceUntil *time.Time
	if req.Recurrence != nil {
		recurrenceRule = req.Recurrence.Rule
		recurrenceInterval = req.Recurrence.Interval
		recurrenceUntil, err = parseOptionalRFC3339(req.Recurrence.Until)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid recurrence.until")
			return healthuc.CreateScheduledItemParams{}, false
		}
	}
	return healthuc.CreateScheduledItemParams{
		UserID:              userID,
		PetID:               petID,
		SourceType:          req.SourceType,
		SourceID:            sourceID,
		Title:               req.Title,
		Note:                req.Note,
		StartsAt:            *startsAt,
		PushEnabled:         req.PushEnabled,
		RemindOffsetMinutes: req.RemindOffsetMinutes,
		RecurrenceRule:      recurrenceRule,
		RecurrenceInterval:  recurrenceInterval,
		RecurrenceUntil:     recurrenceUntil,
	}, true
}

func encodeScheduledTimeCursor(cur *healthports.TimeCursor) *string {
	if cur == nil {
		return nil
	}
	raw := cur.SortAt.UTC().Format(time.RFC3339Nano) + "|" + cur.ID.String()
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	return &encoded
}

func decodeScheduledTimeCursor(raw string) (*healthports.TimeCursor, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, healthuc.ErrInvalidInput
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}
	return &healthports.TimeCursor{SortAt: ts, ID: id}, nil
}

func medicalEntityReminderToUsecaseParams(req *dto.MedicalEntityReminderRequest) *healthuc.MedicalEntityReminderParams {
	if req == nil {
		return nil
	}
	return &healthuc.MedicalEntityReminderParams{
		PushEnabled:         req.PushEnabled,
		RemindOffsetMinutes: req.RemindOffsetMinutes,
	}
}
