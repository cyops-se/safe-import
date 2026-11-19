package db

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// MinPasswordLength is the minimum accepted length of a new password
const MinPasswordLength = 10

var ErrPasswordTooShort = errors.New("password is too short")

// HashPassword returns a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}
	return hashWithoutPolicy(password)
}

// hashWithoutPolicy is used when migrating existing passwords, which may not
// fulfill the current password policy
func hashWithoutPolicy(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword returns true if the password matches the bcrypt hash
func CheckPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// IsPasswordHash returns true if the value looks like a bcrypt hash
func IsPasswordHash(value string) bool {
	if len(value) != 60 {
		return false
	}
	return strings.HasPrefix(value, "$2a$") || strings.HasPrefix(value, "$2b$") || strings.HasPrefix(value, "$2y$")
}
