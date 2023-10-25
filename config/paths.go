package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the path to the directory where the application's configuration files are stored.
// On Windows, the directory is %LOCALAPPDATA%\radiogogo.
// On other platforms, the directory is ~/.config/radiogogo.
func ConfigDir() string {
	var cfgDir string
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		cfgDir = filepath.Join(localAppData, "radiogogo")
	} else {
		home := os.Getenv("HOME")
		cfgDir = filepath.Join(home, ".config", "radiogogo")
	}
	return cfgDir
}

// ConfigFile returns the path to the configuration file.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}
