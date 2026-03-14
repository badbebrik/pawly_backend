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
	LastLoggedAt *time.Time `json:"last_logged_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}

func (u *User) IsPasswordAccount() bool {
	return u.HasPassword()
}

func (u *User) RequireActive() error {
	if !u.IsActive {
		return ErrUserInactive
	}
	return nil
}

func (u *User) RequireVerified() error {
	if !u.IsVerified {
		return ErrUserUnverified
	}
	return nil
}

func (u *User) RequirePasswordAccount() error {
	if !u.IsPasswordAccount() {
		return ErrPasswordAuthUnavailable
	}
	return nil
}

func (u *User) CanLoginWithPassword() error {
	if err := u.RequireActive(); err != nil {
		return err
	}
	if err := u.RequireVerified(); err != nil {
		return err
	}
	if err := u.RequirePasswordAccount(); err != nil {
		return err
	}
	return nil
}

func (u *User) CanChangePassword() error {
	if err := u.RequireActive(); err != nil {
		return err
	}
	if err := u.RequirePasswordAccount(); err != nil {
		return err
	}
	return nil
}

func (u *User) CanRequestPasswordReset() bool {
	return u.IsActive && u.IsVerified
}

func (u *User) MarkVerified() {
	u.IsVerified = true
}
