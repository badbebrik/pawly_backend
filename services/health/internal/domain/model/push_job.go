package model

type ScheduledOccurrencePushJob struct {
	Event           string   `json:"event"`
	PetID           string   `json:"pet_id"`
	OccurrenceID    string   `json:"occurrence_id"`
	ScheduledItemID string   `json:"scheduled_item_id"`
	UserIDs         []string `json:"user_ids"`
	SourceType      string   `json:"source_type"`
	Title           string   `json:"title"`
	Note            string   `json:"note,omitempty"`
	ScheduledFor    string   `json:"scheduled_for"`
}
