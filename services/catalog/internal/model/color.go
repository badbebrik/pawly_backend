package model

import "time"

type Color struct {
	ID        int
	NameRu    string
	NameEn    string
	Hex       string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
