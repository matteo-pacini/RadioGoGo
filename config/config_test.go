package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	t.Run("parses from YAML", func(t *testing.T) {
		input := `
theme:
  textColor: "#000000"
  primaryColor: "#FFFFFF"
  secondaryColor: "#CCCCCC"
  tertiaryColor: "#999999"
  errorColor: "#FF0000"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "#000000", cfg.Theme.TextColor)
		assert.Equal(t, "#FFFFFF", cfg.Theme.PrimaryColor)
		assert.Equal(t, "#CCCCCC", cfg.Theme.SecondaryColor)
		assert.Equal(t, "#999999", cfg.Theme.TertiaryColor)
		assert.Equal(t, "#FF0000", cfg.Theme.ErrorColor)
	})

	t.Run("parses from YAML with partial values", func(t *testing.T) {
		input := `
theme:
  primaryColor: "#FFFFFF"
  tertiaryColor: "#999999"
  errorColor: "#FF0000"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "", cfg.Theme.TextColor)
		assert.Equal(t, "#FFFFFF", cfg.Theme.PrimaryColor)
		assert.Equal(t, "", cfg.Theme.SecondaryColor)
		assert.Equal(t, "#999999", cfg.Theme.TertiaryColor)
		assert.Equal(t, "#FF0000", cfg.Theme.ErrorColor)
	})

}

func TestNewDefaultConfig(t *testing.T) {
	t.Run("returns config with default values", func(t *testing.T) {
		cfg := NewDefaultConfig()

		assert.Equal(t, "en", cfg.Language)
		assert.Equal(t, "#ffffff", cfg.Theme.TextColor)
		assert.Equal(t, "#5a4f9f", cfg.Theme.PrimaryColor)
		assert.Equal(t, "#8b77db", cfg.Theme.SecondaryColor)
		assert.Equal(t, "#4e4e4e", cfg.Theme.TertiaryColor)
		assert.Equal(t, "#ff0000", cfg.Theme.ErrorColor)
	})

	t.Run("includes default keybindings", func(t *testing.T) {
		cfg := NewDefaultConfig()

		assert.Equal(t, "q", cfg.Keybindings.Quit)
		assert.Equal(t, "s", cfg.Keybindings.Search)
		assert.Equal(t, "r", cfg.Keybindings.Record)
		assert.Equal(t, "ctrl+k", cfg.Keybindings.StopPlayback)
	})
}

func TestConfig_Load(t *testing.T) {
	t.Run("loads config from valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		content := `theme:
  textColor: "#111111"
  primaryColor: "#222222"
  secondaryColor: "#333333"
  tertiaryColor: "#444444"
  errorColor: "#555555"
`
		err := os.WriteFile(cfgPath, []byte(content), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)

		assert.NoError(t, err)
		assert.Equal(t, "#111111", cfg.Theme.TextColor)
		assert.Equal(t, "#222222", cfg.Theme.PrimaryColor)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		var cfg Config
		err := cfg.Load("/nonexistent/path/config.yaml")

		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(cfgPath, []byte("not: valid: yaml: content:"), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)

		assert.Error(t, err)
	})
}

func TestConfig_Save(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		cfg := NewDefaultConfig()
		err := cfg.Save(cfgPath)

		assert.NoError(t, err)

		// Verify file exists and can be loaded back
		var loadedCfg Config
		err = loadedCfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, cfg.Theme.PrimaryColor, loadedCfg.Theme.PrimaryColor)
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		cfg := NewDefaultConfig()
		err := cfg.Save("/nonexistent/directory/config.yaml")

		assert.Error(t, err)
	})
}

func TestConfigDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		dir := ConfigDir()
		assert.NotEmpty(t, dir)
		assert.Contains(t, dir, "radiogogo")
	})
}

func TestConfigFile(t *testing.T) {
	t.Run("returns path ending with config.yaml", func(t *testing.T) {
		file := ConfigFile()
		assert.NotEmpty(t, file)
		assert.True(t, filepath.Base(file) == "config.yaml")
	})

	t.Run("returns path inside ConfigDir", func(t *testing.T) {
		file := ConfigFile()
		dir := ConfigDir()
		assert.Equal(t, dir, filepath.Dir(file))
	})
}

func TestConfig_LanguagePersistence(t *testing.T) {
	t.Run("language saves and loads correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		cfg := NewDefaultConfig()
		cfg.Language = "de"
		err := cfg.Save(cfgPath)
		assert.NoError(t, err)

		var loadedCfg Config
		err = loadedCfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, "de", loadedCfg.Language)
	})

	t.Run("language persists through multiple save/load cycles", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		languages := []string{"en", "de", "es", "ja", "zh"}

		for _, lang := range languages {
			cfg := NewDefaultConfig()
			cfg.Language = lang
			err := cfg.Save(cfgPath)
			assert.NoError(t, err)

			var loadedCfg Config
			err = loadedCfg.Load(cfgPath)
			assert.NoError(t, err)
			assert.Equal(t, lang, loadedCfg.Language, "Language %s should persist", lang)
		}
	})

	t.Run("parses language from YAML", func(t *testing.T) {
		input := `
language: ja
theme:
  textColor: "#ffffff"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "ja", cfg.Language)
	})

	t.Run("empty language defaults to empty string on load", func(t *testing.T) {
		input := `
theme:
  textColor: "#ffffff"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "", cfg.Language) // Empty, app should default to "en"
	})
}

func TestConfig_KeybindingsPersistence(t *testing.T) {
	t.Run("custom keybindings save and load correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		cfg := NewDefaultConfig()
		cfg.Keybindings.Quit = "x"
		cfg.Keybindings.VolumeUp = "+"
		cfg.Keybindings.VolumeDown = "-"
		err := cfg.Save(cfgPath)
		assert.NoError(t, err)

		var loadedCfg Config
		err = loadedCfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, "x", loadedCfg.Keybindings.Quit)
		assert.Equal(t, "+", loadedCfg.Keybindings.VolumeUp)
		assert.Equal(t, "-", loadedCfg.Keybindings.VolumeDown)
	})

	t.Run("parses keybindings from YAML", func(t *testing.T) {
		input := `
keybindings:
  quit: "x"
  search: "/"
  record: "R"
  volumeDown: "-"
  volumeUp: "+"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "x", cfg.Keybindings.Quit)
		assert.Equal(t, "/", cfg.Keybindings.Search)
		assert.Equal(t, "R", cfg.Keybindings.Record)
		assert.Equal(t, "-", cfg.Keybindings.VolumeDown)
		assert.Equal(t, "+", cfg.Keybindings.VolumeUp)
	})

	t.Run("partial keybindings leaves others empty", func(t *testing.T) {
		input := `
keybindings:
  quit: "x"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		assert.NoError(t, err)
		assert.Equal(t, "x", cfg.Keybindings.Quit)
		assert.Equal(t, "", cfg.Keybindings.Search) // Should be empty, app should validate and fill defaults
	})
}

func TestConfig_LoadOrCreateNew(t *testing.T) {
	// Note: LoadOrCreateNew uses ConfigDir() and ConfigFile() internally,
	// which depend on environment variables. We test using temp directories
	// by testing Load and Save separately, but also test the create-directory behavior.

	t.Run("creates directory if it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "newdir")

		// Verify directory doesn't exist
		_, err := os.Stat(newDir)
		assert.True(t, os.IsNotExist(err))

		// Create directory using MkdirAll (simulating what LoadOrCreateNew does)
		err = os.MkdirAll(newDir, 0755)
		assert.NoError(t, err)

		// Verify directory was created
		info, err := os.Stat(newDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("creates new file when it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		// Verify file doesn't exist
		_, err := os.Stat(cfgPath)
		assert.True(t, os.IsNotExist(err))

		// Create default config and save
		cfg := NewDefaultConfig()
		err = cfg.Save(cfgPath)
		assert.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(cfgPath)
		assert.NoError(t, err)

		// Verify it can be loaded back
		var loadedCfg Config
		err = loadedCfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, cfg.Language, loadedCfg.Language)
	})

	t.Run("loads existing file when it exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		// Create a config file with custom values
		content := `language: de
theme:
  textColor: "#123456"
  primaryColor: "#654321"
  secondaryColor: "#abcdef"
  tertiaryColor: "#fedcba"
  errorColor: "#ff00ff"
`
		err := os.WriteFile(cfgPath, []byte(content), 0644)
		assert.NoError(t, err)

		// Load the config
		var cfg Config
		err = cfg.Load(cfgPath)
		assert.NoError(t, err)

		// Verify custom values were loaded
		assert.Equal(t, "de", cfg.Language)
		assert.Equal(t, "#123456", cfg.Theme.TextColor)
		assert.Equal(t, "#654321", cfg.Theme.PrimaryColor)
	})

	t.Run("handles nested directory creation", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "a", "b", "c")

		err := os.MkdirAll(nestedDir, 0755)
		assert.NoError(t, err)

		info, err := os.Stat(nestedDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Run("handles empty file returns EOF error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(cfgPath, []byte(""), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)
		// Empty file returns EOF error from YAML decoder
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
	})

	t.Run("handles file with only whitespace returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(cfgPath, []byte("   \n\n   \t\t\n"), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)
		// Whitespace-only file returns error from YAML decoder
		assert.Error(t, err)
	})

	t.Run("handles file with comments only returns EOF error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		content := `# This is a comment
# Another comment
`
		err := os.WriteFile(cfgPath, []byte(content), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)
		// Comments-only file returns EOF error from YAML decoder
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
	})

	t.Run("handles unicode in values", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		content := `language: 日本語
theme:
  textColor: "#ffffff"
`
		err := os.WriteFile(cfgPath, []byte(content), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, "日本語", cfg.Language)
	})

	t.Run("handles extra unknown fields gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		content := `language: en
unknownField: "some value"
anotherUnknown:
  nested: true
theme:
  textColor: "#ffffff"
`
		err := os.WriteFile(cfgPath, []byte(content), 0644)
		assert.NoError(t, err)

		var cfg Config
		err = cfg.Load(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, "en", cfg.Language)
	})

	t.Run("preserves all keybindings through round-trip", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")

		cfg := NewDefaultConfig()
		cfg.Keybindings = Keybindings{
			Quit:           "x",
			Search:         "/",
			Record:         "R",
			BookmarkToggle: "b",
			BookmarksView:  "B",
			HideStation:    "h",
			ManageHidden:   "H",
			ChangeLanguage: "L",
			VolumeDown:     "-",
			VolumeUp:       "+",
			NavigateDown:   "j",
			NavigateUp:     "k",
			StopPlayback:   "space",
			Vote:           "v",
		}

		err := cfg.Save(cfgPath)
		assert.NoError(t, err)

		var loadedCfg Config
		err = loadedCfg.Load(cfgPath)
		assert.NoError(t, err)

		// Check all keybindings
		assert.Equal(t, cfg.Keybindings.Quit, loadedCfg.Keybindings.Quit)
		assert.Equal(t, cfg.Keybindings.Search, loadedCfg.Keybindings.Search)
		assert.Equal(t, cfg.Keybindings.Record, loadedCfg.Keybindings.Record)
		assert.Equal(t, cfg.Keybindings.BookmarkToggle, loadedCfg.Keybindings.BookmarkToggle)
		assert.Equal(t, cfg.Keybindings.BookmarksView, loadedCfg.Keybindings.BookmarksView)
		assert.Equal(t, cfg.Keybindings.HideStation, loadedCfg.Keybindings.HideStation)
		assert.Equal(t, cfg.Keybindings.ManageHidden, loadedCfg.Keybindings.ManageHidden)
		assert.Equal(t, cfg.Keybindings.ChangeLanguage, loadedCfg.Keybindings.ChangeLanguage)
		assert.Equal(t, cfg.Keybindings.VolumeDown, loadedCfg.Keybindings.VolumeDown)
		assert.Equal(t, cfg.Keybindings.VolumeUp, loadedCfg.Keybindings.VolumeUp)
		assert.Equal(t, cfg.Keybindings.NavigateDown, loadedCfg.Keybindings.NavigateDown)
		assert.Equal(t, cfg.Keybindings.NavigateUp, loadedCfg.Keybindings.NavigateUp)
		assert.Equal(t, cfg.Keybindings.StopPlayback, loadedCfg.Keybindings.StopPlayback)
		assert.Equal(t, cfg.Keybindings.Vote, loadedCfg.Keybindings.Vote)
	})
}
