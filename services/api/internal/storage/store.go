package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// ErrInvalidFile indicates an invalid or unsupported file payload.
var ErrInvalidFile = errors.New("invalid or empty file payload")

// Store defines the pluggable object/asset storage interface.
type Store interface {
	// Upload stores an asset and returns its public retrieval URL or local path.
	Upload(ctx context.Context, filename string, r io.Reader, sizeBytes int64, contentType string) (string, error)
	// Delete removes a stored asset.
	Delete(ctx context.Context, filename string) error
	// URL returns the public URL for an asset key.
	URL(filename string) string
}

// LocalStore implements Store on the local filesystem for zero-dependency development.
type LocalStore struct {
	baseDir string
	baseURL string
}

// NewLocalStore initializes a local filesystem store, creating baseDir if necessary.
func NewLocalStore(baseDir, baseURL string) (*LocalStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStore{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

// Upload writes the file stream to the local directory.
func (s *LocalStore) Upload(ctx context.Context, filename string, r io.Reader, sizeBytes int64, contentType string) (string, error) {
	if r == nil {
		return "", ErrInvalidFile
	}

	targetPath := filepath.Join(s.baseDir, filename)
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}

	return s.URL(filename), nil
}

// Delete removes the file from local storage.
func (s *LocalStore) Delete(ctx context.Context, filename string) error {
	targetPath := filepath.Join(s.baseDir, filename)
	err := os.Remove(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// URL constructs a local access URL.
func (s *LocalStore) URL(filename string) string {
	return s.baseURL + "/uploads/" + filename
}
