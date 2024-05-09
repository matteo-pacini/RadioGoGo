package config

import (
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
