package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	saveInterval = 5 * time.Second
)

type Checkpoint struct {
	filename    string
	lastSavedAt time.Time

	mu sync.RWMutex
}

func NewCheckpoint(filename string) *Checkpoint {
	return &Checkpoint{
		filename:    filename,
		lastSavedAt: time.Time{},
	}
}

func (s *Checkpoint) Save(q string, ts time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ts.Sub(s.lastSavedAt) < saveInterval {
		return nil
	}

	d := filepath.Dir(s.filename)

	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	if err := os.WriteFile(s.filename, []byte(q), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	s.lastSavedAt = ts

	return nil
}

func (s *Checkpoint) Load() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q, err := os.ReadFile(s.filename)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	return string(q), nil
}
