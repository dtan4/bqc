package history

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
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

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)

		return b.Put(s.keyFromTimestamp(s.tsFunc()), v.Bytes())
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

			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&r); err != nil {
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

// Use milliseconds as key
func (s *LocalStorage) keyFromTimestamp(ts time.Time) []byte {
	b := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(b, uint64(ts.UnixMilli()))

	return b
}