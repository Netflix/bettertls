package impltests

import (
	"fmt"
	"testing"
)

func TestBoringSSL(t *testing.T) {
	t.Log("boringssl revision a406ad76ad31c07b094ff60300146724a1448251")

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{"bssl", "s_client",
			"-root-certs", caPath,
			"-connect", fmt.Sprintf("localhost:%d", tlsPort),
			"-server-name", testCaseName + ".localhost",
		}
	})
}
