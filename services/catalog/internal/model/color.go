package model

import (
	"time"

	"github.com/google/uuid"
)

type Color struct {
	ID        uuid.UUID
	NameRu    string
	NameEn    string
	Hex       string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
