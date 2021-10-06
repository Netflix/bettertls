package impltests

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestRustls(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "clone", "https://github.com/rustls/rustls.git", tmpDir)
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	version, err := execAndCaptureInDir(tmpDir, "cargo", "tree", "-e", "normal")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	testExecDir(t, tmpDir, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{
			"cargo", "run", "--example", "tlsclient", "--",
			"--cafile", caPath,
			"-p", fmt.Sprintf("%d", tlsPort),
			"--http", fmt.Sprintf("%s.localhost", testCaseName),
		}
	})
}
