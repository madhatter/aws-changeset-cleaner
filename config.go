package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

// CleanerConfig is public collection of necessary settings
type CleanerConfig struct {
	profile      string
	processAll   bool
	stackToClean string
	keep         int
	verbose      bool
}

func NewCleanerConfig() *CleanerConfig {
	return &CleanerConfig{
		profile:      "",
		processAll:   true,
		stackToClean: "",
		keep:         0,
		verbose:      false,
	}
}

// parsecliarguments parses the command line arguments and adds them to the config
func (config *CleanerConfig) parseCLIArguments() {
	flag.StringVar(&config.profile, "profile", "", "AWS profile to use. (Required)")
	flag.StringVar(&config.stackToClean, "stack", "all", "Stack to clean {all stacks|<stackname>}.")
	flag.IntVar(&config.keep, "keep", 10, "Number of changesets to keep.")
	flag.BoolVar(&config.verbose, "verbose", false, "Verbose logging.")
	flag.Parse()

	if config.profile == "" {
		fmt.Println("No profile set.")
		flag.PrintDefaults()
		os.Exit(3)
	}

	if config.stackToClean == "all" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println()
		fmt.Print("Processing on all stacks. Deleting all failed changesets on _all_ stacks. Continue (y/n)? ")
		text, _ := reader.ReadString('\n')
		if text != "y\n" {
			fmt.Println("Coward.")
			os.Exit(3)
		}
	} else {
		config.processAll = false
	}
}
