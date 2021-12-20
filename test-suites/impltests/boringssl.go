package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
)

type BoringSslRunner struct{}

func (b *BoringSslRunner) Name() string {
	return "boringssl"
}

func (b *BoringSslRunner) Initialize() error {
	return nil
}

func (b *BoringSslRunner) Close() error {
	return nil
}

func (b *BoringSslRunner) GetVersion() string {
	return "boringssl revision a406ad76ad31c07b094ff60300146724a1448251"
}

func (b *BoringSslRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{"bssl", "s_client",
			"-root-certs", caPath,
			"-connect", fmt.Sprintf("%s:%d", hostname, tlsPort),
		}
	})
}
