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

package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Language          string            `yaml:"language"`
	Theme             Theme             `yaml:"theme"`
	Keybindings       Keybindings       `yaml:"keybindings"`
	PlayerPreferences PlayerPreferences `yaml:"playerPreferences"`
}

// PlayerPreferences holds user preferences for the audio player.
type PlayerPreferences struct {
	// DefaultVolume is the initial volume level (0-100) when starting the application.
	// If not set or out of range, defaults to 80.
	DefaultVolume int `yaml:"defaultVolume"`
}

// Theme holds the color configuration for the UI.
type Theme struct {
	TextColor      string `yaml:"textColor"`
	PrimaryColor   string `yaml:"primaryColor"`
	SecondaryColor string `yaml:"secondaryColor"`
	TertiaryColor  string `yaml:"tertiaryColor"`
	ErrorColor     string `yaml:"errorColor"`
}

// NewDefaultConfig returns a Config struct with default values for RadioGoGo.
func NewDefaultConfig() Config {
	return Config{
		Language: "en",
		Theme: Theme{
			TextColor:      "#ffffff",
			PrimaryColor:   "#5a4f9f",
			SecondaryColor: "#8b77db",
			TertiaryColor:  "#4e4e4e",
			ErrorColor:     "#ff0000",
		},
		Keybindings:       NewDefaultKeybindings(),
		PlayerPreferences: NewDefaultPlayerPreferences(),
	}
}

// NewDefaultPlayerPreferences returns PlayerPreferences with sensible defaults.
func NewDefaultPlayerPreferences() PlayerPreferences {
	return PlayerPreferences{
		DefaultVolume: 80,
	}
}

// ValidateAndNormalize ensures PlayerPreferences values are within valid ranges.
// Returns the normalized preferences.
func (p PlayerPreferences) ValidateAndNormalize() PlayerPreferences {
	normalized := p
	// Clamp volume to valid range (0-100)
	if normalized.DefaultVolume < 0 {
		normalized.DefaultVolume = 0
	} else if normalized.DefaultVolume > 100 {
		normalized.DefaultVolume = 100
	}
	return normalized
}

// Load reads the configuration file from the given path and decodes it into the Config struct.
// It returns an error if the file cannot be opened or if there is an error decoding the file.
func (c *Config) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the configuration to a file at the given path.
// It returns an error if the file cannot be created or if there is an error encoding the configuration.
func (c Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return err
	}

	return nil
}

// LoadOrCreateNew loads the configuration file if it exists, or creates a new one if it doesn't.
// It returns an error if it fails to create the directory or load/save the configuration file.
func (c *Config) LoadOrCreateNew() error {

	err := os.MkdirAll(ConfigDir(), 0755)

	if err != nil {
		return err
	}

	cfgFile := ConfigFile()

	if _, err := os.Stat(cfgFile); errors.Is(err, os.ErrNotExist) {
		err := c.Save(cfgFile)
		if err != nil {
			return err
		}
	} else {
		err := c.Load(cfgFile)
		if err != nil {
			return err
		}
	}

	return nil

}
