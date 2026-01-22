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
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/config"
)

// FileStorage implements StationStorageService using text files.
type FileStorage struct {
	mu        sync.RWMutex
	bookmarks map[uuid.UUID]bool
	hidden    map[uuid.UUID]bool
}

// NewFileStorage creates a new FileStorage instance, loading existing data from disk.
func NewFileStorage() (*FileStorage, error) {
	fs := &FileStorage{
		bookmarks: make(map[uuid.UUID]bool),
		hidden:    make(map[uuid.UUID]bool),
	}

	// Ensure config directory exists
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return nil, err
	}

	// Load bookmarks
	bookmarks, err := fs.loadFile(bookmarksFilePath())
	if err != nil {
		return nil, err
	}
	fs.bookmarks = bookmarks

	// Load hidden
	hidden, err := fs.loadFile(hiddenFilePath())
	if err != nil {
		return nil, err
	}
	fs.hidden = hidden

	return fs, nil
}

func bookmarksFilePath() string {
	return filepath.Join(config.ConfigDir(), "bookmarks.txt")
}

func hiddenFilePath() string {
	return filepath.Join(config.ConfigDir(), "hidden.txt")
}

// loadFile reads UUIDs from a text file (one per line).
// Returns empty map if file doesn't exist.
func (fs *FileStorage) loadFile(path string) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool)

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		id, err := uuid.Parse(line)
		if err != nil {
			// Skip invalid lines
			continue
		}
		result[id] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// saveFile writes UUIDs to a text file (one per line).
func (fs *FileStorage) saveFile(path string, uuids map[uuid.UUID]bool) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for id := range uuids {
		if _, err := file.WriteString(id.String() + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// GetBookmarks returns all bookmarked station UUIDs.
func (fs *FileStorage) GetBookmarks() ([]uuid.UUID, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]uuid.UUID, 0, len(fs.bookmarks))
	for id := range fs.bookmarks {
		result = append(result, id)
	}
	return result, nil
}

// AddBookmark adds a station to bookmarks.
func (fs *FileStorage) AddBookmark(stationUUID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.bookmarks[stationUUID] = true
	return fs.saveFile(bookmarksFilePath(), fs.bookmarks)
}

// RemoveBookmark removes a station from bookmarks.
func (fs *FileStorage) RemoveBookmark(stationUUID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.bookmarks, stationUUID)
	return fs.saveFile(bookmarksFilePath(), fs.bookmarks)
}

// IsBookmarked returns true if the station is bookmarked.
func (fs *FileStorage) IsBookmarked(stationUUID uuid.UUID) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.bookmarks[stationUUID]
}

// GetHidden returns all hidden station UUIDs.
func (fs *FileStorage) GetHidden() ([]uuid.UUID, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]uuid.UUID, 0, len(fs.hidden))
	for id := range fs.hidden {
		result = append(result, id)
	}
	return result, nil
}

// AddHidden hides a station from search results.
func (fs *FileStorage) AddHidden(stationUUID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.hidden[stationUUID] = true
	return fs.saveFile(hiddenFilePath(), fs.hidden)
}

// RemoveHidden unhides a station.
func (fs *FileStorage) RemoveHidden(stationUUID uuid.UUID) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.hidden, stationUUID)
	return fs.saveFile(hiddenFilePath(), fs.hidden)
}

// IsHidden returns true if the station is hidden.
func (fs *FileStorage) IsHidden(stationUUID uuid.UUID) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.hidden[stationUUID]
}
