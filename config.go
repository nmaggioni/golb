package golb

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type configTOML struct {
	IP        string     `toml:"ip"`
	Port      string     `toml:"port"`
	Strategy  string     `toml:"strategy"`
	Timeout   int        `toml:"timeout"`
	Verbose   bool       `toml:"verbose"`
	Upstreams []upstream `toml:"upstream"`
	MaxCycles  int        `toml:"maxCycles"`
}

type upstream struct {
	Name string `toml:"name"`
	IP   string `toml:"ip"`
	Port string `toml:"port"`
}

var config configTOML

func ParseConfig(path string) error {
	if _, err := toml.DecodeFile(path, &config); err != nil {
		if err != nil {
			return err
		}
	}
	return nil
}

func FindConfigPath(path string) (string, error) {
	// Given path
	if path != "" {
		stats, err := os.Stat(path)
		if err == nil {
			if !stats.IsDir() {
				return path, nil
			}
			return "", errors.New("the given path is not a valid file")
		}
	}

	// File in current directory
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	confPath := filepath.Join(pwd, "config.toml")
	_, err = os.Stat(confPath)
	if err == nil {
		return confPath, nil
	}

	// File in UNIX config path
	confPath = "/etc/golb/config.toml"
	_, err = os.Stat(confPath)
	if err == nil {
		return confPath, nil
	}

	return "", errors.New("none found automatically, specify one manually")
}

func SetVerbose(value bool) {
	config.Verbose = value
}
