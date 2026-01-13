package model

import "time"

type Pattern struct {
	ID        int
	NameRu    string
	NameEn    string
	IconKey   string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
