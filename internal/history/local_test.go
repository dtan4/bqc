package history

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"strconv"
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

	count := 0

	s := &LocalStorage{
		db:     db,
		bucket: []byte(bucket),
		tsFunc: func() time.Time {
			count += 1

			return time.Date(2023, 5, 24, 12, 34, 56, count, time.UTC)
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

// TestLocalStorageCompatibility checks whether the history file with serialized
// Go objects can read with the current setup (i.e. the latest dependencies) or
// not.
func TestLocalStorageCompatibility(t *testing.T) {
	t.Parallel()

	db, err := bolt.Open(filepath.Join("testdata", "test.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	bucket := "test-bucket"

	want := []*bigquery.Result{
		{
			Query:               "select 1",
			TotalBytesProcessed: 0,
			DryRun:              true,
			EndTime:             time.Date(2023, 5, 24, 12, 34, 56, 0, time.UTC),
		},
		{
			Query:               "select 1;",
			TotalBytesProcessed: 12345,
			DryRun:              false,
			EndTime:             time.Date(2023, 5, 25, 13, 24, 59, 0, time.UTC),
		},
	}

	if os.Getenv("UPDATE") == "1" {
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			return err
		}); err != nil {
			t.Fatal(err)
		}

		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))

			for i, r := range want {
				var bb bytes.Buffer

				if err := gob.NewEncoder(&bb).Encode(r); err != nil {
					return err
				}

				if err := b.Put([]byte(strconv.Itoa(i)), bb.Bytes()); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	got := []*bigquery.Result{}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r bigquery.Result

			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&r); err != nil {
				return err
			}

			got = append(got, &r)
		}

		return nil
	})
	if err != nil {
		t.Errorf("want no error, got: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("data mismatch (-want +got):\n%s", diff)
	}
}
