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
	"time"

	"github.com/google/uuid"
)

// StationStorageService defines operations for persistent station lists (bookmarks and hidden).
type StationStorageService interface {
	// GetBookmarks returns all bookmarked station UUIDs.
	GetBookmarks() ([]uuid.UUID, error)
	// AddBookmark adds a station to bookmarks.
	AddBookmark(stationUUID uuid.UUID) error
	// RemoveBookmark removes a station from bookmarks.
	RemoveBookmark(stationUUID uuid.UUID) error
	// IsBookmarked returns true if the station is bookmarked.
	IsBookmarked(stationUUID uuid.UUID) bool

	// GetHidden returns all hidden station UUIDs.
	GetHidden() ([]uuid.UUID, error)
	// AddHidden hides a station from search results.
	AddHidden(stationUUID uuid.UUID) error
	// RemoveHidden unhides a station.
	RemoveHidden(stationUUID uuid.UUID) error
	// IsHidden returns true if the station is hidden.
	IsHidden(stationUUID uuid.UUID) bool

	// GetLastVoteTimestamp returns the last global vote timestamp.
	// Returns the timestamp and true if found, zero time and false if not.
	// RadioBrowser API enforces a 10-minute cooldown per IP for all votes.
	GetLastVoteTimestamp() (time.Time, bool)
	// SetLastVoteTimestamp records the last global vote timestamp.
	SetLastVoteTimestamp(timestamp time.Time) error
}
