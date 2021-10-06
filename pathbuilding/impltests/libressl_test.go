package impltests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibreSSL(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	libreSslPath := filepath.Join(homeDir, "src", "libressl-3.4.0", "apps", "openssl", "openssl")
	version, err := execAndCapture(libreSslPath, "version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{"bash", "-c", strings.Join([]string{
			libreSslPath, "s_client",
			"-CAfile", caPath,
			"-connect", fmt.Sprintf("localhost:%d", tlsPort),
			"-servername", testCaseName + ".localhost",
			"-verify_return_error",
			"|", "grep", "\"Verify return code: 0\"",
		}, " "),
		}
	})
}
