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
	"time"

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

func TestSQLiteStorage_VoteTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("returns false when no vote timestamp exists", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		timestamp, found := s.GetLastVoteTimestamp()
		assert.False(t, found)
		assert.True(t, timestamp.IsZero())
	})

	t.Run("sets and retrieves vote timestamp", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		now := time.Now().Truncate(time.Second)
		err = s.SetLastVoteTimestamp(now)
		assert.NoError(t, err)

		timestamp, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
		assert.Equal(t, now.UTC(), timestamp.UTC())
	})

	t.Run("overwrites existing vote timestamp", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		old := time.Now().Add(-20 * time.Minute).Truncate(time.Second)
		err = s.SetLastVoteTimestamp(old)
		assert.NoError(t, err)

		new := time.Now().Truncate(time.Second)
		err = s.SetLastVoteTimestamp(new)
		assert.NoError(t, err)

		timestamp, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
		assert.Equal(t, new.UTC(), timestamp.UTC())
	})

	t.Run("persists vote timestamp across reload", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s1, err := NewSQLiteStorage()
		assert.NoError(t, err)

		now := time.Now().Truncate(time.Second)
		err = s1.SetLastVoteTimestamp(now)
		assert.NoError(t, err)
		s1.Close()

		// Create new instance (simulates app restart)
		s2, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s2.Close()

		timestamp, found := s2.GetLastVoteTimestamp()
		assert.True(t, found)
		assert.Equal(t, now.UTC(), timestamp.UTC())
	})

	t.Run("handles timestamps in different timezones", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		// Create a timestamp in a specific timezone
		loc, _ := time.LoadLocation("America/New_York")
		nyTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)

		err = s.SetLastVoteTimestamp(nyTime)
		assert.NoError(t, err)

		timestamp, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
		// Should be equal when compared in UTC
		assert.Equal(t, nyTime.UTC(), timestamp.UTC())
	})

	t.Run("handles zero timestamp", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		zero := time.Time{}
		err = s.SetLastVoteTimestamp(zero)
		assert.NoError(t, err)

		_, found := s.GetLastVoteTimestamp()
		assert.True(t, found) // We set it, so it should be found
		// The timestamp may not be exactly zero due to format/parse cycle
	})

	t.Run("handles far future timestamp", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		future := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
		err = s.SetLastVoteTimestamp(future)
		assert.NoError(t, err)

		timestamp, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
		assert.Equal(t, future, timestamp.UTC())
	})

	t.Run("handles far past timestamp", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		past := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		err = s.SetLastVoteTimestamp(past)
		assert.NoError(t, err)

		timestamp, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
		assert.Equal(t, past, timestamp.UTC())
	})
}

func TestSQLiteStorage_VoteTimestamp_Concurrent(t *testing.T) {
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

	t.Run("concurrent vote timestamp reads and writes", func(t *testing.T) {
		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				now := time.Now()
				_ = s.SetLastVoteTimestamp(now)
			}()
		}

		// Concurrent reads
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = s.GetLastVoteTimestamp()
			}()
		}

		wg.Wait()

		// Verify a timestamp was recorded
		_, found := s.GetLastVoteTimestamp()
		assert.True(t, found)
	})
}

func TestSQLiteStorage_RemoveNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("removing nonexistent bookmark does not error", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		nonexistent := uuid.New()
		err = s.RemoveBookmark(nonexistent)
		assert.NoError(t, err)
	})

	t.Run("removing nonexistent hidden station does not error", func(t *testing.T) {
		os.Remove(filepath.Join(configDir, databaseFileName))

		s, err := NewSQLiteStorage()
		assert.NoError(t, err)
		defer s.Close()

		nonexistent := uuid.New()
		err = s.RemoveHidden(nonexistent)
		assert.NoError(t, err)
	})
}

func TestSQLiteStorage_NilUUID(t *testing.T) {
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

	t.Run("can bookmark nil UUID", func(t *testing.T) {
		err := s.AddBookmark(uuid.Nil)
		assert.NoError(t, err)
		assert.True(t, s.IsBookmarked(uuid.Nil))

		err = s.RemoveBookmark(uuid.Nil)
		assert.NoError(t, err)
		assert.False(t, s.IsBookmarked(uuid.Nil))
	})

	t.Run("can hide nil UUID", func(t *testing.T) {
		err := s.AddHidden(uuid.Nil)
		assert.NoError(t, err)
		assert.True(t, s.IsHidden(uuid.Nil))

		err = s.RemoveHidden(uuid.Nil)
		assert.NoError(t, err)
		assert.False(t, s.IsHidden(uuid.Nil))
	})
}
