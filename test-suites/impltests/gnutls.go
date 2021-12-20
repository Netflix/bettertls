package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

type GnutlsRunner struct {
	version string
}

func (g *GnutlsRunner) Name() string {
	return "gnutls"
}

func (g *GnutlsRunner) Initialize() error {
	var err error
	g.version, err = execAndCapture("gnutls-cli", "--version")
	if err != nil {
		return err
	}
	return nil
}

func (g *GnutlsRunner) Close() error {
	return nil
}

func (g *GnutlsRunner) GetVersion() string {
	return g.version
}

func (g *GnutlsRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"gnutls-cli", "--x509cafile", caPath,
			fmt.Sprintf("%s:%d", hostname, tlsPort),
		}
	})
}
