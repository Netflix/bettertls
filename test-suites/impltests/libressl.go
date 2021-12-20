package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"os"
	"path/filepath"
	"strings"
)

type LibresslRunner struct {
	libresslPath string
	version      string
}

func (l *LibresslRunner) Name() string {
	return "libressl"
}

func (l *LibresslRunner) Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	l.libresslPath = filepath.Join(homeDir, "src", "libressl-3.4.0", "apps", "openssl", "openssl")

	version, err := execAndCapture(l.libresslPath, "version")
	if err != nil {
		return err
	}
	l.version = version

	return nil
}

func (l *LibresslRunner) Close() error {
	return nil
}

func (l *LibresslRunner) GetVersion() string {
	return l.version
}

func (l *LibresslRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{"bash", "-c", strings.Join([]string{
			l.libresslPath, "s_client",
			"-CAfile", caPath,
			"-connect", fmt.Sprintf("%s:%d", hostname, tlsPort),
			"-verify_return_error",
			"|", "grep", "\"Verify return code: 0\"",
		}, " "),
		}
	})
}
