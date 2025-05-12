package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSave(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "checkpoint.ckpt")

	ckpt := &Checkpoint{
		filename:    filename,
		lastSavedAt: time.Time{},
	}

	q := "q1"
	ts := time.Date(2023, 5, 21, 19, 0, 0, 0, time.UTC)

	if err := ckpt.Save(q, ts); err != nil {
		t.Errorf("want no error, got: %s", err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("file does not exist: %s", err)
	}

	if diff := cmp.Diff(q, string(got)); diff != "" {
		t.Errorf("checkpoint body mismatch (-want +got):\n%s", diff)
	}
}

func TestSave_createDir(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "directory", "does", "not", "exist", "checkpoint.ckpt")

	ckpt := &Checkpoint{
		filename:    filename,
		lastSavedAt: time.Time{},
	}

	q := "q1"
	ts := time.Date(2023, 5, 21, 19, 0, 0, 0, time.UTC)

	if err := ckpt.Save(q, ts); err != nil {
		t.Errorf("want no error, got: %s", err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("file does not exist: %s", err)
	}

	if diff := cmp.Diff(q, string(got)); diff != "" {
		t.Errorf("checkpoint body mismatch (-want +got):\n%s", diff)
	}
}

func TestSave_savedAtTooClose(t *testing.T) {
	t.Parallel()

	d, err := os.MkdirTemp("", "testSave_savedAtTooClose")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	filename := filepath.Join(d, "checkpoint.ckpt")
	ts := time.Date(2023, 5, 21, 19, 0, 0, 0, time.UTC)
	lastSavedAt := ts.Add(-1 * time.Second)

	ckpt := &Checkpoint{
		filename:    filename,
		lastSavedAt: lastSavedAt,
	}

	q := "q1"

	if err := ckpt.Save(q, ts); err != nil {
		t.Errorf("want no error, got: %s", err)
	}

	if diff := cmp.Diff(lastSavedAt, ckpt.lastSavedAt); diff != "" {
		t.Errorf("lastSavedAt mismatch (-want +got):\n%s", diff)
	}
}
