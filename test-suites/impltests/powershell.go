package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

// Compatibility notes:
// This runner requires Administrator privileges so that it can add/remove root CA's.
// You must also have "test-suites/scripts/" in your PATH.

type PowerShellRunner struct {
	version string
}

func (c *PowerShellRunner) Name() string {
	return "powershell"
}

func (c *PowerShellRunner) Initialize() error {
	var err error
	c.version, err = execAndCapture("powershell", "$PSVersionTable.PSEdition + \" \" + $PSVersionTable.PSVersion")
	if err != nil {
		return err
	}
	return nil
}

func (c *PowerShellRunner) Close() error {
	return nil
}

func (c *PowerShellRunner) GetVersion() string {
	return c.version
}

func (c *PowerShellRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"powershell", "-ExecutionPolicy", "Unrestricted", "-Command", "try-tls-handshake.ps1", "-url", fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort), "-capath", caPath,
		}
	})
}
