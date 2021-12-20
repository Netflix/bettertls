package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"net"
)

type OpensslRunner struct {
	version string
}

func (o *OpensslRunner) Name() string {
	return "openssl"
}

func (o *OpensslRunner) Initialize() error {
	var err error
	o.version, err = execAndCapture("openssl", "version")
	if err != nil {
		return err
	}
	return nil
}

func (o *OpensslRunner) Close() error {
	return nil
}

func (o *OpensslRunner) GetVersion() string {
	return o.version
}

func (o *OpensslRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		args := []string{"openssl", "s_client",
			"-CAfile", caPath,
			"-connect", fmt.Sprintf("%s:%d", hostname, tlsPort),
			"-verify_return_error"}

		ipAddr := net.ParseIP(hostname)
		if ipAddr == nil {
			args = append(args, "-verify_hostname", hostname)
		} else {
			args = append(args, "-verify_ip", hostname)
		}

		return args
	})
}
