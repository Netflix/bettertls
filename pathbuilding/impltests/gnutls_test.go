package impltests

import (
	"fmt"
	"testing"
)

func TestGnuTls(t *testing.T) {
	version, err := execAndCapture("gnutls-cli", "--version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{
			"gnutls-cli", "--x509cafile", caPath,
			"--sni-hostname", testCaseName + ".localhost",
			"--verify-hostname", testCaseName + ".localhost",
			fmt.Sprintf("localhost:%d", tlsPort),
		}
	})
}
