package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Netflix/bettertls/test-suites/impltests"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"github.com/golang/protobuf/proto"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"time"
)

type ImplementationTestResults struct {
	ImplementationInfo string            `json:"implementation"`
	VersionInfo        string            `json:"version"`
	Date               time.Time         `json:"date"`
	BetterTlsRevision  string            `json:"betterTlsRevision"`
	Suites             map[string][]byte `json:"suites"`
}

func runTests(args []string) error {
	flagSet := flag.NewFlagSet("run-tests", flag.ContinueOnError)
	var implementation string
	flagSet.StringVar(&implementation, "implementation", "", "Implementation to test.")
	var suite string
	flagSet.StringVar(&suite, "suite", "", "Run only the given suite instead of all suites.")
	var testCase int
	flagSet.IntVar(&testCase, "testCase", -1, "Run only the given test case in the suite instead of all tests. Requires --suite to be sepecified as well.")
	var outputDir string
	flagSet.StringVar(&outputDir, "outputDir", ".", "Directory to which test results will be written.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	manifest, err := buildManifest()
	if err != nil {
		return err
	}

	var runners []impltests.ImplementationRunner
	if implementation == "" {
		for _, runner := range impltests.Runners {
			runners = append(runners, runner)
		}
	} else {
		runner := impltests.Runners[implementation]
		if runner == nil {
			return fmt.Errorf("invalid implementation: %s", implementation)
		}
		runners = []impltests.ImplementationRunner{runner}
	}

	for _, runner := range runners {
		err := runner.Initialize()
		if err != nil {
			return fmt.Errorf("failed to initialize runner %s: %v", runner.Name(), err)
		}

		var bar *progressbar.ProgressBar
		ctx := &test_executor.ExecutionContext{
			RunOnlySuite: suite,
			RunOnlyTest:  testCase,
			OnStartSuite: func(suite string, testCount uint) {
				bar = progressbar.Default(int64(testCount), runner.Name()+"/"+suite)
				progressbar.OptionSetItsString("tests")(bar)
			},
			OnStartTest: func(idx uint) {
				bar.Add(1)
			},
		}

		version := runner.GetVersion()
		suiteResults, err := runner.RunTests(ctx)
		if err != nil {
			return fmt.Errorf("error running tests: %v", err)
		}

		suiteResultsEncoded := make(map[string][]byte, len(suiteResults))
		for suiteName, result := range suiteResults {
			resultBytes, err := proto.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to proto-marshal results: %v", err)
			}
			buffer := bytes.NewBuffer(nil)
			gz := gzip.NewWriter(buffer)
			_, err = gz.Write(resultBytes)
			if err != nil {
				return fmt.Errorf("failed to gzip results: %v", err)
			}
			if err = gz.Flush(); err != nil {
				return fmt.Errorf("failed to gzip results: %v", err)
			}
			if err = gz.Close(); err != nil {
				return fmt.Errorf("failed to gzip results: %v", err)
			}
			suiteResultsEncoded[suiteName] = buffer.Bytes()
		}

		results := &ImplementationTestResults{
			ImplementationInfo: runner.Name(),
			VersionInfo:        version,
			Date:               time.Now(),
			BetterTlsRevision:  test_executor.GetBuildRevision(),
			Suites:             suiteResultsEncoded,
		}

		f, err := os.OpenFile(filepath.Join(outputDir, fmt.Sprintf("%s_results.json", runner.Name())),
			os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file for saving results: %v", err)
		}
		err = json.NewEncoder(f).Encode(results)
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to save results: %v", err)
		}

		if summary, err := buildSummary(results, manifest); err == nil {
			printSummary(summary)
		}
	}

	return nil
}
