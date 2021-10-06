package impltests

import (
	"encoding/pem"
	"fmt"
	"github.com/Netflix/bettertls/pathbuilding"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
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

func cmdWaitWithTimeout(t *testing.T, cmd *exec.Cmd, d time.Duration) error {
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
		t.Fatal(err)
		return err
	case err := <-cmdDone:
		return err
	}
}

func testExec(t *testing.T, getCommand func(caPath string, testCaseName string, tlsPort int) []string) {
	testExecDir(t, "", getCommand)
}

func testExecDir(t *testing.T, workingDir string, getCommand func(caPath string, testCaseName string, tlsPort int) []string) {
	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		t.Fatal(err)
	}

	server, err := pathbuilding.StartServer(provider, noplog.Logger, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	manifest, err := provider.GetManifest()
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "ca.pem")
	err = ioutil.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: manifest.Root}), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = pathbuilding.ExecuteTests(t, provider, func(testCaseName string) (bool, error) {
		cmdParts := getCommand(caPath, testCaseName, server.TlsPort())
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
	if err != nil {
		t.Fatal(err)
	}
}
