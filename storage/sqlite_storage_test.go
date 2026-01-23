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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteStorage_Bookmarks(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("adds and retrieves bookmarks", func(t *testing.T) {
		// Clean up any existing database
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		id1 := uuid.New()
		id2 := uuid.New()

		err = s.AddBookmark(id1)
		assert.NoError(t, err)
		err = s.AddBookmark(id2)
		assert.NoError(t, err)

		assert.True(t, s.IsBookmarked(id1))
		assert.True(t, s.IsBookmarked(id2))
		assert.False(t, s.IsBookmarked(uuid.New()))

		bookmarks, err := s.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 2)
	})

	t.Run("removes bookmarks", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		id := uuid.New()

		err = s.AddBookmark(id)
		assert.NoError(t, err)
		assert.True(t, s.IsBookmarked(id))

		err = s.RemoveBookmark(id)
		assert.NoError(t, err)
		assert.False(t, s.IsBookmarked(id))
	})

	t.Run("persists bookmarks across reload", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s1, err := NewSQLiteStorage()
		assert.NoError(t, err)

		id := uuid.New()
		err = s1.AddBookmark(id)
		assert.NoError(t, err)
		s1.Close()

		// Create new instance (simulates app restart)
		s2, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s2.Close()

		assert.True(t, s2.IsBookmarked(id))
	})

	t.Run("handles duplicate bookmarks", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		id := uuid.New()

		err = s.AddBookmark(id)
		assert.NoError(t, err)
		err = s.AddBookmark(id) // Add same bookmark again
		assert.NoError(t, err)

		bookmarks, err := s.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 1) // Should only have one entry
	})
}

func TestSQLiteStorage_Hidden(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("adds and retrieves hidden stations", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		id1 := uuid.New()
		id2 := uuid.New()

		err = s.AddHidden(id1)
		assert.NoError(t, err)
		err = s.AddHidden(id2)
		assert.NoError(t, err)

		assert.True(t, s.IsHidden(id1))
		assert.True(t, s.IsHidden(id2))
		assert.False(t, s.IsHidden(uuid.New()))

		hidden, err := s.GetHidden()
		assert.NoError(t, err)
		assert.Len(t, hidden, 2)
	})

	t.Run("removes hidden stations", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		id := uuid.New()

		err = s.AddHidden(id)
		assert.NoError(t, err)
		assert.True(t, s.IsHidden(id))

		err = s.RemoveHidden(id)
		assert.NoError(t, err)
		assert.False(t, s.IsHidden(id))
	})

	t.Run("persists hidden across reload", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s1, err := NewSQLiteStorage()
		assert.NoError(t, err)

		id := uuid.New()
		err = s1.AddHidden(id)
		assert.NoError(t, err)
		s1.Close()

		s2, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s2.Close()

		assert.True(t, s2.IsHidden(id))
	})
}

func TestSQLiteStorage_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("creates storage with empty lists when database doesn't exist", func(t *testing.T) {
		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		bookmarks, err := s.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 0)

		hidden, err := s.GetHidden()
		assert.NoError(t, err)
		assert.Len(t, hidden, 0)
	})
}

func TestSQLiteStorage_CorruptedDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("recovers from corrupted database", func(t *testing.T) {
		dbPath := filepath.Join(configDir, databaseFileName)

		// Create a corrupted database file
		err := os.WriteFile(dbPath, []byte("this is not a valid sqlite database"), 0644)
		assert.NoError(t, err)

		// Should recover by renaming the corrupted file
		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		// Should work with fresh database
		id := uuid.New()
		err = s.AddBookmark(id)
		assert.NoError(t, err)
		assert.True(t, s.IsBookmarked(id))

		// Verify corrupted file was renamed
		matches, _ := filepath.Glob(filepath.Join(configDir, "radiogogo.db.corrupted.*"))
		assert.Len(t, matches, 1)
	})
}

func TestSQLiteStorage_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	os.Remove(filepath.Join(configDir, databaseFileName))

	s, err := NewSQLiteStorage()
	assert.NoError(t, err)
	defer s.Close()

	t.Run("concurrent reads and writes", func(t *testing.T) {
		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				id := uuid.New()
				_ = s.AddBookmark(id)
			}()
		}

		// Concurrent reads
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = s.GetBookmarks()
			}()
		}

		wg.Wait()

		// Verify all writes completed
		bookmarks, err := s.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 50)
	})
}

func TestSQLiteStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("close is idempotent", func(t *testing.T) {
		s, err := NewSQLiteStorage()
		assert.NoError(t, err)

		err = s.Close()
		assert.NoError(t, err)

		// Second close should not error
		err = s.Close()
		assert.NoError(t, err)
	})
}
