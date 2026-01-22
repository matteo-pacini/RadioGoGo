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

package playback

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic station name",
			input:    "BBC Radio 1",
			expected: "bbc_radio_1",
		},
		{
			name:     "name with special characters",
			input:    "Radio 538 FM!",
			expected: "radio_538_fm",
		},
		{
			name:     "unicode characters removed",
			input:    "Радио Россия",
			expected: "recording", // all removed, falls back
		},
		{
			name:     "mixed unicode and ascii",
			input:    "Radio Россия FM",
			expected: "radio_fm",
		},
		{
			name:     "multiple spaces",
			input:    "  Spaces  Here  ",
			expected: "spaces_here",
		},
		{
			name:     "multiple underscores collapsed",
			input:    "Under__scores___test",
			expected: "under_scores_test",
		},
		{
			name:     "uppercase converted",
			input:    "ALLCAPS",
			expected: "allcaps",
		},
		{
			name:     "special characters removed",
			input:    "Special@#$%Chars!",
			expected: "specialchars",
		},
		{
			name:     "empty string fallback",
			input:    "",
			expected: "recording",
		},
		{
			name:     "only special chars fallback",
			input:    "@#$%^&*()",
			expected: "recording",
		},
		{
			name:     "hyphens preserved",
			input:    "Radio-One",
			expected: "radio-one",
		},
		{
			name:     "leading/trailing underscores trimmed",
			input:    "_test_name_",
			expected: "test_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeFilename_LengthLimit(t *testing.T) {
	// Create a string longer than 100 characters
	longName := ""
	for i := 0; i < 150; i++ {
		longName += "a"
	}

	result := SanitizeFilename(longName)
	assert.LessOrEqual(t, len(result), 100)
}

func TestNormalizeCodec(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mp3 lowercase",
			input:    "mp3",
			expected: "mp3",
		},
		{
			name:     "MP3 uppercase",
			input:    "MP3",
			expected: "mp3",
		},
		{
			name:     "aac",
			input:    "aac",
			expected: "aac",
		},
		{
			name:     "AAC+ maps to aac",
			input:    "AAC+",
			expected: "aac",
		},
		{
			name:     "aacp maps to aac",
			input:    "aacp",
			expected: "aac",
		},
		{
			name:     "ogg",
			input:    "ogg",
			expected: "ogg",
		},
		{
			name:     "vorbis maps to ogg",
			input:    "vorbis",
			expected: "ogg",
		},
		{
			name:     "opus",
			input:    "OPUS",
			expected: "opus",
		},
		{
			name:     "flac",
			input:    "flac",
			expected: "flac",
		},
		{
			name:     "empty string defaults to mp3",
			input:    "",
			expected: "mp3",
		},
		{
			name:     "whitespace only defaults to mp3",
			input:    "   ",
			expected: "mp3",
		},
		{
			name:     "unknown short codec used as-is",
			input:    "m4a",
			expected: "m4a",
		},
		{
			name:     "unknown long codec defaults to mp3",
			input:    "unknowncodec",
			expected: "mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeCodec(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	testTime := time.Date(2026, 1, 22, 18, 32, 45, 0, time.UTC)
	result := FormatTimestamp(testTime)
	assert.Equal(t, "2026-01-22-18-32-45", result)
}

func TestGenerateRecordingFilename(t *testing.T) {
	t.Run("generates correct format", func(t *testing.T) {
		filename := GenerateRecordingFilename("BBC Radio 1", "MP3")
		// Should match pattern: bbc_radio_1-YYYY-MM-DD-HH-MM-SS.mp3
		pattern := `^bbc_radio_1-\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}\.mp3$`
		matched, err := regexp.MatchString(pattern, filename)
		assert.NoError(t, err)
		assert.True(t, matched, "filename %s should match pattern %s", filename, pattern)
	})

	t.Run("handles empty codec", func(t *testing.T) {
		filename := GenerateRecordingFilename("Test Radio", "")
		assert.Contains(t, filename, ".mp3")
	})

	t.Run("handles unicode station name", func(t *testing.T) {
		filename := GenerateRecordingFilename("Радио", "aac")
		assert.Contains(t, filename, "recording-")
		assert.Contains(t, filename, ".aac")
	})
}
