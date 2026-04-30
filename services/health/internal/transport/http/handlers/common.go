package handlers

import (
	"encoding/base64"
	"errors"
	healthports "health/internal/application/ports"
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	appmw "health/internal/transport/http/middleware"
	"net/http"
	"pawly/pkg/httpjson"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	logs           *healthuc.Logs
	scheduled      *healthuc.Scheduled
	vetVisits      *healthuc.VetVisits
	vaccinations   *healthuc.Vaccinations
	procedures     *healthuc.Procedures
	medicalRecords *healthuc.MedicalRecords
	analytics      *healthuc.Analytics
	documents      *healthuc.Documents
	overview       *healthuc.Overview
}

func New(logs *healthuc.Logs, scheduled *healthuc.Scheduled, vetVisits *healthuc.VetVisits, vaccinations *healthuc.Vaccinations, procedures *healthuc.Procedures, medicalRecords *healthuc.MedicalRecords, analytics *healthuc.Analytics, documents *healthuc.Documents, overview *healthuc.Overview) *Handlers {
	return &Handlers{
		logs:           logs,
		scheduled:      scheduled,
		vetVisits:      vetVisits,
		vaccinations:   vaccinations,
		procedures:     procedures,
		medicalRecords: medicalRecords,
		analytics:      analytics,
		documents:      documents,
		overview:       overview,
	}
}

func (h *Handlers) notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not_implemented", "not implemented")
}

func (h *Handlers) getUserAndPet(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return uuid.Nil, uuid.Nil, false
	}
	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid pet_id")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, petID, true
}

func (h *Handlers) getUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return uuid.Nil, false
	}
	return userID, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, healthuc.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid input")
	case errors.Is(err, healthuc.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, healthuc.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, healthuc.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpjson.WriteError(w, status, normalizeErrorCode(code), message)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	httpjson.Write(w, status, body)
}

func normalizeErrorCode(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "invalid_input":
		return "invalid_input"
	case "unauthorized":
		return "unauthorized"
	case "forbidden":
		return "forbidden"
	case "not_found":
		return "not_found"
	case "conflict":
		return "conflict"
	case "internal_error":
		return "internal_error"
	case "not_implemented":
		return "not_implemented"
	default:
		return strings.ToLower(strings.TrimSpace(code))
	}
}

func parseMetricValues(in []dto.MetricValueRequest) ([]healthuc.CreateOrUpdateMetricValue, error) {
	out := make([]healthuc.CreateOrUpdateMetricValue, 0, len(in))
	for i := range in {
		id, err := uuid.Parse(in[i].MetricID)
		if err != nil {
			return nil, err
		}
		out = append(out, healthuc.CreateOrUpdateMetricValue{
			MetricID: id,
			ValueNum: in[i].ValueNum,
		})
	}
	return out, nil
}

func parseLogTypeRequirements(in []dto.LogTypeMetricRequirementRequest) ([]healthuc.LogTypeRequirementInput, error) {
	out := make([]healthuc.LogTypeRequirementInput, 0, len(in))
	for i := range in {
		id, err := uuid.Parse(in[i].MetricID)
		if err != nil {
			return nil, err
		}
		out = append(out, healthuc.LogTypeRequirementInput{
			MetricID:   id,
			IsRequired: in[i].IsRequired,
		})
	}
	return out, nil
}

func parseAttachmentParams(attachments []dto.AttachmentRequest) ([]healthuc.AttachmentParam, error) {
	out := make([]healthuc.AttachmentParam, 0, len(attachments))
	for i := range attachments {
		id, err := uuid.Parse(strings.TrimSpace(attachments[i].FileID))
		if err != nil {
			return nil, err
		}
		out = append(out, healthuc.AttachmentParam{
			FileID:   id,
			FileName: attachments[i].FileName,
		})
	}
	return out, nil
}

func parseHealthDictionaryItemRefs(in []dto.HealthDictionaryItemRefRequest) ([]healthuc.HealthDictionaryItemRefParam, error) {
	out := make([]healthuc.HealthDictionaryItemRefParam, 0, len(in))
	for i := range in {
		var id *uuid.UUID
		if in[i].ID != nil {
			parsed, err := uuid.Parse(*in[i].ID)
			if err != nil {
				return nil, err
			}
			id = &parsed
		}
		out = append(out, healthuc.HealthDictionaryItemRefParam{
			ID:   id,
			Name: in[i].Name,
		})
	}
	return out, nil
}

func optionalQueryString(r *http.Request, key string) *string {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil
	}
	return &raw
}

func optionalUUIDFromString(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseOptionalBool(q map[string][]string, key string) (*bool, error) {
	vals, ok := q[key]
	if !ok || len(vals) == 0 {
		return nil, nil
	}
	v, err := strconv.ParseBool(vals[0])
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func parseUUIDCSVOrMulti(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return []uuid.UUID{}, nil
	}
	pieces := make([]string, 0, len(raw))
	for i := range raw {
		for _, p := range strings.Split(raw[i], ",") {
			if strings.TrimSpace(p) != "" {
				pieces = append(pieces, strings.TrimSpace(p))
			}
		}
	}
	out := make([]uuid.UUID, 0, len(pieces))
	for i := range pieces {
		id, err := uuid.Parse(pieces[i])
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func parseOptionalDateTime(raw string, endOfDay bool) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if strings.Contains(trimmed, "T") {
		t, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t, nil
}

func encodeLogCursor(cur *healthports.LogCursor) *string {
	if cur == nil {
		return nil
	}
	raw := cur.OccurredAt.UTC().Format(time.RFC3339Nano) + "|" + cur.ID.String()
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	return &encoded
}

func decodeLogCursor(raw string) (*healthports.LogCursor, error) {
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
		return nil, errors.New("bad cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}
	return &healthports.LogCursor{OccurredAt: ts, ID: id}, nil
}

func parseIntOrDefault(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
