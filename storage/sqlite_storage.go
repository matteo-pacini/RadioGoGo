// Copyright (c) 2023-2026 Matteo Pacini
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/config"
	_ "modernc.org/sqlite"
)

const (
	currentSchemaVersion = 1
	databaseFileName     = "radiogogo.db"
)

// SQLiteStorage implements StationStorageService using SQLite.
type SQLiteStorage struct {
	mu        sync.RWMutex
	db        *sql.DB
	bookmarks map[uuid.UUID]bool
	hidden    map[uuid.UUID]bool
}

// NewSQLiteStorage creates a new SQLiteStorage instance.
func NewSQLiteStorage() (*SQLiteStorage, error) {
	s := &SQLiteStorage{
		bookmarks: make(map[uuid.UUID]bool),
		hidden:    make(map[uuid.UUID]bool),
	}

	// Ensure config directory exists
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(config.ConfigDir(), databaseFileName)

	// Open database with WAL mode for better concurrent access
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	// Validate database integrity
	if err := s.validateDatabase(db); err != nil {
		db.Close()
		// Attempt recovery: rename corrupted DB and create fresh
		if recoverErr := s.recoverCorruptedDatabase(dbPath); recoverErr != nil {
			return nil, fmt.Errorf("database corrupted and recovery failed: %w", err)
		}
		// Retry with fresh database
		db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
		if err != nil {
			return nil, err
		}
	}

	s.db = db

	// Initialize schema
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	// Load data into memory cache
	if err := s.loadCaches(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// validateDatabase checks database integrity.
func (s *SQLiteStorage) validateDatabase(db *sql.DB) error {
	var result string
	err := db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("database integrity check failed: %s", result)
	}
	return nil
}

// recoverCorruptedDatabase renames a corrupted database file.
func (s *SQLiteStorage) recoverCorruptedDatabase(dbPath string) error {
	backupPath := dbPath + ".corrupted." + time.Now().Format("20060102-150405")
	return os.Rename(dbPath, backupPath)
}

// initSchema creates the database tables if they don't exist.
func (s *SQLiteStorage) initSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		);

		CREATE TABLE IF NOT EXISTS bookmarks (
			station_uuid TEXT PRIMARY KEY,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS hidden (
			station_uuid TEXT PRIMARY KEY,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);

		INSERT OR IGNORE INTO schema_version (version) VALUES (?);
	`, currentSchemaVersion)
	return err
}

// loadCaches loads bookmarks and hidden stations into memory.
func (s *SQLiteStorage) loadCaches() error {
	// Load bookmarks into cache
	rows, err := s.db.Query("SELECT station_uuid FROM bookmarks")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uuidStr string
		if err := rows.Scan(&uuidStr); err != nil {
			continue
		}
		if id, err := uuid.Parse(uuidStr); err == nil {
			s.bookmarks[id] = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Load hidden into cache
	rows, err = s.db.Query("SELECT station_uuid FROM hidden")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uuidStr string
		if err := rows.Scan(&uuidStr); err != nil {
			continue
		}
		if id, err := uuid.Parse(uuidStr); err == nil {
			s.hidden[id] = true
		}
	}

	return rows.Err()
}

// Close closes the database connection.
func (s *SQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetBookmarks returns all bookmarked station UUIDs.
func (s *SQLiteStorage) GetBookmarks() ([]uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]uuid.UUID, 0, len(s.bookmarks))
	for id := range s.bookmarks {
		result = append(result, id)
	}
	return result, nil
}

// AddBookmark adds a station to bookmarks.
func (s *SQLiteStorage) AddBookmark(stationUUID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT OR IGNORE INTO bookmarks (station_uuid) VALUES (?)",
		stationUUID.String())
	if err != nil {
		return err
	}
	s.bookmarks[stationUUID] = true
	return nil
}

// RemoveBookmark removes a station from bookmarks.
func (s *SQLiteStorage) RemoveBookmark(stationUUID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM bookmarks WHERE station_uuid = ?",
		stationUUID.String())
	if err != nil {
		return err
	}
	delete(s.bookmarks, stationUUID)
	return nil
}

// IsBookmarked returns true if the station is bookmarked.
func (s *SQLiteStorage) IsBookmarked(stationUUID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bookmarks[stationUUID]
}

// GetHidden returns all hidden station UUIDs.
func (s *SQLiteStorage) GetHidden() ([]uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]uuid.UUID, 0, len(s.hidden))
	for id := range s.hidden {
		result = append(result, id)
	}
	return result, nil
}

// AddHidden hides a station from search results.
func (s *SQLiteStorage) AddHidden(stationUUID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT OR IGNORE INTO hidden (station_uuid) VALUES (?)",
		stationUUID.String())
	if err != nil {
		return err
	}
	s.hidden[stationUUID] = true
	return nil
}

// RemoveHidden unhides a station.
func (s *SQLiteStorage) RemoveHidden(stationUUID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM hidden WHERE station_uuid = ?",
		stationUUID.String())
	if err != nil {
		return err
	}
	delete(s.hidden, stationUUID)
	return nil
}

// IsHidden returns true if the station is hidden.
func (s *SQLiteStorage) IsHidden(stationUUID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hidden[stationUUID]
}
