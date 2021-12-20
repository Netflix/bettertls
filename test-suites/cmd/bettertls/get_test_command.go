package main

import (
	"encoding/json"
	"flag"
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

func getTest(args []string) error {
	flagSet := flag.NewFlagSet("get-test", flag.ContinueOnError)
	var providerName string
	flagSet.StringVar(&providerName, "suite", "", "Suite to run. One of \"pathbuilding\", \"nameconstraints\".")
	var testId uint
	flagSet.UintVar(&testId, "testId", 0, "Test id to describe.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return err
	}
	provider := suites.GetProvider(providerName)
	if provider == nil {
		return fmt.Errorf("invalid test suite: %s", providerName)
	}
	testCase, err := provider.GetTestCase(testId)
	if err != nil {
		return err
	}

	var output struct {
		Suite          string                   `json:"suite"`
		TestId         uint                     `json:"testId"`
		Definition     test_case.TestCase       `json:"definition"`
		ExpectedResult test_case.ExpectedResult `json:"expectedResult"`
		Certificates   [][]byte                 `json:"certificates"`
	}

	output.Suite = provider.Name()
	output.TestId = testId
	output.Definition = testCase
	output.ExpectedResult = testCase.ExpectedResult()
	certs, err := suites.GetTestCaseCertificates(testCase)
	if err != nil {
		return err
	}
	output.Certificates = certs.Certificate

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(outputBytes))
	return nil
}
