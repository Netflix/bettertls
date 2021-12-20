package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
)

type resultsSummary struct {
	Implementation string                   `json:"implementation"`
	Version        string                   `json:"version"`
	SuiteSummary   map[string]*suiteSummary `json:"suiteSummary"`
}

type suiteSummary struct {
	SupportedFeatures   []string `json:"supportedFeatures"`
	UnsupportedFeatures []string `json:"unsupportedFeatures"`

	PassedTests        []uint `json:"passedTests"`
	WarningTests       []uint `json:"warningTests"`
	SkippedTests       []uint `json:"skippedTests"`
	FalsePositiveTests []uint `json:"falsePositiveTests"`
	FalseNegativeTests []uint `json:"falseNegativeTests"`
}

func buildSummary(results *ImplementationTestResults, manifest *Manifest) (*resultsSummary, error) {
	summary := &resultsSummary{
		Implementation: results.ImplementationInfo,
		Version:        results.VersionInfo,
		SuiteSummary:   make(map[string]*suiteSummary),
	}

	for suiteName, suiteResultsEncoded := range results.Suites {
		suiteManifest := manifest.SuiteManifests[suiteName]
		if suiteManifest == nil {
			return nil, fmt.Errorf("no manifest for suite %s", suiteName)
		}
		gz, err := gzip.NewReader(bytes.NewReader(suiteResultsEncoded))
		if err != nil {
			return nil, err
		}
		resultsBytes, err := ioutil.ReadAll(gz)
		if err != nil {
			return nil, err
		}
		suiteResults := new(test_executor.SuiteTestResults)
		err = proto.Unmarshal(resultsBytes, suiteResults)
		if err != nil {
			return nil, err
		}

		suiteSummary := new(suiteSummary)
		for _, f := range suiteResults.SupportedFeatures {
			suiteSummary.SupportedFeatures = append(suiteSummary.SupportedFeatures, suiteManifest.Features[test_case.Feature(f)])
		}
		for _, f := range suiteResults.UnsupportedFeatures {
			suiteSummary.UnsupportedFeatures = append(suiteSummary.UnsupportedFeatures, suiteManifest.Features[test_case.Feature(f)])
		}

		for testCaseId, result := range suiteResults.TestCaseResults {
			expectedResult := suiteManifest.ExpectedResults[testCaseId]
			if result == test_executor.TestCaseResult_ACCEPTED {
				if expectedResult == test_case.EXPECTED_RESULT_PASS || expectedResult == test_case.EXPECTED_RESULT_SOFT_PASS {
					suiteSummary.PassedTests = append(suiteSummary.PassedTests, uint(testCaseId))
				}
				if expectedResult == test_case.EXPECTED_RESULT_SOFT_FAIL {
					suiteSummary.WarningTests = append(suiteSummary.WarningTests, uint(testCaseId))
				}
				if expectedResult == test_case.EXPECTED_RESULT_FAIL {
					suiteSummary.FalseNegativeTests = append(suiteSummary.FalseNegativeTests, uint(testCaseId))
				}
			}
			if result == test_executor.TestCaseResult_REJECTED {
				if expectedResult == test_case.EXPECTED_RESULT_FAIL || expectedResult == test_case.EXPECTED_RESULT_SOFT_FAIL {
					suiteSummary.PassedTests = append(suiteSummary.PassedTests, uint(testCaseId))
				}
				if expectedResult == test_case.EXPECTED_RESULT_SOFT_PASS {
					suiteSummary.WarningTests = append(suiteSummary.WarningTests, uint(testCaseId))
				}
				if expectedResult == test_case.EXPECTED_RESULT_PASS {
					suiteSummary.FalsePositiveTests = append(suiteSummary.FalsePositiveTests, uint(testCaseId))
				}
			}
			if result == test_executor.TestCaseResult_SKIPPED {
				suiteSummary.SkippedTests = append(suiteSummary.SkippedTests, uint(testCaseId))
			}
		}

		summary.SuiteSummary[suiteName] = suiteSummary
	}

	return summary, nil
}

func printSummary(summary *resultsSummary) error {
	fmt.Printf("Implementation: %s\n", summary.Implementation)
	fmt.Printf("Version: %s\n", summary.Version)
	for suiteName, suiteSummary := range summary.SuiteSummary {
		fmt.Printf("Suite: %s\n", suiteName)
		fmt.Printf("  Supported Features: ")
		for _, feature := range suiteSummary.SupportedFeatures {
			fmt.Printf("%s, ", feature)
		}
		fmt.Println()

		fmt.Printf("  Unsupported Features: ")
		for _, feature := range suiteSummary.UnsupportedFeatures {
			fmt.Printf("%s, ", feature)
		}
		fmt.Println()

		fmt.Printf("  Passed: %d\n", len(suiteSummary.PassedTests)+len(suiteSummary.WarningTests))
		if len(suiteSummary.WarningTests) > 0 {
			fmt.Printf("    Passed with warnings: %d\n", len(suiteSummary.WarningTests))
		}
		fmt.Printf("  Skipped: %d\n", len(suiteSummary.SkippedTests))
		fmt.Printf("  Failures: %d\n", len(suiteSummary.FalsePositiveTests)+len(suiteSummary.FalseNegativeTests))
		if len(suiteSummary.FalsePositiveTests) > 0 {
			fmt.Printf("    False positives: %d\n", len(suiteSummary.FalsePositiveTests))
		}
		if len(suiteSummary.FalseNegativeTests) > 0 {
			fmt.Printf("    False negatives: %d\n", len(suiteSummary.FalseNegativeTests))
		}

		fmt.Println()
	}

	return nil
}
