package impltests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"net/http"
	"runtime"
)

type GolangRunner struct{}

func (g *GolangRunner) Name() string {
	return "go"
}

func (g *GolangRunner) Initialize() error {
	return nil
}

func (g *GolangRunner) Close() error {
	return nil
}

func (g *GolangRunner) GetVersion() string {
	return runtime.Version()
}

func (g *GolangRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return nil, err
	}

	truststore := x509.NewCertPool()
	truststore.AddCert(suites.GetRootCert())
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: truststore,
			},
		},
	}

	return test_executor.ExecuteAllTestsRemote(ctx, suites, func(hostname string, port uint) (bool, error) {
		resp, err := client.Get(fmt.Sprintf("https://%s:%d/ok", hostname, port))
		if err != nil {
			return false, nil
		}
		resp.Body.Close()
		return true, nil
	})
}
