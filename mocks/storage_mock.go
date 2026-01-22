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

package mocks

import "github.com/google/uuid"

type MockStationStorageService struct {
	GetBookmarksFunc    func() ([]uuid.UUID, error)
	AddBookmarkFunc     func(stationUUID uuid.UUID) error
	RemoveBookmarkFunc  func(stationUUID uuid.UUID) error
	IsBookmarkedFunc    func(stationUUID uuid.UUID) bool
	GetHiddenFunc       func() ([]uuid.UUID, error)
	AddHiddenFunc       func(stationUUID uuid.UUID) error
	RemoveHiddenFunc    func(stationUUID uuid.UUID) error
	IsHiddenFunc        func(stationUUID uuid.UUID) bool
}

func (m *MockStationStorageService) GetBookmarks() ([]uuid.UUID, error) {
	if m.GetBookmarksFunc != nil {
		return m.GetBookmarksFunc()
	}
	return []uuid.UUID{}, nil
}

func (m *MockStationStorageService) AddBookmark(stationUUID uuid.UUID) error {
	if m.AddBookmarkFunc != nil {
		return m.AddBookmarkFunc(stationUUID)
	}
	return nil
}

func (m *MockStationStorageService) RemoveBookmark(stationUUID uuid.UUID) error {
	if m.RemoveBookmarkFunc != nil {
		return m.RemoveBookmarkFunc(stationUUID)
	}
	return nil
}

func (m *MockStationStorageService) IsBookmarked(stationUUID uuid.UUID) bool {
	if m.IsBookmarkedFunc != nil {
		return m.IsBookmarkedFunc(stationUUID)
	}
	return false
}

func (m *MockStationStorageService) GetHidden() ([]uuid.UUID, error) {
	if m.GetHiddenFunc != nil {
		return m.GetHiddenFunc()
	}
	return []uuid.UUID{}, nil
}

func (m *MockStationStorageService) AddHidden(stationUUID uuid.UUID) error {
	if m.AddHiddenFunc != nil {
		return m.AddHiddenFunc(stationUUID)
	}
	return nil
}

func (m *MockStationStorageService) RemoveHidden(stationUUID uuid.UUID) error {
	if m.RemoveHiddenFunc != nil {
		return m.RemoveHiddenFunc(stationUUID)
	}
	return nil
}

func (m *MockStationStorageService) IsHidden(stationUUID uuid.UUID) bool {
	if m.IsHiddenFunc != nil {
		return m.IsHiddenFunc(stationUUID)
	}
	return false
}
