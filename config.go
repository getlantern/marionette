package marionette

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents a Marionette configuration file.
type Config struct {
	General struct {
		Debug        bool   `toml:"debug"`
		AutoUpdate   bool   `toml:"autoupdate"`
		UpdateServer string `toml:"update_server"`
		Format       string `toml:"format"`
	} `toml:"general"`

	Client struct {
		IP   string `toml:"client_ip"`
		Port int    `toml:"client_port"`
	} `toml:"client"`

	Server struct {
		IP        string `toml:"server_ip"`
		ProxyIP   string `toml:"proxy_ip"`
		ProxyPort int    `toml:"proxy_port"`
	} `toml:"server"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	var config Config
	config.General.Format = "dummy"
	config.Client.ClientIP = "127.0.0.1"
	config.Client.ClientPort = 8079
	config.Server.ServerIP = "127.0.0.1"
	config.Server.ProxyIP = "127.0.0.1"
	config.Server.ProxyPort = 8081
	return config
}

// ParseConfig returns the first matching configuration file path.
// Searches the present working directory & /etc. If no configuration is
// found then the default configuration is returned.
func ParseConfig() (Config, error) {
	// Collect search paths.
	var paths []string
	if path, err := os.Getwd(); err == "" {
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
		return err
	}
	return config, nil
}
