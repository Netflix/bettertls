package impltests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Netflix/bettertls/pathbuilding"
	"net/http"
	"runtime"
	"testing"
)

func x509Verify(trustAnchor *x509.Certificate, certs [][]byte) (bool, error) {
	leafCert, err := x509.ParseCertificate(certs[0])
	if err != nil {
		return false, err
	}
	intermediates := x509.NewCertPool()
	for i := 1; i < len(certs); i++ {
		cert, err := x509.ParseCertificate(certs[i])
		if err != nil {
			return false, err
		}
		intermediates.AddCert(cert)
	}
	roots := x509.NewCertPool()
	roots.AddCert(trustAnchor)

	_, err = leafCert.Verify(x509.VerifyOptions{
		DNSName:       "localhost",
		Intermediates: intermediates,
		Roots:         roots,
	})
	return err == nil, nil
}

func TestX509Verify(t *testing.T) {
	t.Log("Go version: " + runtime.Version())

	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		t.Fatal(err)
	}
	rootCert := provider.GetRootCert()

	err = pathbuilding.ExecuteTests(t, provider, func(testCaseName string) (bool, error) {
		testCaseManifest, err := provider.GetTestCase(testCaseName)
		if err != nil {
			return false, err
		}
		return x509Verify(rootCert, testCaseManifest.Certificates)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHttpClient(t *testing.T) {
	t.Log("Go version: " + runtime.Version())

	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		t.Fatal(err)
	}
	server, err := pathbuilding.StartServer(provider, noplog.Logger, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	truststore := x509.NewCertPool()
	truststore.AddCert(provider.GetRootCert())

	err = pathbuilding.ExecuteTests(t, provider, func(testCaseName string) (bool, error) {
		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    truststore,
					ServerName: testCaseName + ".localhost",
				},
			},
		}
		_, err := client.Get(fmt.Sprintf("https://localhost:%d/ok", server.TlsPort()))
		return err == nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
