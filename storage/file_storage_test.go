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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage_Bookmarks(t *testing.T) {
	// Create temp directory and override config dir
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create .config/radiogogo directory
	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("adds and retrieves bookmarks", func(t *testing.T) {
		fs, err := NewFileStorage()
		assert.NoError(t, err)

		id1 := uuid.New()
		id2 := uuid.New()

		err = fs.AddBookmark(id1)
		assert.NoError(t, err)
		err = fs.AddBookmark(id2)
		assert.NoError(t, err)

		assert.True(t, fs.IsBookmarked(id1))
		assert.True(t, fs.IsBookmarked(id2))
		assert.False(t, fs.IsBookmarked(uuid.New()))

		bookmarks, err := fs.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 2)
	})

	t.Run("removes bookmarks", func(t *testing.T) {
		fs, err := NewFileStorage()
		assert.NoError(t, err)

		id := uuid.New()

		err = fs.AddBookmark(id)
		assert.NoError(t, err)
		assert.True(t, fs.IsBookmarked(id))

		err = fs.RemoveBookmark(id)
		assert.NoError(t, err)
		assert.False(t, fs.IsBookmarked(id))
	})

	t.Run("persists bookmarks across reload", func(t *testing.T) {
		fs1, err := NewFileStorage()
		assert.NoError(t, err)

		id := uuid.New()
		err = fs1.AddBookmark(id)
		assert.NoError(t, err)

		// Create new instance (simulates app restart)
		fs2, err := NewFileStorage()
		assert.NoError(t, err)

		assert.True(t, fs2.IsBookmarked(id))
	})
}

func TestFileStorage_Hidden(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("adds and retrieves hidden stations", func(t *testing.T) {
		fs, err := NewFileStorage()
		assert.NoError(t, err)

		id1 := uuid.New()
		id2 := uuid.New()

		err = fs.AddHidden(id1)
		assert.NoError(t, err)
		err = fs.AddHidden(id2)
		assert.NoError(t, err)

		assert.True(t, fs.IsHidden(id1))
		assert.True(t, fs.IsHidden(id2))
		assert.False(t, fs.IsHidden(uuid.New()))

		hidden, err := fs.GetHidden()
		assert.NoError(t, err)
		assert.Len(t, hidden, 2)
	})

	t.Run("removes hidden stations", func(t *testing.T) {
		fs, err := NewFileStorage()
		assert.NoError(t, err)

		id := uuid.New()

		err = fs.AddHidden(id)
		assert.NoError(t, err)
		assert.True(t, fs.IsHidden(id))

		err = fs.RemoveHidden(id)
		assert.NoError(t, err)
		assert.False(t, fs.IsHidden(id))
	})

	t.Run("persists hidden across reload", func(t *testing.T) {
		fs1, err := NewFileStorage()
		assert.NoError(t, err)

		id := uuid.New()
		err = fs1.AddHidden(id)
		assert.NoError(t, err)

		fs2, err := NewFileStorage()
		assert.NoError(t, err)

		assert.True(t, fs2.IsHidden(id))
	})
}

func TestFileStorage_HandlesMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	t.Run("creates storage with empty lists when files don't exist", func(t *testing.T) {
		fs, err := NewFileStorage()
		assert.NoError(t, err)

		bookmarks, err := fs.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 0)

		hidden, err := fs.GetHidden()
		assert.NoError(t, err)
		assert.Len(t, hidden, 0)
	})
}

func TestFileStorage_HandlesInvalidUUIDs(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "radiogogo")
	err := os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	t.Run("skips invalid UUIDs in file", func(t *testing.T) {
		validID := uuid.New()
		content := "invalid-uuid\n" + validID.String() + "\nalso-invalid\n"
		err := os.WriteFile(filepath.Join(configDir, "bookmarks.txt"), []byte(content), 0644)
		assert.NoError(t, err)

		fs, err := NewFileStorage()
		assert.NoError(t, err)

		bookmarks, err := fs.GetBookmarks()
		assert.NoError(t, err)
		assert.Len(t, bookmarks, 1)
		assert.Equal(t, validID, bookmarks[0])
	})
}
