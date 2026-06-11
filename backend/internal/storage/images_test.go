package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	"github.com/google/uuid"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestSaveOpenDelete(t *testing.T) {
	t.Parallel()

	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	id := uuid.New()
	data := pngBytes(t, 12, 8)

	path, width, height, err := st.Save(id, "image/png", data)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if width != 12 || height != 8 {
		t.Fatalf("expected dimensions 12x8, got %dx%d", width, height)
	}

	rc, err := st.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("read bytes differ from written bytes")
	}

	if err := st.Delete(path); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := st.Open(path); err == nil {
		t.Fatalf("expected open to fail after delete")
	}
	// Deleting a missing file is a no-op.
	if err := st.Delete(path); err != nil {
		t.Fatalf("delete missing: %v", err)
	}
}

func TestSaveRejectsUnsupportedType(t *testing.T) {
	t.Parallel()

	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if _, _, _, err := st.Save(uuid.New(), "image/webp", pngBytes(t, 2, 2)); err != ErrUnsupportedType {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestSaveRejectsInvalidImage(t *testing.T) {
	t.Parallel()

	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if _, _, _, err := st.Save(uuid.New(), "image/png", []byte("not a png")); err != ErrInvalidImage {
		t.Fatalf("expected ErrInvalidImage, got %v", err)
	}
}

func TestSaveRejectsTooLarge(t *testing.T) {
	t.Parallel()

	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	oversized := make([]byte, MaxImageBytes+1)
	if _, _, _, err := st.Save(uuid.New(), "image/png", oversized); err != ErrTooLarge {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}
