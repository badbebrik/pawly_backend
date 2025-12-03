package util

import (
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidatePassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false

	for _, c := range p {
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
			hasLetter = true
		case '0' <= c && c <= '9':
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
