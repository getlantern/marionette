package marionette

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents a Marionette configuration file.
type Config struct {
	General struct {
		Debug  bool   `toml:"debug"`
		Format string `toml:"format"`
	} `toml:"general"`

	Client struct {
		Bind string `toml:"bind"`
	} `toml:"client"`

	Server struct {
		IP    string `toml:"ip"`
		Bind  string `toml:"bind"`
		Proxy string `toml:"proxy"`
	} `toml:"server"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	var config Config
	config.General.Format = "dummy"
	config.Client.Bind = "127.0.0.1:8079"
	config.Server.IP = "127.0.0.1"
	config.Server.Bind = ""
	config.Server.Proxy = "127.0.0.1:8081"
	return config
}

// ParseConfig returns the first matching configuration file path.
// Searches the present working directory & /etc. If no configuration is
// found then the default configuration is returned.
func ParseConfig() (Config, error) {
	// Collect search paths.
	var paths []string
	if path, err := os.Getwd(); err == nil {
		paths = append(paths, path)
	}
	paths = append(paths, "/etc")

	// Iterate over each search path.
	for _, path := range paths {
		// Attempt to parse "marionette.conf".
		if config, err := ParseConfigFile(filepath.Join(path, "marionette.conf")); err != nil && !os.IsNotExist(err) {
			return Config{}, err
		} else if err == nil {
			return config, nil
		}

		// Attempt to parse "marionette_tg/marionette.conf".
		if config, err := ParseConfigFile(filepath.Join(path, "marionette_tg", "marionette.conf")); err != nil && !os.IsNotExist(err) {
			return Config{}, err
		} else if err == nil {
			return config, nil
		}
	}

	// Return default configuration if no files were found.
	return DefaultConfig(), nil
}

// ParseConfigFile parses a configuration file at path.
func ParseConfigFile(path string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}
