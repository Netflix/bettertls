package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

type BotanRunner struct {
	version string
}

func (o *BotanRunner) Name() string {
	return "botan"
}

func (o *BotanRunner) Initialize() error {
	var err error
	o.version, err = execAndCapture("botan", "version", "--full")
	if err != nil {
		return err
	}
	return nil
}

func (o *BotanRunner) Close() error {
	return nil
}

func (o *BotanRunner) GetVersion() string {
	return o.version
}

func (o *BotanRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		args := []string{"botan", "tls_client",
			"--skip-system-cert-store",
			fmt.Sprintf("--trusted-cas=%s", caPath),
			fmt.Sprintf("--port=%d", tlsPort),
			hostname}

		return args
	})
}
