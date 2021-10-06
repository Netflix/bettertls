package impltests

import (
	"fmt"
	"testing"
)

func TestOpenSSL(t *testing.T) {
	version, err := execAndCapture("openssl", "version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{"openssl", "s_client",
			"-CAfile", caPath,
			"-connect", fmt.Sprintf("localhost:%d", tlsPort),
			"-servername", testCaseName + ".localhost",
			"-verify_return_error",
		}
	})
}
