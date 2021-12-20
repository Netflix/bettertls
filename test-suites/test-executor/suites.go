package test_executor

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"github.com/Netflix/bettertls/test-suites/certutil"
	"github.com/Netflix/bettertls/test-suites/nameconstraints"
	"github.com/Netflix/bettertls/test-suites/pathbuilding"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
)

type TestSuites struct {
	rootCert  *x509.Certificate
	rootKey   crypto.Signer
	providers []test_case.TestCaseProvider
}

func (ts *TestSuites) GetRootCert() *x509.Certificate {
	return ts.rootCert
}

func (ts *TestSuites) GetProviderNames() []string {
	names := make([]string, 0, len(ts.providers))
	for _, provider := range ts.providers {
		names = append(names, provider.Name())
	}
	return names
}

func (ts *TestSuites) GetProvider(name string) test_case.TestCaseProvider {
	for _, provider := range ts.providers {
		if provider.Name() == name {
			return provider
		}
	}
	return nil
}

func (ts *TestSuites) GetTestCaseCertificates(testCase test_case.TestCase) (*tls.Certificate, error) {
	return testCase.GetCertificates(ts.rootCert, ts.rootKey)
}

func BuildTestSuites() (*TestSuites, error) {
	rootCa, rootKey, err := certutil.GenerateSelfSignedCert("bettertls_trust_root")
	if err != nil {
		return nil, err
	}
	return BuildTestSuitesWithRootCa(rootCa, rootKey)
}

func BuildTestSuitesWithRootCa(rootCert *x509.Certificate, rootKey crypto.Signer) (*TestSuites, error) {
	return &TestSuites{
		rootCert: rootCert,
		rootKey:  rootKey,
		providers: []test_case.TestCaseProvider{
			nameconstraints.NewTestCaseProvider(),
			pathbuilding.NewTestCaseProvider(),
		},
	}, nil
}
