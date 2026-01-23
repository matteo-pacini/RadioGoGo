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

// Package i18n provides internationalization support for RadioGoGo.
package i18n

import (
	"embed"
	"io/fs"
	"sort"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var localeFS embed.FS

var (
	bundle      *i18n.Bundle
	localizer   *i18n.Localizer
	currentLang string
	availLangs  []string
)

// Init initializes the i18n system with the specified language.
// Falls back to English if the requested language is not available.
func Init(lang string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load all embedded locale files and track available languages
	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return err
	}

	availLangs = make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := localeFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			return err
		}
		bundle.MustParseMessageFileBytes(data, entry.Name())

		// Extract language code from filename (e.g., "en.yaml" -> "en")
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".yaml" {
			availLangs = append(availLangs, name[:len(name)-5])
		}
	}

	// Create localizer with language preference chain
	if lang == "" {
		lang = "en"
	}
	currentLang = lang
	localizer = i18n.NewLocalizer(bundle, lang, "en")

	return nil
}

// AvailableLanguages returns a sorted list of available language codes.
func AvailableLanguages() []string {
	if len(availLangs) == 0 {
		_ = Init("en")
	}
	// Return a sorted copy to ensure consistent ordering
	sorted := make([]string, len(availLangs))
	copy(sorted, availLangs)
	sort.Strings(sorted)
	return sorted
}

// CurrentLanguage returns the current language code.
func CurrentLanguage() string {
	return currentLang
}

// SetLanguage switches to a new language at runtime.
func SetLanguage(lang string) error {
	if lang == "" {
		lang = "en"
	}
	currentLang = lang
	localizer = i18n.NewLocalizer(bundle, lang, "en")
	return nil
}

// T returns the localized string for the given message ID.
func T(messageID string) string {
	// Auto-initialize if not done (for tests)
	if localizer == nil {
		_ = Init("en")
	}
	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return messageID // Fallback to message ID if not found
	}
	return result
}

// Tf returns the localized string for the given message ID with template data.
func Tf(messageID string, data map[string]interface{}) string {
	// Auto-initialize if not done (for tests)
	if localizer == nil {
		_ = Init("en")
	}
	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return result
}

// Tn returns the localized string for the given message ID with pluralization.
func Tn(messageID string, count int) string {
	// Auto-initialize if not done (for tests)
	if localizer == nil {
		_ = Init("en")
	}
	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:   messageID,
		PluralCount: count,
	})
	if err != nil {
		return messageID
	}
	return result
}

// Tfn returns the localized string with both template data and pluralization.
func Tfn(messageID string, count int, data map[string]interface{}) string {
	// Auto-initialize if not done (for tests)
	if localizer == nil {
		_ = Init("en")
	}
	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return result
}
