package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"os"
	"os/exec"
)

type RustlsRunner struct {
	tmpDir  string
	version string
}

func (r *RustlsRunner) Name() string {
	return "rustls"
}

func (r *RustlsRunner) Initialize() error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", "https://github.com/rustls/rustls.git", tmpDir)
	err = cmd.Run()
	if err != nil {
		return err
	}
	version, err := execAndCaptureInDir(tmpDir, "cargo", "tree", "-e", "normal")
	if err != nil {
		return err
	}

	r.tmpDir = tmpDir
	r.version = version
	return nil
}

func (r *RustlsRunner) Close() error {
	if r.tmpDir != "" {
		return os.RemoveAll(r.tmpDir)
	}
	return nil
}

func (r *RustlsRunner) GetVersion() string {
	return r.version
}

func (r *RustlsRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExecDir(ctx, r.tmpDir, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"cargo", "run", "--example", "tlsclient", "--",
			"--cafile", caPath,
			"-p", fmt.Sprintf("%d", tlsPort),
			"--http", hostname,
		}
	})
}
