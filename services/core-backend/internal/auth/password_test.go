package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_Success(t *testing.T) {
	password := "SecureHackways2026!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected hash success, got error: %v", err)
	}

	if !strings.HasPrefix(hash, "$2a$12$") && !strings.HasPrefix(hash, "$2b$12$") {
		t.Errorf("expected bcrypt cost 12 prefix, got %q", hash[:7])
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("failed to extract bcrypt cost: %v", err)
	}
	if cost < DefaultBcryptCost {
		t.Errorf("expected cost >= %d, got %d", DefaultBcryptCost, cost)
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("expected valid password check to pass")
	}

	if CheckPasswordHash("WrongPassword!", hash) {
		t.Errorf("expected invalid password check to fail")
	}
}

func TestHashPassword_Validations(t *testing.T) {
	_, err := HashPassword("short")
	if err != ErrPasswordTooShort {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}

	longPass := strings.Repeat("A", 73)
	_, err = HashPassword(longPass)
	if err != ErrPasswordTooLong {
		t.Errorf("expected ErrPasswordTooLong, got %v", err)
	}
}
