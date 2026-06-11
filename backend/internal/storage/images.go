// Package storage persists item images on the local filesystem. The items
// table keeps only the relative file path and image metadata; the bytes live
// on disk under a configured directory.
package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	// Register the decoders we support so image.DecodeConfig can read them.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// MaxImageBytes caps the size of an uploaded image. There was previously no
// limit at all, which let arbitrarily large blobs into the database.
const MaxImageBytes = 10 << 20 // 10 MiB

// ErrUnsupportedType is returned when an upload's MIME type is not allowed.
var ErrUnsupportedType = errors.New("unsupported image type")

// ErrTooLarge is returned when an upload exceeds MaxImageBytes.
var ErrTooLarge = errors.New("image exceeds maximum allowed size")

// ErrInvalidImage is returned when the bytes cannot be decoded as an image.
var ErrInvalidImage = errors.New("invalid image content")

// allowedExtensions maps supported MIME types to the on-disk file extension.
var allowedExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
}

// ImageStore writes, reads, and removes image files under a base directory.
type ImageStore struct {
	dir string
}

// New creates the image directory if needed and returns a store rooted at it.
func New(dir string) (*ImageStore, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create images dir: %w", err)
	}
	return &ImageStore{dir: dir}, nil
}

// Save validates the upload, writes it to disk under a unique filename, and
// returns the relative path to persist plus the decoded dimensions.
func (s *ImageStore) Save(itemID uuid.UUID, mime string, data []byte) (path string, width, height int, err error) {
	ext, ok := allowedExtensions[mime]
	if !ok {
		return "", 0, 0, ErrUnsupportedType
	}
	if len(data) > MaxImageBytes {
		return "", 0, 0, ErrTooLarge
	}

	cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(data))
	if decodeErr != nil {
		return "", 0, 0, ErrInvalidImage
	}

	suffix, err := randomSuffix()
	if err != nil {
		return "", 0, 0, err
	}
	name := fmt.Sprintf("%s-%s%s", itemID.String(), suffix, ext)

	if err := os.WriteFile(filepath.Join(s.dir, name), data, 0o640); err != nil {
		return "", 0, 0, fmt.Errorf("write image file: %w", err)
	}

	return name, cfg.Width, cfg.Height, nil
}

// Open returns a reader for the image at the given relative path.
func (s *ImageStore) Open(path string) (io.ReadCloser, error) {
	f, err := os.Open(filepath.Join(s.dir, filepath.Base(path)))
	if err != nil {
		return nil, fmt.Errorf("open image file: %w", err)
	}
	return f, nil
}

// Delete removes the image at the given relative path. A missing file is not
// treated as an error so callers can clean up idempotently.
func (s *ImageStore) Delete(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(filepath.Join(s.dir, filepath.Base(path))); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete image file: %w", err)
	}
	return nil
}

func randomSuffix() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate image filename: %w", err)
	}
	return hex.EncodeToString(b), nil
}
