package impltests

import (
	"encoding/pem"
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"
)

func execAndCapture(cmdParts ...string) (string, error) {
	return execAndCaptureInDir("", cmdParts...)
}

func execAndCaptureInDir(dir string, cmdParts ...string) (string, error) {
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Dir = dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func cmdWaitWithTimeout(cmd *exec.Cmd, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	cmdDone := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		cmdDone <- err
	}()

	select {
	case <-timer.C:
		cmd.Process.Kill()
		err := fmt.Errorf("command did not complete before timeout: %v", d)
		return err
	case err := <-cmdDone:
		return err
	}
}

func testExec(ctx *test_executor.ExecutionContext, getCommand func(caPath string, hostname string, tlsPort uint) []string) (map[string]*test_executor.SuiteTestResults, error) {
	return testExecDir(ctx, "", getCommand)
}

func testExecDir(ctx *test_executor.ExecutionContext, workingDir string, getCommand func(caPath string, hostname string, tlsPort uint) []string) (map[string]*test_executor.SuiteTestResults, error) {
	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return nil, err
	}

	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	_, err = tmpFile.Write(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: suites.GetRootCert().Raw}))
	tmpFile.Close()
	if err != nil {
		return nil, err
	}

	caPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	return test_executor.ExecuteAllTestsRemote(ctx, suites, func(hostname string, port uint) (bool, error) {
		cmdParts := getCommand(caPath, hostname, port)
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		if workingDir != "" {
			cmd.Dir = workingDir
		}

		err = cmd.Start()
		if err != nil {
			return false, err
		}
		err = cmd.Wait()
		return err == nil, nil
	})
}
