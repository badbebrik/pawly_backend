package util

import (
	"errors"
	"regexp"
	"strconv"
)

func AtoiPositive(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, errors.New("invalid")
	}
	return n, nil
}

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func IsHexColor(s string) bool {
	return hexColorRe.MatchString(s)
}
