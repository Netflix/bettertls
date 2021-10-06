package pathbuilding

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
)

type Manifest struct {
	Root        []byte                `json:"root"`
	TrustGraphs []*TrustGraphManifest `json:"trustGraphs"`
}

type TrustGraphManifest struct {
	Name  string     `json:"name"`
	Nodes []string   `json:"nodes"`
	Edges [][]string `json:"edges"`
}

type TestCaseList struct {
	TestCases []string `json:"testCases"`
}

type TestCaseManifest struct {
	Name          string     `json:"name"`
	TrustGraph    string     `json:"trustGraph"`
	SrcNode       string     `json:"srcNode"`
	DstNode       string     `json:"dstNode"`
	InvalidEdges  [][]string `json:"invalidEdges"`
	InvalidReason string     `json:"invalidReason"`
	ExpectedPath  []string   `json:"expectedPath"`
	Certificates  [][]byte   `json:"certificates"`
}

type TestCaseProvider struct {
	rootCert *x509.Certificate
	rootKey  crypto.Signer
}

func NewTestCaseProvider() (*TestCaseProvider, error) {
	rootCa, rootKey, err := generateSelfSignedCert("bettertls_trust_root")
	if err != nil {
		return nil, err
	}
	return &TestCaseProvider{rootCa, rootKey}, nil
}

func (p *TestCaseProvider) GetRootCert() *x509.Certificate {
	return p.rootCert
}

func (p *TestCaseProvider) GetManifest() (*Manifest, error) {
	var trustGraphs []*TrustGraphManifest
	for _, trustGraph := range ALL_TRUST_GRAPHS {
		var edges [][]string
		for _, edge := range trustGraph.GetAllEdges() {
			edges = append(edges, []string{edge.Source, edge.Destination})
		}
		trustGraphs = append(trustGraphs, &TrustGraphManifest{
			Name:  trustGraph.Name(),
			Nodes: trustGraph.NodeNames(),
			Edges: edges,
		})
	}

	return &Manifest{
		Root:        p.rootCert.Raw,
		TrustGraphs: trustGraphs,
	}, nil
}

func (p *TestCaseProvider) GetTestCases() (*TestCaseList, error) {
	var testCases []string

	// Start with the explicit tests
	for _, testCase := range EXPLICIT_TEST_CASES {
		if len(testCase.InvalidEdges) > 0 && testCase.InvalidReason == INVALID_REASON_UNSPECIFIED {
			for _, reason := range InvalidReasons() {
				if reason == INVALID_REASON_UNSPECIFIED {
					continue
				}
				tc := &ExplicitTestCase{
					TrustGraph:    testCase.TrustGraph,
					SrcNode:       testCase.SrcNode,
					DstNode:       testCase.DstNode,
					InvalidEdges:  testCase.InvalidEdges,
					InvalidReason: reason,
					ExpectFailure: testCase.ExpectFailure,
					Comment:       testCase.Comment,
				}
				hostname, err := EncodeHostname(tc)
				if err != nil {
					return nil, err
				}
				testCases = append(testCases, hostname)
			}
		} else {
			hostname, err := EncodeHostname(testCase)
			if err != nil {
				return nil, err
			}
			testCases = append(testCases, hostname)
		}
	}

	return &TestCaseList{
		TestCases: testCases,
	}, nil
}

func (p *TestCaseProvider) GetTestCase(testCaseName string) (*TestCaseManifest, error) {
	testCase, err := DecodeHostname(testCaseName)
	if err != nil {
		return nil, err
	}

	invalidEdges := make([][]string, 0, len(testCase.InvalidEdges))
	for _, edge := range testCase.InvalidEdges {
		invalidEdges = append(invalidEdges, []string{edge.Source, edge.Destination})
	}

	expectedPath := testCase.TrustGraph.Reachable(testCase.InvalidEdges, testCase.SrcNode, testCase.DstNode)
	certificates, err := p.GetCertificatesForTestCase(testCase, "localhost")
	if err != nil {
		return nil, err
	}

	return &TestCaseManifest{
		Name:          testCaseName,
		TrustGraph:    testCase.TrustGraph.Name(),
		SrcNode:       testCase.SrcNode,
		DstNode:       testCase.DstNode,
		InvalidEdges:  invalidEdges,
		InvalidReason: testCase.InvalidReason.String(),
		ExpectedPath:  expectedPath,
		Certificates:  certificates.Certificate,
	}, nil
}

func (p *TestCaseProvider) GetCertificatesForTestCase(testCase *ExplicitTestCase, hostname string) (*tls.Certificate, error) {
	return GenerateCerts(p.rootCert, p.rootKey, hostname, testCase)
}
