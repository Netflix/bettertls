package test_case

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
)

type ExpectedResult int

const (
	EXPECTED_RESULT_FAIL = iota
	EXPECTED_RESULT_PASS
	EXPECTED_RESULT_SOFT_FAIL
	EXPECTED_RESULT_SOFT_PASS
)

func (r ExpectedResult) String() string {
	switch r {
	case EXPECTED_RESULT_FAIL:
		return "FAIL"
	case EXPECTED_RESULT_PASS:
		return "PASS"
	case EXPECTED_RESULT_SOFT_FAIL:
		return "SOFT_FAIL"
	case EXPECTED_RESULT_SOFT_PASS:
		return "SOFT_PASS"
	}
	return fmt.Sprintf("%v", r)
}
func (r ExpectedResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}
func (r *ExpectedResult) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	case "FAIL":
		*r = EXPECTED_RESULT_FAIL
	case "PASS":
		*r = EXPECTED_RESULT_PASS
	case "SOFT_FAIL":
		*r = EXPECTED_RESULT_SOFT_FAIL
	case "SOFT_PASS":
		*r = EXPECTED_RESULT_SOFT_PASS
	default:
		return fmt.Errorf("invalid expected result: %s", s)
	}
	return nil
}

type Feature int

type TestCase interface {
	// The expected result (whether the client should reject or accept the TLS connection)
	ExpectedResult() ExpectedResult
	// What hostname should be used in the request, e.g. "localhost" or "127.0.0.1".
	GetHostname() string
	// A callback to get the server certificates for this test case
	GetCertificates(rootCert *x509.Certificate, rootKey crypto.Signer) (*tls.Certificate, error)
	// Which supported client features are required in order to meaningfully run this test
	RequiredFeatures() []Feature
}

type TestCaseProvider interface {
	Name() string
	// How many test cases does this provider supply?
	GetTestCaseCount() (uint, error)
	// Get the "index"th test case.
	GetTestCase(index uint) (TestCase, error)
	// Get a test case index that verifies the client can trusts certificates under the server's root
	GetSanityCheckTestCase() (uint, error)
	// Get a list of features used by this test provider
	GetFeatures() []Feature
	// Get a string description of a feature
	DescribeFeature(feature Feature) string
	// For a given feature, a list of test cases (as indices) that must pass for the client to be considered to support the feature.
	GetTestCasesForFeature(feature Feature) ([]uint, error)
}
