package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash *string    `json:"password_hash"`
	IsVerified   bool       `json:"is_verified"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoggedAt time.Time  `json:"last_logged_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}
