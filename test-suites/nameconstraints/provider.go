package nameconstraints

import (
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	"net"
)

type TestCaseProvider struct {
	testCases []NameConstraintsTestCase
}

const (
	SANITY_CHECK_TEST_CASE uint = iota
	FEATURE_NAME_CONSTRAINTS_TEST_CASE
	FEATURE_VALIDATE_DNS_TEST_CASE_1
	FEATURE_VALIDATE_DNS_TEST_CASE_2
	FEATURE_VALIDATE_IP_TEST_CASE_1
	FEATURE_VALIDATE_IP_TEST_CASE_2
)

func NewTestCaseProvider() *TestCaseProvider {
	testCases := make([]NameConstraintsTestCase, 6, 9491)

	testCases[SANITY_CHECK_TEST_CASE] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
		CommonNameType:              CN_TYPE_DNS,
		CommonNameValue:             EXTVAL_VALID,
		DnsSan:                      EXTVAL_VALID,
		IpSan:                       EXTVAL_NONE,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_NONE,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases[FEATURE_NAME_CONSTRAINTS_TEST_CASE] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
		CommonNameType:              CN_TYPE_DNS,
		CommonNameValue:             EXTVAL_VALID,
		DnsSan:                      EXTVAL_VALID,
		IpSan:                       EXTVAL_NONE,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_VALID,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases[FEATURE_VALIDATE_DNS_TEST_CASE_1] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
		CommonNameType:              CN_TYPE_DNS,
		CommonNameValue:             EXTVAL_NONE,
		DnsSan:                      EXTVAL_VALID,
		IpSan:                       EXTVAL_NONE,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_NONE,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases[FEATURE_VALIDATE_DNS_TEST_CASE_2] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
		CommonNameType:              CN_TYPE_DNS,
		CommonNameValue:             EXTVAL_NONE,
		DnsSan:                      EXTVAL_INVALID,
		IpSan:                       EXTVAL_NONE,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_NONE,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases[FEATURE_VALIDATE_IP_TEST_CASE_1] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_IP,
		CommonNameType:              CN_TYPE_IP,
		CommonNameValue:             EXTVAL_NONE,
		DnsSan:                      EXTVAL_NONE,
		IpSan:                       EXTVAL_VALID,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_NONE,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases[FEATURE_VALIDATE_IP_TEST_CASE_2] = NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_IP,
		CommonNameType:              CN_TYPE_IP,
		CommonNameValue:             EXTVAL_NONE,
		DnsSan:                      EXTVAL_NONE,
		IpSan:                       EXTVAL_INVALID,
		NameConstraintsIpWhitelist:  EXTVAL_NONE,
		NameConstraintsDnsWhitelist: EXTVAL_NONE,
		NameConstraintsIpBlacklist:  EXTVAL_NONE,
		NameConstraintsDnsBlacklist: EXTVAL_NONE,
	}

	testCases = append(testCases, NameConstraintsTestCase{
		ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
		DnsSan:                      EXTVAL_INVALID,
		NameConstraintsDnsBlacklist: EXTVAL_VALID,
		ExtraSan:                    &ExtraSan{tag: 6, value: []byte("http://foo.bar, DNS:" + VALID_DNS_NAME)},
	})

	for _, clientHostnameType := range []ClientHostnameType{CLIENT_HOSTNAME_TYPE_DNS, CLIENT_HOSTNAME_TYPE_IP} {
		for _, commonNameType := range []CommonNameType{CN_TYPE_DNS, CN_TYPE_IP} {
			for _, commonNameValue := range ALL_TRINARY_VALUES {
				for _, dnsSan := range ALL_TRINARY_VALUES {
					for _, ipSan := range ALL_TRINARY_VALUES {
						for _, dnsWhitelist := range ALL_TRINARY_VALUES {
							for _, ipWhitelist := range ALL_TRINARY_VALUES {
								for _, dnsBlacklist := range ALL_TRINARY_VALUES {
									for _, ipBlacklist := range ALL_TRINARY_VALUES {
										tc := NameConstraintsTestCase{
											ClientHostnameType:          clientHostnameType,
											CommonNameType:              commonNameType,
											CommonNameValue:             commonNameValue,
											DnsSan:                      dnsSan,
											IpSan:                       ipSan,
											NameConstraintsIpWhitelist:  ipWhitelist,
											NameConstraintsDnsWhitelist: dnsWhitelist,
											NameConstraintsIpBlacklist:  ipBlacklist,
											NameConstraintsDnsBlacklist: dnsBlacklist,
										}
										testCases = append(testCases, tc)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Try encoding the client-requested hostname/IP into SANs with other tags
	for sanTag := 0; sanTag < 16; sanTag++ {
		// Encode the DNS hostname into SANs with a violating NC (both whitelist and blacklist). Try encoding it with
		// the raw DNS name and also with a few URI schemes with the valid DNS name in the hostname component
		for _, hostnameEncoding := range []string{VALID_DNS_NAME,
			"http://" + VALID_DNS_NAME, "https://" + VALID_DNS_NAME, "spiffe://" + VALID_DNS_NAME, "unspec://" + VALID_DNS_NAME,
		} {
			testCases = append(testCases, NameConstraintsTestCase{
				ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
				DnsSan:                      EXTVAL_INVALID,
				NameConstraintsDnsWhitelist: EXTVAL_INVALID,
				ExtraSan:                    &ExtraSan{tag: sanTag, value: []byte(hostnameEncoding)},
			})
			testCases = append(testCases, NameConstraintsTestCase{
				ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
				DnsSan:                      EXTVAL_INVALID,
				NameConstraintsDnsBlacklist: EXTVAL_VALID,
				ExtraSan:                    &ExtraSan{tag: sanTag, value: []byte(hostnameEncoding)},
			})
		}

		validIp := net.ParseIP(VALID_IP)
		if ip := validIp.To4(); ip != nil {
			validIp = ip
		}

		// Add tests that attempt to encode the IP as both its raw bytes and stringified, also for both whilelist and blacklist NCs.
		for _, ipEncoding := range [][]byte{validIp, []byte(VALID_IP)} {
			testCases = append(testCases, NameConstraintsTestCase{
				ClientHostnameType:         CLIENT_HOSTNAME_TYPE_IP,
				IpSan:                      EXTVAL_INVALID,
				NameConstraintsIpWhitelist: EXTVAL_INVALID,
				ExtraSan:                   &ExtraSan{tag: sanTag, value: ipEncoding},
			})
			testCases = append(testCases, NameConstraintsTestCase{
				ClientHostnameType:         CLIENT_HOSTNAME_TYPE_IP,
				IpSan:                      EXTVAL_INVALID,
				NameConstraintsIpBlacklist: EXTVAL_VALID,
				ExtraSan:                   &ExtraSan{tag: sanTag, value: ipEncoding},
			})
		}
	}

	// Attempt to smuggle additional SANs using separator strings. For example, old versions of Node would split based on ", ".
	for _, separator := range []string{", ", ",", ";", ":"} {
		for _, prefix := range []string{"", "DNS:"} {
			for _, leadingSan := range []string{INVALID_DNS_NAME, "http://" + INVALID_DNS_NAME} {
				sanValue := []byte(leadingSan + separator + prefix + VALID_DNS_NAME)
				for sanTag := 0; sanTag < 16; sanTag++ {
					testCases = append(testCases, NameConstraintsTestCase{
						ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
						DnsSan:                      EXTVAL_INVALID,
						NameConstraintsDnsWhitelist: EXTVAL_INVALID,
						ExtraSan:                    &ExtraSan{tag: sanTag, value: sanValue},
					})
					testCases = append(testCases, NameConstraintsTestCase{
						ClientHostnameType:          CLIENT_HOSTNAME_TYPE_DNS,
						DnsSan:                      EXTVAL_INVALID,
						NameConstraintsDnsBlacklist: EXTVAL_VALID,
						ExtraSan:                    &ExtraSan{tag: sanTag, value: sanValue},
					})
				}
			}
		}
	}

	// Adding more test cases? Be a pal and remember to update the testCases slice's capacity at the top.

	return &TestCaseProvider{
		testCases: testCases,
	}
}

func (p *TestCaseProvider) Name() string {
	return "nameconstraints"
}

func (p *TestCaseProvider) GetTestCaseCount() (uint, error) {
	return uint(len(p.testCases)), nil
}

func (p *TestCaseProvider) GetTestCase(index uint) (test_case.TestCase, error) {
	return p.testCases[index], nil
}

func (p *TestCaseProvider) GetSanityCheckTestCase() (uint, error) {
	return SANITY_CHECK_TEST_CASE, nil
}

const (
	FEATURE_NAME_CONSTRAINTS test_case.Feature = iota + 1
	FEATURE_VALIDATE_DNS
	FEATURE_VALIDATE_IP
)

func (p *TestCaseProvider) GetFeatures() []test_case.Feature {
	return []test_case.Feature{FEATURE_NAME_CONSTRAINTS, FEATURE_VALIDATE_DNS, FEATURE_VALIDATE_IP}
}

func (p *TestCaseProvider) DescribeFeature(feature test_case.Feature) string {
	switch feature {
	case FEATURE_NAME_CONSTRAINTS:
		return "NAME_CONSTRAINTS"
	case FEATURE_VALIDATE_DNS:
		return "VALIDATE_DNS"
	case FEATURE_VALIDATE_IP:
		return "VALIDATE_IP"
	}
	panic(fmt.Errorf("unsupported feature: %d", feature))
}

func (p *TestCaseProvider) GetTestCasesForFeature(feature test_case.Feature) ([]uint, error) {
	if feature == FEATURE_NAME_CONSTRAINTS {
		return []uint{FEATURE_NAME_CONSTRAINTS_TEST_CASE}, nil
	}
	if feature == FEATURE_VALIDATE_DNS {
		return []uint{FEATURE_VALIDATE_DNS_TEST_CASE_1, FEATURE_VALIDATE_DNS_TEST_CASE_2}, nil
	}
	if feature == FEATURE_VALIDATE_IP {
		return []uint{FEATURE_VALIDATE_IP_TEST_CASE_1, FEATURE_VALIDATE_IP_TEST_CASE_2}, nil
	}

	return nil, fmt.Errorf("invalid feature: %v", feature)
}
