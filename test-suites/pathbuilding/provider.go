package pathbuilding

import (
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
)

const (
	SANITY_CHECK_TEST_CASE uint = iota
	BRANCHING_FEATURE_TEST_CASE_1
	BRANCHING_FEATURE_TEST_CASE_2
	FIRST_INVALID_REASON_TEST_CASE
)

type TestCaseProvider struct {
	testCases []test_case.TestCase
}

func NewTestCaseProvider() *TestCaseProvider {
	testCases := make([]test_case.TestCase, 3)

	testCases[SANITY_CHECK_TEST_CASE] = &TestCaseImpl{
		ExplicitTestCase: &ExplicitTestCase{
			TrustGraph: LINEAR_TRUST_GRAPH,
			SrcNode:    "Trust Anchor",
			DstNode:    "EE",
		},
	}

	testCases[BRANCHING_FEATURE_TEST_CASE_1] = &TestCaseImpl{
		ExplicitTestCase: &ExplicitTestCase{
			TrustGraph: TWO_ROOTS,
			SrcNode:    "Root1",
			DstNode:    "EE",
		},
	}
	testCases[BRANCHING_FEATURE_TEST_CASE_2] = &TestCaseImpl{
		ExplicitTestCase: &ExplicitTestCase{
			TrustGraph: TWO_ROOTS,
			SrcNode:    "Root2",
			DstNode:    "EE",
		},
	}

	for _, reason := range InvalidReasons() {
		if reason == INVALID_REASON_UNSPECIFIED {
			continue
		}
		testCases = append(testCases, &TestCaseImpl{
			ExplicitTestCase: &ExplicitTestCase{
				TrustGraph:    LINEAR_TRUST_GRAPH,
				SrcNode:       "Trust Anchor",
				DstNode:       "EE",
				InvalidEdges:  []Edge{{"Trust Anchor", "ICA"}},
				InvalidReason: reason,
				ExpectFailure: true,
			},
			InvalidReason: reason,
		})
	}

	for _, testCase := range EXPLICIT_TEST_CASES {
		if len(testCase.InvalidEdges) > 0 && testCase.InvalidReason == INVALID_REASON_UNSPECIFIED {
			for _, reason := range InvalidReasons() {
				if reason == INVALID_REASON_UNSPECIFIED {
					continue
				}
				testCases = append(testCases, &TestCaseImpl{
					ExplicitTestCase: testCase,
					InvalidReason:    reason,
				})
			}
		} else {
			testCases = append(testCases, &TestCaseImpl{
				ExplicitTestCase: testCase,
				InvalidReason:    testCase.InvalidReason,
			})
		}
	}

	return &TestCaseProvider{
		testCases: testCases,
	}
}

func (p *TestCaseProvider) Name() string {
	return "pathbuilding"
}

func (p *TestCaseProvider) GetTestCaseCount() (uint, error) {
	return uint(len(p.testCases)), nil
}

func (p *TestCaseProvider) GetTestCase(testCaseIdx uint) (test_case.TestCase, error) {
	return p.testCases[testCaseIdx], nil
}

func (p *TestCaseProvider) GetSanityCheckTestCase() (uint, error) {
	return SANITY_CHECK_TEST_CASE, nil
}

func (p *TestCaseProvider) GetFeatures() []test_case.Feature {
	features := make([]test_case.Feature, 0, 1+len(InvalidReasons()))
	features = append(features, FEATURE_BRANCHING)
	for _, invalidReason := range InvalidReasons() {
		if invalidReason == INVALID_REASON_UNSPECIFIED {
			continue
		}
		features = append(features, invalidReasonToFeature(invalidReason))
	}
	return features
}

func (p *TestCaseProvider) DescribeFeature(feature test_case.Feature) string {
	if feature == FEATURE_BRANCHING {
		return "BRANCHING"
	}
	for _, reason := range InvalidReasons() {
		if reason == INVALID_REASON_UNSPECIFIED {
			continue
		}
		if feature == invalidReasonToFeature(reason) {
			return "INVALID_REASON_" + reason.String()
		}
	}
	panic(fmt.Errorf("unsupported feature: %d", feature))
}

func (p *TestCaseProvider) GetTestCasesForFeature(feature test_case.Feature) ([]uint, error) {
	if feature == FEATURE_BRANCHING {
		return []uint{BRANCHING_FEATURE_TEST_CASE_1, BRANCHING_FEATURE_TEST_CASE_2}, nil
	}
	for idx, reason := range InvalidReasons() {
		if feature == invalidReasonToFeature(reason) {
			return []uint{FIRST_INVALID_REASON_TEST_CASE + uint(idx-1)}, nil
		}
	}
	return nil, fmt.Errorf("invalid feature: %v", feature)
}
