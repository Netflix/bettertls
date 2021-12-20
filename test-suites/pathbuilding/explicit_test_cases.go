package pathbuilding

type InvalidReason int

const (
	INVALID_REASON_UNSPECIFIED = iota
	INVALID_REASON_EXPIRED
	INVALID_REASON_NAME_CONSTRAINTS
	INVALID_REASON_BAD_EKU
	INVALID_REASON_MISSING_BASIC_CONSTRAINTS
	INVALID_REASON_NOT_A_CA
	INVALID_REASON_DEPRECATED_CRYPTO
)

func (r InvalidReason) String() string {
	return []string{"UNSPECIFIED", "EXPIRED", "NAME_CONSTRAINTS", "BAD_EKU", "MISSING_BASIC_CONSTRAINTS", "NOT_A_CA", "DEPRECATED_CRYPTO"}[r]
}
func InvalidReasons() []InvalidReason {
	return []InvalidReason{INVALID_REASON_UNSPECIFIED, INVALID_REASON_EXPIRED, INVALID_REASON_NAME_CONSTRAINTS, INVALID_REASON_BAD_EKU,
		INVALID_REASON_MISSING_BASIC_CONSTRAINTS, INVALID_REASON_NOT_A_CA, INVALID_REASON_DEPRECATED_CRYPTO}
}
func InvalidReasonFromString(s string) InvalidReason {
	for _, reason := range InvalidReasons() {
		if s == reason.String() {
			return reason
		}
	}
	return INVALID_REASON_UNSPECIFIED
}

type ExplicitTestCase struct {
	// A trust graph, as defined in trust_graph.go
	TrustGraph *TrustGraph
	// The source (trust anchor) the client will trust in this test case
	SrcNode string
	// The destination entity which will create the leaf certificate the client needs to verify
	DstNode string
	// Edges from the trust graph that will be made invalid (e.g. expired) in the test case
	InvalidEdges []Edge
	// The reason that the above-listed edges will be invalid, such as expired or bad EKU constraints. If unspecified
	// and there are non-zero invalid edges, this test case is a "meta" test case that will be expanded to include
	// all supported invalid reasons
	InvalidReason InvalidReason
	// Whether it is expected that the client will fail to be able to build a trust path
	ExpectFailure bool
	// An optional comment, explaining what the test is and/or why a client be able to succeed/fail at building the
	// trust path.
	Comment string
}

var EXPLICIT_TEST_CASES = []*ExplicitTestCase{
	{
		TrustGraph: LINEAR_TRUST_GRAPH,
		SrcNode:    "Trust Anchor",
		DstNode:    "EE",
		Comment:    "The most basic linear chain.",
	},
	{
		TrustGraph:    LINEAR_TRUST_GRAPH,
		SrcNode:       "Trust Anchor",
		DstNode:       "EE",
		InvalidEdges:  []Edge{{"Trust Anchor", "ICA"}},
		ExpectFailure: true,
		Comment:       "The most basic linear chain, broken by an expired ICA.",
	},
	{
		TrustGraph:   LINEAR_TRUST_GRAPH,
		SrcNode:      "ICA",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"Trust Anchor", "ICA"}},
		Comment:      "ICA is trusted directly, so the expired TA => ICA cert should not cause validation to fail.",
	},
	{
		TrustGraph: TWO_ROOTS,
		SrcNode:    "Root1",
		DstNode:    "EE",
		Comment:    "Should be able to discover a path to either root.",
	},
	{
		TrustGraph: TWO_ROOTS,
		SrcNode:    "Root2",
		DstNode:    "EE",
		Comment:    "Should be able to discover a path to either root.",
	},
	{
		TrustGraph:   TWO_ROOTS,
		SrcNode:      "Root1",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"Root2", "ICA"}},
		Comment:      "Should be able to discover a path when alternate root is invalid.",
	},
	{
		TrustGraph:   TWO_ROOTS,
		SrcNode:      "Root2",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"Root1", "ICA"}},
		Comment:      "Should be able to discover a path when alternate root is invalid.",
	},
	{
		TrustGraph:    TWO_ROOTS,
		SrcNode:       "Root1",
		DstNode:       "EE",
		InvalidEdges:  []Edge{{"Root1", "ICA"}},
		ExpectFailure: true,
		Comment:       "Should not be able to find a path when only trusted root is invalid.",
	},
	{
		TrustGraph:    TWO_ROOTS,
		SrcNode:       "Root2",
		DstNode:       "EE",
		InvalidEdges:  []Edge{{"Root2", "ICA"}},
		ExpectFailure: true,
		Comment:       "Should not be able to find a path when only trusted root is invalid.",
	},
	{
		TrustGraph: FIGURE_SEVEN,
		SrcNode:    "Trust Anchor",
		DstNode:    "EE",
		Comment:    "Should be able to find a path through a more complicated tree.",
	},
	{
		TrustGraph:   FIGURE_SEVEN,
		SrcNode:      "Trust Anchor",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"Trust Anchor", "C"}, {"A", "B"}},
		Comment:      "Should be able to find an alternate path through a more complicated tree.",
	},
	{
		TrustGraph:   FIGURE_SEVEN,
		SrcNode:      "Trust Anchor",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"Trust Anchor", "A"}, {"C", "B"}},
		Comment:      "Should be able to find an alternate path through a more complicated tree.",
	},
	{
		TrustGraph: BRIDGE_CA_PKI,
		SrcNode:    "TA Z",
		DstNode:    "D",
		Comment:    "RFC 4158, section 2.3",
	},
	{
		TrustGraph: BRIDGE_CA_PKI,
		SrcNode:    "TA Z",
		DstNode:    "EE",
		Comment:    "RFC 4158, section 2.4.2",
	},
	{
		TrustGraph:   BRIDGE_CA_PKI,
		SrcNode:      "TA Z",
		DstNode:      "EE",
		InvalidEdges: []Edge{{"TA X", "Bridge CA"}},
		Comment:      "Irrelevant expired cross-signed cert",
	},
	{
		TrustGraph:    BRIDGE_CA_PKI,
		SrcNode:       "TA Z",
		DstNode:       "EE",
		InvalidEdges:  []Edge{{"Bridge CA", "TA X"}},
		ExpectFailure: true,
		Comment:       "Certificate from bridge CA into infrastructure X is invalid.",
	},
	{
		TrustGraph:    BRIDGE_CA_PKI,
		SrcNode:       "TA Z",
		DstNode:       "EE",
		InvalidEdges:  []Edge{{"TA Z", "Bridge CA"}},
		ExpectFailure: true,
		Comment:       "Certificate from infrastructure Z to bridge CA is invalid.",
	},
}
