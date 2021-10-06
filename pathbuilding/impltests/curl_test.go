package impltests

import (
	"fmt"
	"testing"
)

func TestCurl(t *testing.T) {
	version, err := execAndCapture("curl", "--version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{
			"curl", "-s", "-v", "--cacert", caPath,
			"--resolve", fmt.Sprintf("%s.localhost:%d:127.0.0.1", testCaseName, tlsPort),
			fmt.Sprintf("https://%s.localhost:%d/ok", testCaseName, tlsPort),
		}
	})
}
