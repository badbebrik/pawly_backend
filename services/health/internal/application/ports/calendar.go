package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type CalendarRepository interface {
	ListCalendarDayMedicalFacts(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.CalendarDayItem, error)
}
