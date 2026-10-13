package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultBcryptCost is the mandatory minimum cost factor enforced across Hackways.
const DefaultBcryptCost = 12

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong  = errors.New("password exceeds maximum allowed length of 72 bytes")
)

// HashPassword hashes a plaintext password using bcrypt with cost >= 12.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	// Bcrypt has a hard limit of 72 bytes
	if len([]byte(password)) > 72 {
		return "", ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash in constant time.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
