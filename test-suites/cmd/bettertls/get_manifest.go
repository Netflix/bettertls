package main

import (
	"encoding/json"
	"flag"
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"os"
)

type Manifest struct {
	BetterTlsRevision string                    `json:"betterTlsRevision"`
	SuiteManifests    map[string]*SuiteManifest `json:"suiteManifests"`
}

type SuiteManifest struct {
	Features        map[test_case.Feature]string `json:"features"`
	ExpectedResults []test_case.ExpectedResult   `json:"expectedResults"`
}

func getSuiteManifest(provider test_case.TestCaseProvider) (*SuiteManifest, error) {
	manifest := &SuiteManifest{
		Features: make(map[test_case.Feature]string),
	}
	for _, feature := range provider.GetFeatures() {
		manifest.Features[feature] = provider.DescribeFeature(feature)
	}
	testCaseCount, err := provider.GetTestCaseCount()
	if err != nil {
		return nil, err
	}
	for idx := uint(0); idx < testCaseCount; idx++ {
		testCase, err := provider.GetTestCase(idx)
		if err != nil {
			return nil, err
		}
		manifest.ExpectedResults = append(manifest.ExpectedResults, testCase.ExpectedResult())
	}
	return manifest, nil
}

func generateManifests(args []string) error {
	flagSet := flag.NewFlagSet("generate-manifests", flag.ContinueOnError)
	var outFile string
	flagSet.StringVar(&outFile, "out", "", "Write manifest to file instead of stdout.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	manifest, err := buildManifest()
	if err != nil {
		return err
	}

	var writer *json.Encoder
	if outFile == "" {
		writer = json.NewEncoder(os.Stdout)
		writer.SetIndent("", "   ")
	} else {
		f, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file for saving results: %v", err)
		}
		defer f.Close()
		writer = json.NewEncoder(f)
	}
	err = writer.Encode(manifest)
	if err != nil {
		return fmt.Errorf("failed to save results: %v", err)
	}

	return nil
}

func buildManifest() (*Manifest, error) {
	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{
		BetterTlsRevision: test_executor.GetBuildRevision(),
		SuiteManifests:    make(map[string]*SuiteManifest),
	}
	for _, suiteName := range suites.GetProviderNames() {
		provider := suites.GetProvider(suiteName)
		suiteManifest, err := getSuiteManifest(provider)
		if err != nil {
			return nil, err
		}
		manifest.SuiteManifests[suiteName] = suiteManifest
	}
	return manifest, nil
}
