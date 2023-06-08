package history

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/dtan4/bqc/internal/bigquery"
)

type LocalStorage struct {
	db     *bolt.DB
	bucket []byte
	tsFunc func() time.Time
}

var _ Storage = (*LocalStorage)(nil)

func NewLocalStorage(filename, bucket string) (*LocalStorage, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))

		return err
	}); err != nil {
		return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
	}

	return &LocalStorage{
		db:     db,
		bucket: []byte(bucket),
		tsFunc: time.Now,
	}, nil
}

func (s *LocalStorage) Close() error {
	return s.db.Close()
}

func (s *LocalStorage) Append(result *bigquery.Result) error {
	var v bytes.Buffer

	if err := gob.NewEncoder(&v).Encode(*result); err != nil {
		return fmt.Errorf("encode result to gob: %w", err)
	}

	c, err := compressZstd(v.Bytes())
	if err != nil {
		return fmt.Errorf("compress history with zstd: %w", err)
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)

		return b.Put(s.keyFromTimestamp(s.tsFunc()), c)
	})
	if err != nil {
		return fmt.Errorf("append: %w", err)
	}

	return nil
}

func (s *LocalStorage) List() ([]*bigquery.Result, error) {
	results := []*bigquery.Result{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r bigquery.Result

			uv, err := decompressZstd(v)
			if err != nil {
				return fmt.Errorf("decompress history from zstd: %w", err)
			}

			if err := gob.NewDecoder(bytes.NewBuffer(uv)).Decode(&r); err != nil {
				return fmt.Errorf("decode result from gob: %w", err)
			}

			results = append(results, &r)
		}

		return nil
	})
	if err != nil {
		return []*bigquery.Result{}, fmt.Errorf("view: %w", err)
	}

	return results, nil
}

// Use nanoseconds as key
func (s *LocalStorage) keyFromTimestamp(ts time.Time) []byte {
	return []byte(strconv.FormatInt(ts.UnixNano(), 10))
}
