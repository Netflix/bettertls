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
	"botan":           &BotanRunner{},
	"curl":            &CurlRunner{},
	"envoy":           &EnvoyRunner{},
	"gnutls":          &GnutlsRunner{},
	"golang":          &GolangRunner{},
	"java":            &JavaRunner{},
	"libressl":        &LibresslRunner{},
	"node":            &NodeRunner{},
	"openssl":         &OpensslRunner{},
	"pkijs":           &PkijsRunner{},
	"powershell":      &PowerShellRunner{},
	"python_requests": &PythonRequestsRunner{},
	"rustls":          &RustlsRunner{},
}
