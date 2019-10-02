package main

import (
	"flag"
	"fmt"
)

// CleanerConfig is public collection of necessary settings
type CleanerConfig struct {
	profile      string
	processAll   bool
	stackToClean string
	keep         int
	verbose      bool
	yesyesyes    bool
}

func NewCleanerConfig() *CleanerConfig {
	return &CleanerConfig{
		profile:      "",
		processAll:   false,
		stackToClean: "",
		keep:         0,
		verbose:      false,
		yesyesyes:    false,
	}
}

// parsecliarguments parses the command line arguments and adds them to the config
func (config *CleanerConfig) parseCLIArguments() {
	flag.StringVar(&config.profile, "profile", "", "AWS profile to use. (Required)")
	flag.StringVar(&config.stackToClean, "stack", "all", "Stack to clean {all stacks|<stackname>}.")
	flag.IntVar(&config.keep, "keep", 10, "Number of changesets to keep.")
	flag.BoolVar(&config.verbose, "verbose", false, "Verbose logging.")
	flag.BoolVar(&config.yesyesyes, "yes", false, "Don't bother me. Do it.")
	flag.Parse()

}

func (config *CleanerConfig) validate() error {
	if err := checkStringFlagNotEmpty("profile", config.profile); err != nil {
		return err
	}

	if err := checkStringFlagNotEmpty("stack", config.stackToClean); err != nil {
		return err
	}

	if config.stackToClean == "all" {
		config.processAll = true
	}

	return nil
}

func checkStringFlagNotEmpty(name string, f string) error {
	if f == "" {
		return fmt.Errorf("Missing mandatory parameter: %s", name)
	}
	return nil
}
