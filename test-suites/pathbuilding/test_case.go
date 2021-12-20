package pathbuilding

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
)

const FEATURE_BRANCHING = 0

func invalidReasonToFeature(reason InvalidReason) test_case.Feature {
	return test_case.Feature(1 + reason)
}

type TestCaseImpl struct {
	ExplicitTestCase *ExplicitTestCase
	InvalidReason    InvalidReason
}

func (p *TestCaseImpl) GetHostname() string {
	return "localhost"
}

func (p *TestCaseImpl) ExpectedResult() test_case.ExpectedResult {
	etc := p.ExplicitTestCase
	path := etc.TrustGraph.Reachable(etc.InvalidEdges, etc.SrcNode, etc.DstNode)
	if len(path) > 0 {
		return test_case.EXPECTED_RESULT_PASS
	}
	return test_case.EXPECTED_RESULT_FAIL
}

func (p *TestCaseImpl) RequiredFeatures() []test_case.Feature {
	requiredFeatures := make([]test_case.Feature, 0, 2)
	if p.ExplicitTestCase.TrustGraph != LINEAR_TRUST_GRAPH {
		requiredFeatures = append(requiredFeatures, FEATURE_BRANCHING)
	}
	if len(p.ExplicitTestCase.InvalidEdges) > 0 && p.InvalidReason != INVALID_REASON_UNSPECIFIED {
		requiredFeatures = append(requiredFeatures, invalidReasonToFeature(p.InvalidReason))
	}
	return requiredFeatures
}

func (p *TestCaseImpl) GetCertificates(rootCert *x509.Certificate, rootKey crypto.Signer) (*tls.Certificate, error) {
	return GenerateCerts(rootCert, rootKey, "localhost", p)
}
