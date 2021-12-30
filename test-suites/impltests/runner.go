package impltests

import test_executor "github.com/Netflix/bettertls/test-suites/test-executor"

type ImplementationRunner interface {
	Name() string
	Initialize() error
	Close() error
	GetVersion() string
	RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error)
}

var Runners = map[string]ImplementationRunner{
	"boringssl":       &BoringSslRunner{},
	"curl":            &CurlRunner{},
	"gnutls":          &GnutlsRunner{},
	"golang":          &GolangRunner{},
	"java":            &JavaRunner{},
	"libressl":        &LibresslRunner{},
	"openssl":         &OpensslRunner{},
	"pkijs":           &PkijsRunner{},
	"python_requests": &PythonRequestsRunner{},
	"rustls":          &RustlsRunner{},
}
