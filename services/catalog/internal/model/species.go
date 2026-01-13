package model

import "time"

type Species struct {
	ID        int
	NameRu    string
	NameEn    string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
