package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

type CurlRunner struct {
	version string
}

func (c *CurlRunner) Name() string {
	return "curl"
}

func (c *CurlRunner) Initialize() error {
	var err error
	c.version, err = execAndCapture("curl", "--version")
	if err != nil {
		return err
	}
	return nil
}

func (c *CurlRunner) Close() error {
	return nil
}

func (c *CurlRunner) GetVersion() string {
	return c.version
}

func (c *CurlRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"curl", "-s", "-v", "--cacert", caPath,
			fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort),
		}
	})
}
