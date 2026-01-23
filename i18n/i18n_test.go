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

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("initializes with valid language", func(t *testing.T) {
		err := Init("en")
		assert.NoError(t, err)
		assert.Equal(t, "en", CurrentLanguage())
	})

	t.Run("initializes with another valid language", func(t *testing.T) {
		err := Init("de")
		assert.NoError(t, err)
		assert.Equal(t, "de", CurrentLanguage())
	})

	t.Run("handles empty language by defaulting to en", func(t *testing.T) {
		err := Init("")
		assert.NoError(t, err)
		assert.Equal(t, "en", CurrentLanguage())
	})

	t.Run("accepts unknown language but still initializes", func(t *testing.T) {
		// The i18n library accepts unknown languages but falls back to English for translations
		err := Init("xx")
		assert.NoError(t, err)
		assert.Equal(t, "xx", CurrentLanguage())
	})
}

func TestSetLanguage(t *testing.T) {
	// Initialize first
	_ = Init("en")

	t.Run("switches to another language", func(t *testing.T) {
		err := SetLanguage("de")
		assert.NoError(t, err)
		assert.Equal(t, "de", CurrentLanguage())
	})

	t.Run("switches back to English", func(t *testing.T) {
		err := SetLanguage("en")
		assert.NoError(t, err)
		assert.Equal(t, "en", CurrentLanguage())
	})

	t.Run("handles empty language by defaulting to en", func(t *testing.T) {
		_ = SetLanguage("de")
		err := SetLanguage("")
		assert.NoError(t, err)
		assert.Equal(t, "en", CurrentLanguage())
	})
}

func TestCurrentLanguage(t *testing.T) {
	t.Run("returns current language after Init", func(t *testing.T) {
		_ = Init("es")
		assert.Equal(t, "es", CurrentLanguage())
	})

	t.Run("returns current language after SetLanguage", func(t *testing.T) {
		_ = Init("en")
		_ = SetLanguage("it")
		assert.Equal(t, "it", CurrentLanguage())
	})
}

func TestAvailableLanguages(t *testing.T) {
	t.Run("returns all 9 languages", func(t *testing.T) {
		langs := AvailableLanguages()
		assert.Len(t, langs, 9)
	})

	t.Run("returns sorted list", func(t *testing.T) {
		langs := AvailableLanguages()
		expected := []string{"de", "el", "en", "es", "it", "ja", "pt", "ru", "zh"}
		assert.Equal(t, expected, langs)
	})

	t.Run("includes all expected languages", func(t *testing.T) {
		langs := AvailableLanguages()
		assert.Contains(t, langs, "en")
		assert.Contains(t, langs, "de")
		assert.Contains(t, langs, "es")
		assert.Contains(t, langs, "it")
		assert.Contains(t, langs, "ja")
		assert.Contains(t, langs, "zh")
	})
}

func TestT(t *testing.T) {
	_ = Init("en")

	t.Run("returns translated string for valid message ID", func(t *testing.T) {
		result := T("current_language")
		assert.Equal(t, "EN", result)
	})

	t.Run("returns message ID for unknown message", func(t *testing.T) {
		result := T("unknown_message_id")
		assert.Equal(t, "unknown_message_id", result)
	})

	t.Run("returns different translation for different language", func(t *testing.T) {
		_ = SetLanguage("de")
		result := T("current_language")
		assert.Equal(t, "DE", result)

		_ = SetLanguage("en") // Reset
	})
}

func TestTf(t *testing.T) {
	_ = Init("en")

	t.Run("substitutes template variables", func(t *testing.T) {
		result := Tf("cmd_quit", map[string]interface{}{"Key": "q"})
		assert.Equal(t, "q: quit", result)
	})

	t.Run("substitutes multiple template variables", func(t *testing.T) {
		result := Tf("cmd_volume", map[string]interface{}{
			"VolumeDown": "9",
			"VolumeUp":   "0",
		})
		assert.Equal(t, "9/0: vol", result)
	})

	t.Run("returns message ID for unknown message", func(t *testing.T) {
		result := Tf("unknown_message_id", map[string]interface{}{"Key": "x"})
		assert.Equal(t, "unknown_message_id", result)
	})

	t.Run("handles nil data", func(t *testing.T) {
		// search_placeholder doesn't use template variables
		result := Tf("search_placeholder", nil)
		assert.Equal(t, "Name", result)
	})

	t.Run("handles empty data map", func(t *testing.T) {
		result := Tf("search_placeholder", map[string]interface{}{})
		assert.Equal(t, "Name", result)
	})
}

func TestTn(t *testing.T) {
	_ = Init("en")

	t.Run("returns message ID for unknown message", func(t *testing.T) {
		result := Tn("unknown_message_id", 1)
		assert.Equal(t, "unknown_message_id", result)
	})

	t.Run("returns message ID when no plural form defined", func(t *testing.T) {
		// go-i18n requires explicit plural forms (one/other) for Tn to work
		// Messages with only "other" key don't match PluralCount requests
		result := Tn("search_placeholder", 1)
		// Returns message ID since no plural form is defined
		assert.Equal(t, "search_placeholder", result)
	})
}

func TestTfn(t *testing.T) {
	_ = Init("en")

	t.Run("returns message ID for unknown message", func(t *testing.T) {
		result := Tfn("unknown_message_id", 1, map[string]interface{}{"Key": "x"})
		assert.Equal(t, "unknown_message_id", result)
	})

	t.Run("returns message ID when no plural form defined", func(t *testing.T) {
		// go-i18n requires explicit plural forms for Tfn to work
		result := Tfn("search_title", 1, map[string]interface{}{"Type": "stations"})
		// Returns message ID since no plural form is defined
		assert.Equal(t, "search_title", result)
	})
}

func TestLanguageSwitchingUpdatesTranslations(t *testing.T) {
	t.Run("translations change when language switches", func(t *testing.T) {
		_ = Init("en")
		enResult := T("current_language")
		assert.Equal(t, "EN", enResult)

		_ = SetLanguage("de")
		deResult := T("current_language")
		assert.Equal(t, "DE", deResult)

		_ = SetLanguage("es")
		esResult := T("current_language")
		assert.Equal(t, "ES", esResult)

		_ = SetLanguage("en") // Reset
	})
}

func TestAutoInitialization(t *testing.T) {
	// Reset internal state to test auto-initialization
	localizer = nil
	bundle = nil

	t.Run("T auto-initializes if not initialized", func(t *testing.T) {
		result := T("current_language")
		// Should auto-init to English
		assert.Equal(t, "EN", result)
	})
}
