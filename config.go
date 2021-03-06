package golb

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
)

type configTOML struct {
	IP         string     `toml:"ip"`
	Port       string     `toml:"port"`
	Verbose    bool       `toml:"verbose"`
	Strategy   string     `toml:"strategy"`
	Sticky     bool       `toml:"sticky"`
	Stickiness int        `toml:"stickyness"`
	Timeout    int        `toml:"timeout"`
	MaxCycles  int        `toml:"maxCycles"`
	Upstreams  []upstream `toml:"upstream"`
}

type upstream struct {
	Name           string `toml:"name"`
	IP             string `toml:"ip"`
	Port           string `toml:"port"`
	MonitoringPort string `toml:"monitoring-port"`
	Weight         int    `toml:"weight"`
}

var config configTOML
var confPath string

func ParseConfig(path string) error {
	if _, err := toml.DecodeFile(path, &config); err != nil {
		if err != nil {
			return err
		}
	}
	confPath = path

	if config.Sticky {
		if config.Verbose {
			logrus.Info("Kickstarting GC")
		}
		go staleSessionsCleaner()
	}

	switch config.Strategy {
	case "weighted-round-robin":
		var weightedUpstreams []upstream
		for _, upstr := range config.Upstreams {
			for i := 0; i < upstr.Weight; i++ {
				weightedUpstreams = append(weightedUpstreams, upstr)
			}
		}
		config.Upstreams = weightedUpstreams
	case "active-polling":
		if config.Verbose {
			logrus.Info("Kickstarting poller")
		}
		polledLoads = make(map[string]float64)
		go func() {
			for {
				pollLoads()
				time.Sleep(3 * time.Second)
			}
		}()
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

func ConfigLoaded() {
	if config.Verbose {
		logrus.WithFields(logrus.Fields{
			"file": confPath,
		}).Info("Configuration loaded")
	}
}