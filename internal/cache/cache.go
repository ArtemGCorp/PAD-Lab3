// Package cache stores the last successful rate table as JSON.
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"pad-lab3/internal/bnm"
)

// ErrNoCache — no usable cache (missing or corrupted).
var ErrNoCache = errors.New("нет сохранённых курсов")

// Entry is what is stored on disk.
type Entry struct {
	FetchedAt time.Time `json:"fetched_at"`
	Table     bnm.Table `json:"table"`
}

// Store is a cache file.
type Store struct{ Path string }

// DefaultPath is <UserCacheDir>/pad-lab3/rates.json (falls back to ./rates-cache.json).
func DefaultPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "rates-cache.json"
	}
	return filepath.Join(dir, "pad-lab3", "rates.json")
}

// Save writes the entry atomically (tmp file + rename).
func (s Store) Save(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("кэш: %w", err)
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("кэш: %w", err)
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("кэш: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("кэш: %w", err)
	}
	return nil
}

// Load reads the entry. Missing or corrupted files return ErrNoCache.
func (s Store) Load() (Entry, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return Entry{}, ErrNoCache
	}
	if err != nil {
		return Entry{}, fmt.Errorf("%w: %v", ErrNoCache, err)
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil || len(e.Table.Rates) == 0 {
		return Entry{}, ErrNoCache
	}
	return e, nil
}
