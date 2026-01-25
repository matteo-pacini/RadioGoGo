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

func TestFormatTimestamp_EdgeCases(t *testing.T) {
	t.Run("midnight", func(t *testing.T) {
		midnight := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		result := FormatTimestamp(midnight)
		assert.Equal(t, "2026-01-15-00-00-00", result)
	})

	t.Run("end of day", func(t *testing.T) {
		endOfDay := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
		result := FormatTimestamp(endOfDay)
		assert.Equal(t, "2026-12-31-23-59-59", result)
	})

	t.Run("leap year February 29", func(t *testing.T) {
		leapDay := time.Date(2028, 2, 29, 12, 30, 0, 0, time.UTC)
		result := FormatTimestamp(leapDay)
		assert.Equal(t, "2028-02-29-12-30-00", result)
	})

	t.Run("nanoseconds ignored", func(t *testing.T) {
		withNanos := time.Date(2026, 5, 10, 15, 45, 30, 123456789, time.UTC)
		result := FormatTimestamp(withNanos)
		assert.Equal(t, "2026-05-10-15-45-30", result)
	})

	t.Run("different timezone produces local time", func(t *testing.T) {
		loc, _ := time.LoadLocation("America/New_York")
		nyTime := time.Date(2026, 7, 4, 10, 0, 0, 0, loc)
		result := FormatTimestamp(nyTime)
		// Should format the time as-is (10:00) not convert to UTC
		assert.Equal(t, "2026-07-04-10-00-00", result)
	})

	t.Run("zero time", func(t *testing.T) {
		zeroTime := time.Time{}
		result := FormatTimestamp(zeroTime)
		// Go's zero time is 0001-01-01 00:00:00
		assert.Equal(t, "0001-01-01-00-00-00", result)
	})
}

func TestNormalizeCodec_EdgeCases(t *testing.T) {
	t.Run("2 character codec used as-is", func(t *testing.T) {
		result := NormalizeCodec("ts")
		assert.Equal(t, "ts", result)
	})

	t.Run("5 character codec used as-is", func(t *testing.T) {
		result := NormalizeCodec("webma")
		assert.Equal(t, "webma", result)
	})

	t.Run("6+ character unknown codec defaults to mp3", func(t *testing.T) {
		result := NormalizeCodec("unknwn")
		assert.Equal(t, "mp3", result)
	})

	t.Run("codec with leading whitespace", func(t *testing.T) {
		result := NormalizeCodec("  mp3")
		assert.Equal(t, "mp3", result)
	})

	t.Run("codec with trailing whitespace", func(t *testing.T) {
		result := NormalizeCodec("aac  ")
		assert.Equal(t, "aac", result)
	})

	t.Run("codec with surrounding whitespace", func(t *testing.T) {
		result := NormalizeCodec("  ogg  ")
		assert.Equal(t, "ogg", result)
	})

	t.Run("wma codec", func(t *testing.T) {
		result := NormalizeCodec("WMA")
		assert.Equal(t, "wma", result)
	})

	t.Run("wav codec", func(t *testing.T) {
		result := NormalizeCodec("WAV")
		assert.Equal(t, "wav", result)
	})

	t.Run("mixed case aac+", func(t *testing.T) {
		result := NormalizeCodec("AaC+")
		assert.Equal(t, "aac", result)
	})
}

func TestSanitizeFilename_EdgeCases(t *testing.T) {
	t.Run("exactly 100 characters", func(t *testing.T) {
		// Create exactly 100 character string
		name := ""
		for i := 0; i < 100; i++ {
			name += "a"
		}
		result := SanitizeFilename(name)
		assert.Equal(t, 100, len(result))
	})

	t.Run("single character name", func(t *testing.T) {
		result := SanitizeFilename("A")
		assert.Equal(t, "a", result)
	})

	t.Run("numbers only", func(t *testing.T) {
		result := SanitizeFilename("12345")
		assert.Equal(t, "12345", result)
	})

	t.Run("mixed numbers and letters", func(t *testing.T) {
		result := SanitizeFilename("Radio 101.5 FM")
		assert.Equal(t, "radio_1015_fm", result)
	})

	t.Run("leading hyphen trimmed", func(t *testing.T) {
		result := SanitizeFilename("-test")
		assert.Equal(t, "test", result)
	})

	t.Run("trailing hyphen trimmed", func(t *testing.T) {
		result := SanitizeFilename("test-")
		assert.Equal(t, "test", result)
	})

	t.Run("multiple consecutive hyphens", func(t *testing.T) {
		result := SanitizeFilename("test---name")
		assert.Equal(t, "test---name", result) // hyphens not collapsed, only underscores
	})

	t.Run("tabs converted to underscores", func(t *testing.T) {
		result := SanitizeFilename("test\tname")
		// tabs become empty (removed by regex), spaces become underscore
		assert.Equal(t, "testname", result)
	})

	t.Run("newlines removed", func(t *testing.T) {
		result := SanitizeFilename("test\nname")
		assert.Equal(t, "testname", result)
	})

	t.Run("emoji removed with fallback", func(t *testing.T) {
		result := SanitizeFilename("🎵🎶🎸")
		assert.Equal(t, "recording", result)
	})

	t.Run("emoji with ASCII text", func(t *testing.T) {
		result := SanitizeFilename("🎵 Radio One 🎶")
		assert.Equal(t, "radio_one", result)
	})
}

func TestGenerateRecordingFilename_EdgeCases(t *testing.T) {
	t.Run("empty station name and empty codec", func(t *testing.T) {
		filename := GenerateRecordingFilename("", "")
		// Should use fallback "recording" and default "mp3"
		assert.Contains(t, filename, "recording-")
		assert.Contains(t, filename, ".mp3")
	})

	t.Run("very long station name truncated", func(t *testing.T) {
		longName := ""
		for i := 0; i < 200; i++ {
			longName += "a"
		}
		filename := GenerateRecordingFilename(longName, "mp3")
		// Name part should be truncated to 100 chars
		// Total filename will be: 100 chars + "-" + 19 chars timestamp + ".mp3" = 124 chars
		assert.LessOrEqual(t, len(filename), 130)
	})

	t.Run("station name with only special characters", func(t *testing.T) {
		filename := GenerateRecordingFilename("!@#$%^&*()", "aac")
		assert.Contains(t, filename, "recording-")
		assert.Contains(t, filename, ".aac")
	})
}
