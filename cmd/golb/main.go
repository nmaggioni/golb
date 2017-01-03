package main

import (
	"fmt"
	"github.com/nmaggioni/golb"
	"gopkg.in/alecthomas/kingpin.v2"
	"runtime"
)

var (
	verbose  = kingpin.Flag("verbose", "Enable verbose output (overrides the configuration value).").Default("false").Short('v').Bool()
	confPath = kingpin.Flag("config", "The path to the configuration file.").Short('c').PlaceHolder("PATH").ExistingFile()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	confPath, err := golb.FindConfigPath(*confPath)
	if err != nil {
		kingpin.FatalUsage("Unable to find a configuration file: %v.", err)
	}
	err = golb.ParseConfig(confPath)
	if err != nil {
		kingpin.Fatalf("Failed to decode the configuration file, check the TOML syntax:\n%v", err)
	}

	if *verbose {
		fmt.Printf("Loaded configuration from file: %s\n", confPath)
		golb.SetVerbose(true)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	golb.Listen()
}
