package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash *string
	IsVerified   bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoggedAt time.Time
	DeletedAt    *time.Time
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}
