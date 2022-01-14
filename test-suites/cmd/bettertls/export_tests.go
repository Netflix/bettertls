package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Netflix/bettertls/test-suites/certutil"
	int_set "github.com/Netflix/bettertls/test-suites/int-set"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io"
	"os"
)

type testExport struct {
	BetterTlsRevision string                  `json:"betterTlsRevision"`
	TrustRoot         []byte                  `json:"trustRoot"`
	Suites            map[string]*suiteExport `json:"suites"`
}

type suiteExport struct {
	Features            []string          `json:"features"`
	SanityCheckTestCase uint              `json:"sanityCheckTestCase"`
	FeatureTestCases    map[string][]uint `json:"featureTestCases"`
	TestCases           []*testCaseExport `json:"testCases"`
}

type testCaseExport struct {
	Certificates     [][]byte `json:"certificates"`
	Hostname         string   `json:"hostname"`
	RequiredFeatures []string `json:"requiredFeatures"`
	Expected         string   `json:"expected"`
	FailureIsWarning bool     `json:"failureIsWarning"`
}

func exportTests(args []string) error {
	flagSet := flag.NewFlagSet("run-tests", flag.ContinueOnError)
	var suite string
	flagSet.StringVar(&suite, "suite", "", "Export only the given suite instead of all suites.")
	testCases := new(int_set.IntSet)
	flagSet.Var(testCases, "testCase", "Export only the given test case(s) in the suite instead of all tests. Requires --suite to be specified as well. Use \"123,456-789\" syntax to include a range or set of cases.")
	var outputPath string
	flagSet.StringVar(&outputPath, "out", "", "Write to the given file instead of stdout.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	rootCa, rootKey, err := certutil.GenerateSelfSignedCert("bettertls_trust_root")
	if err != nil {
		return err
	}
	suites, err := test_executor.BuildTestSuitesWithRootCa(rootCa, rootKey)
	if err != nil {
		return err
	}

	output := new(testExport)
	output.BetterTlsRevision = test_executor.GetBuildRevision()
	output.TrustRoot = rootCa.Raw
	output.Suites = make(map[string]*suiteExport)

	for _, suiteName := range suites.GetProviderNames() {
		if suite != "" && suiteName != suite {
			continue
		}
		provider := suites.GetProvider(suiteName)

		suiteExport := new(suiteExport)
		suiteExport.Features = make([]string, 0)
		suiteExport.SanityCheckTestCase, err = provider.GetSanityCheckTestCase()
		suiteExport.FeatureTestCases = make(map[string][]uint)
		for _, feature := range provider.GetFeatures() {
			featureName := provider.DescribeFeature(feature)
			suiteExport.Features = append(suiteExport.Features, featureName)
			testCases, err := provider.GetTestCasesForFeature(feature)
			if err != nil {
				return err
			}
			suiteExport.FeatureTestCases[featureName] = testCases
		}
		if err != nil {
			return err
		}
		testCaseCount, err := provider.GetTestCaseCount()
		if err != nil {
			return nil
		}
		for i := uint(0); i < testCaseCount; i++ {
			testCase, err := provider.GetTestCase(i)
			if err != nil {
				return err
			}
			testCaseExport := new(testCaseExport)
			certs, err := testCase.GetCertificates(rootCa, rootKey)
			if err != nil {
				return err
			}
			testCaseExport.Certificates = certs.Certificate
			testCaseExport.Hostname = testCase.GetHostname()
			testCaseExport.RequiredFeatures = make([]string, 0)
			for _, feature := range testCase.RequiredFeatures() {
				testCaseExport.RequiredFeatures = append(testCaseExport.RequiredFeatures, provider.DescribeFeature(feature))
			}
			switch testCase.ExpectedResult() {
			case test_case.EXPECTED_RESULT_PASS:
				testCaseExport.Expected = "ACCEPT"
			case test_case.EXPECTED_RESULT_FAIL:
				testCaseExport.Expected = "REJECT"
			case test_case.EXPECTED_RESULT_SOFT_PASS:
				testCaseExport.Expected = "ACCEPT"
				testCaseExport.FailureIsWarning = true
			case test_case.EXPECTED_RESULT_SOFT_FAIL:
				testCaseExport.Expected = "REJECT"
				testCaseExport.FailureIsWarning = true
			default:
				panic(fmt.Errorf("unhandled expected result: %v", testCase.ExpectedResult()))
			}

			suiteExport.TestCases = append(suiteExport.TestCases, testCaseExport)
		}

		output.Suites[suiteName] = suiteExport
	}

	var out io.Writer
	if outputPath == "" {
		out = os.Stdout
	} else {
		f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	err = json.NewEncoder(out).Encode(output)
	if err != nil {
		return fmt.Errorf("failed to save results: %v", err)
	}

	return nil
}
