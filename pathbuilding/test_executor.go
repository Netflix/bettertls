package pathbuilding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func ExecuteTests(t *testing.T, provider *TestCaseProvider, execTest func(testCaseHostname string) (bool, error)) error {
	execTestCase := func(testCase *ExplicitTestCase) bool {
		hostname, err := EncodeHostname(testCase)
		if err != nil {
			t.Fatal(err)
		}
		result, err := execTest(hostname)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	passesSanityCheck := execTestCase(&ExplicitTestCase{
		TrustGraph: LINEAR_TRUST_GRAPH,
		SrcNode:    "Trust Anchor",
		DstNode:    "EE",
	})
	if !passesSanityCheck {
		t.FailNow()
	}

	var supportsBranchingGraphs = true
	for _, rootName := range []string{"Root1", "Root2"} {
		if !execTestCase(&ExplicitTestCase{
			TrustGraph: TWO_ROOTS,
			SrcNode:    rootName,
			DstNode:    "EE",
		}) {
			t.Log("Based on testing the implementation, branching path discovery is not supported.")
			supportsBranchingGraphs = false
			break
		}
	}

	invalidReasonSupported := make([]bool, len(InvalidReasons()))
	for _, reason := range InvalidReasons() {
		if reason == INVALID_REASON_UNSPECIFIED {
			continue
		}
		hasValidPath := execTestCase(&ExplicitTestCase{
			TrustGraph:    LINEAR_TRUST_GRAPH,
			SrcNode:       "Trust Anchor",
			DstNode:       "EE",
			InvalidEdges:  []Edge{{"Trust Anchor", "ICA"}},
			InvalidReason: reason,
			ExpectFailure: true,
		})
		if hasValidPath {
			t.Logf("Based on testing the implementation, invalid certificate reason %s is not supported.", reason)
			invalidReasonSupported[reason] = false
		} else {
			invalidReasonSupported[reason] = true
		}
	}

	testCaseList, err := provider.GetTestCases()
	if err != nil {
		return err
	}

	for _, testCase := range testCaseList.TestCases {
		testCaseManifest, err := provider.GetTestCase(testCase)
		if err != nil {
			return err
		}
		t.Run(testCaseManifest.TrustGraph+"/"+testCaseManifest.Name, func(t *testing.T) {
			if !supportsBranchingGraphs && testCaseManifest.TrustGraph != LINEAR_TRUST_GRAPH.Name() {
				t.Skipf("Branching trust graphs not supported: %s", testCaseManifest.TrustGraph)
			}
			if len(testCaseManifest.InvalidEdges) > 0 && !invalidReasonSupported[InvalidReasonFromString(testCaseManifest.InvalidReason)] {
				t.Skipf("Invalid reason not supported: %s", testCaseManifest.InvalidReason)
			}

			chainValidate, err := execTest(testCaseManifest.Name)
			if err != nil {
				t.Fatal(err)
			}
			if len(testCaseManifest.ExpectedPath) > 0 {
				assert.True(t, chainValidate, "should have been able to find a valid chain")
			} else {
				assert.False(t, chainValidate, "should not have found a valid chain")
			}
		})
	}

	return nil
}
