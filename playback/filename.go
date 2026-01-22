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
	"strings"
	"time"
	"unicode"
)

// SanitizeFilename converts a station name to a safe filename component.
// - Converts to lowercase
// - Replaces spaces with underscores
// - Removes special characters (keeps alphanumeric, underscores, hyphens)
// - Handles unicode by removing non-ASCII characters
// - Limits length to prevent filesystem issues
func SanitizeFilename(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove non-ASCII characters (handles unicode)
	name = strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1 // remove
		}
		return r
	}, name)

	// Keep only alphanumeric, underscores, and hyphens
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	name = reg.ReplaceAllString(name, "")

	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	name = reg.ReplaceAllString(name, "_")

	// Trim leading/trailing underscores and hyphens
	name = strings.Trim(name, "_-")

	// Limit length (filesystem safe)
	if len(name) > 100 {
		name = name[:100]
	}

	// Fallback if empty
	if name == "" {
		name = "recording"
	}

	return name
}

// NormalizeCodec returns a lowercase file extension for the given codec.
// Returns "mp3" as default if codec is empty or unknown.
func NormalizeCodec(codec string) string {
	codec = strings.ToLower(strings.TrimSpace(codec))

	// Map common codec names to extensions
	codecMap := map[string]string{
		"mp3":    "mp3",
		"aac":    "aac",
		"aac+":   "aac",
		"aacp":   "aac",
		"ogg":    "ogg",
		"vorbis": "ogg",
		"opus":   "opus",
		"flac":   "flac",
		"wma":    "wma",
		"wav":    "wav",
	}

	if ext, ok := codecMap[codec]; ok {
		return ext
	}

	// Default to mp3 for unknown/empty
	if codec == "" {
		return "mp3"
	}

	// Use the codec as-is if it looks like a valid extension
	if len(codec) >= 2 && len(codec) <= 5 {
		return codec
	}

	return "mp3"
}

// FormatTimestamp returns a timestamp in the format: 2006-01-02-15-04-05
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02-15-04-05")
}

// GenerateRecordingFilename creates a filename for recording.
// Format: {sanitized_name}-{timestamp}.{codec}
// Example: bbc_radio_1-2026-01-22-18-32-00.mp3
func GenerateRecordingFilename(stationName string, codec string) string {
	sanitized := SanitizeFilename(stationName)
	timestamp := FormatTimestamp(time.Now())
	extension := NormalizeCodec(codec)

	return sanitized + "-" + timestamp + "." + extension
}
