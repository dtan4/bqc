package history

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	bolt "go.etcd.io/bbolt"

	"github.com/dtan4/bqc/internal/bigquery"
)

func TestLocalStorageAppendAndList(t *testing.T) {
	t.Parallel()

	filename := filepath.Join(t.TempDir(), "test_append.db")
	bucket := "test-bucket"

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		t.Fatal(err)
	}

	s := &LocalStorage{
		db:     db,
		bucket: []byte(bucket),
		tsFunc: func() time.Time {
			return time.Now()
		},
	}

	results := []*bigquery.Result{
		{
			Keys:                []string{"foo", "bar"},
			TotalBytesProcessed: 12345,
			EndTime:             time.Date(2023, 5, 24, 12, 34, 56, 0, time.UTC),
		},
		{
			Keys:                []string{"baz", "qux"},
			TotalBytesProcessed: 12345,
			EndTime:             time.Date(2023, 5, 25, 13, 24, 59, 0, time.UTC),
		},
	}

	for _, r := range results {
		if err := s.Append(r); err != nil {
			t.Errorf("(Append) want no error, got: %s", err)
		}
	}

	got, err := s.List()
	if err != nil {
		t.Errorf("(List) want no error, got: %s", err)
	}

	if diff := cmp.Diff(results, got); diff != "" {
		t.Errorf("data mismatch (-want +got):\n%s", diff)
	}
}
