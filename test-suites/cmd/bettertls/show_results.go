package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func showResults(args []string) error {
	flagSet := flag.NewFlagSet("show-results", flag.ContinueOnError)
	var resultsPath string
	flagSet.StringVar(&resultsPath, "resultsFile", "", "Path to file with results to display.")
	var manifestPath string
	flagSet.StringVar(&manifestPath, "manifestFile", "", "Path to file with manifest containing expected test results. Will use a manifest for the current revision if unspecified.")
	var jsonFormat bool
	flagSet.BoolVar(&jsonFormat, "json", false, "Output results as JSON instead of human-readable summary.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	if resultsPath == "" {
		return fmt.Errorf("missing required parameter: --resultsFile")
	}

	var manifest *Manifest
	if manifestPath == "" {
		manifest, err = buildManifest()
		if err != nil {
			return err
		}
	} else {
		f, err := os.Open(manifestPath)
		if err != nil {
			return err
		}
		defer f.Close()
		manifest = new(Manifest)
		err = json.NewDecoder(f).Decode(manifest)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(resultsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	results := new(ImplementationTestResults)
	err = json.NewDecoder(f).Decode(results)
	if err != nil {
		return fmt.Errorf("failed to parse results file: %v", err)
	}
	summary, err := buildSummary(results, manifest)
	if err != nil {
		return err
	}
	if jsonFormat {
		err = json.NewEncoder(os.Stdout).Encode(summary)
	} else {
		err = printSummary(summary)
	}
	if err != nil {
		return err
	}
	return nil
}
