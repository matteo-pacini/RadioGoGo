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
