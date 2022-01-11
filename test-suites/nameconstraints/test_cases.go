package nameconstraints

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"github.com/Netflix/bettertls/test-suites/certutil"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	"net"
)

type InvalidReason int

const VALID_DNS_NAME = "test.localhost"
const INVALID_DNS_NAME = "bad.example.com"
const VALID_IP = "127.0.0.1"
const INVALID_IP = "1.1.1.1"
const VALID_DNS_TREE = "localhost"
const INVALID_DNS_TREE = "example.com"
const VALID_IP_RANGE = "127.0.0.0/24"
const INVALID_IP_RANGE = "1.1.1.0/24"

type TrinaryValue byte

const (
	EXTVAL_NONE TrinaryValue = iota
	EXTVAL_VALID
	EXTVAL_INVALID
)

func (tv TrinaryValue) String() string {
	switch tv {
	case EXTVAL_NONE:
		return "NONE"
	case EXTVAL_VALID:
		return "VALID"
	case EXTVAL_INVALID:
		return "INVALID"
	}
	panic(fmt.Errorf("unhandled TrinaryValue: %d", tv))
}
func (tv TrinaryValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(tv.String())
}

var ALL_TRINARY_VALUES = []TrinaryValue{EXTVAL_NONE, EXTVAL_VALID, EXTVAL_INVALID}

type CommonNameType byte

const (
	CN_TYPE_DNS CommonNameType = iota
	CN_TYPE_IP
)

func (cnt CommonNameType) String() string {
	switch cnt {
	case CN_TYPE_DNS:
		return "DNS"
	case CN_TYPE_IP:
		return "IP"
	}
	panic(fmt.Errorf("unhandled CommonNameType: %d", cnt))
}
func (cnt CommonNameType) MarshalJSON() ([]byte, error) {
	return json.Marshal(cnt.String())
}

type ClientHostnameType byte

const (
	CLIENT_HOSTNAME_TYPE_DNS ClientHostnameType = iota
	CLIENT_HOSTNAME_TYPE_IP
)

func (cht ClientHostnameType) String() string {
	switch cht {
	case CLIENT_HOSTNAME_TYPE_DNS:
		return "DNS"
	case CLIENT_HOSTNAME_TYPE_IP:
		return "IP"
	}
	panic(fmt.Errorf("unhandled ClientHostnameType: %d", cht))
}
func (cht ClientHostnameType) MarshalJSON() ([]byte, error) {
	return json.Marshal(cht.String())
}

type NameConstraintsTestCase struct {
	ClientHostnameType          ClientHostnameType
	CommonNameType              CommonNameType
	CommonNameValue             TrinaryValue
	DnsSan                      TrinaryValue
	IpSan                       TrinaryValue
	NameConstraintsIpWhitelist  TrinaryValue
	NameConstraintsDnsWhitelist TrinaryValue
	NameConstraintsIpBlacklist  TrinaryValue
	NameConstraintsDnsBlacklist TrinaryValue
	ExtraSan                    *ExtraSan
}

type ExtraSan struct {
	tag   int
	value []byte
}

func (es *ExtraSan) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"tag":   es.tag,
		"value": es.value,
	})
}

func (n NameConstraintsTestCase) ExpectedResult() test_case.ExpectedResult {
	isSoftFail := false
	isSoftPass := false

	// First check that the client's requested hostname is in the certificate at all
	if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_DNS {
		if n.DnsSan != EXTVAL_VALID {
			if n.CommonNameType != CN_TYPE_DNS || n.CommonNameValue != EXTVAL_VALID {
				return test_case.EXPECTED_RESULT_FAIL
			}
			// CN matches the client hostname, but CN's shouldn't be used for hostname verification anymore, so mark as soft fail
			isSoftFail = true
		}
	} else if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_IP {
		if n.IpSan != EXTVAL_VALID {
			if n.CommonNameType != CN_TYPE_IP || n.CommonNameValue != EXTVAL_VALID {
				return test_case.EXPECTED_RESULT_FAIL
			}
			// CN matches the client hostname, but CN's shouldn't be used for hostname verification anymore, so mark as soft fail
			isSoftFail = true
		}
	}

	// If there is any name constraint invaliding the client's requested hostname, it should be a hard fail
	if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_DNS {
		if n.NameConstraintsDnsWhitelist == EXTVAL_INVALID || n.NameConstraintsDnsBlacklist == EXTVAL_VALID {
			return test_case.EXPECTED_RESULT_FAIL
		}
	} else if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_IP {
		if n.NameConstraintsIpWhitelist == EXTVAL_INVALID || n.NameConstraintsIpBlacklist == EXTVAL_VALID {
			return test_case.EXPECTED_RESULT_FAIL
		}
	}

	// Check if there is any NameConstraints violation other than the above. This is a "soft" fail because clients are
	// inconsistent about enforcing NameConstraints against subjects other than the one being used for hostname verification.
	// If the cert has a valid DNS name and NC extensions deny it
	hasSanViolation := false
	if n.DnsSan == EXTVAL_VALID && (n.NameConstraintsDnsWhitelist == EXTVAL_INVALID || n.NameConstraintsDnsBlacklist == EXTVAL_VALID) {
		hasSanViolation = true
	}
	// If the cert has an invalid DNS name and NC extensions deny it
	if n.DnsSan == EXTVAL_INVALID && (n.NameConstraintsDnsWhitelist == EXTVAL_VALID || n.NameConstraintsDnsBlacklist == EXTVAL_INVALID) {
		hasSanViolation = true
	}
	// If the cert has a valid IP and NC extensions deny it
	if n.IpSan == EXTVAL_VALID && (n.NameConstraintsIpWhitelist == EXTVAL_INVALID || n.NameConstraintsIpBlacklist == EXTVAL_VALID) {
		hasSanViolation = true
	}
	// If the cert has an invalid IP and NC extensions deny it
	if n.IpSan == EXTVAL_INVALID && (n.NameConstraintsIpWhitelist == EXTVAL_VALID || n.NameConstraintsIpBlacklist == EXTVAL_INVALID) {
		hasSanViolation = true
	}

	hasCnViolation := false
	if (n.CommonNameType == CN_TYPE_DNS && n.CommonNameValue == EXTVAL_VALID) && (n.NameConstraintsDnsWhitelist == EXTVAL_INVALID || n.NameConstraintsDnsBlacklist == EXTVAL_VALID) {
		hasCnViolation = true
	}
	if (n.CommonNameType == CN_TYPE_DNS && n.CommonNameValue == EXTVAL_INVALID) && (n.NameConstraintsDnsWhitelist == EXTVAL_VALID || n.NameConstraintsDnsBlacklist == EXTVAL_INVALID) {
		hasCnViolation = true
	}
	if (n.CommonNameType == CN_TYPE_IP && n.CommonNameValue == EXTVAL_VALID) && (n.NameConstraintsIpWhitelist == EXTVAL_INVALID || n.NameConstraintsIpBlacklist == EXTVAL_VALID) {
		hasCnViolation = true
	}
	if (n.CommonNameType == CN_TYPE_IP && n.CommonNameValue == EXTVAL_INVALID) && (n.NameConstraintsIpWhitelist == EXTVAL_VALID || n.NameConstraintsIpBlacklist == EXTVAL_INVALID) {
		hasCnViolation = true
	}

	// If there is any NC violation of SAN values, this _should_ fail.
	if hasSanViolation {
		isSoftFail = true
	} else if hasCnViolation {
		isSoftPass = true
	}

	if isSoftFail && isSoftPass {
		panic("no test should expect a soft pass and soft fail")
	}

	if isSoftFail {
		return test_case.EXPECTED_RESULT_SOFT_FAIL
	}
	if isSoftPass {
		return test_case.EXPECTED_RESULT_SOFT_PASS
	}
	return test_case.EXPECTED_RESULT_PASS
}

func (n NameConstraintsTestCase) GetHostname() string {
	if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_DNS {
		return VALID_DNS_NAME
	} else if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_IP {
		return VALID_IP
	} else {
		panic(fmt.Errorf("unhandled client hostname type: %v", n.ClientHostnameType))
	}
}

func (n NameConstraintsTestCase) RequiredFeatures() []test_case.Feature {
	var requiredFeatures []test_case.Feature
	if n.NameConstraintsIpWhitelist != EXTVAL_NONE || n.NameConstraintsDnsWhitelist != EXTVAL_NONE ||
		n.NameConstraintsIpBlacklist != EXTVAL_NONE || n.NameConstraintsDnsBlacklist != EXTVAL_NONE {
		requiredFeatures = append(requiredFeatures, FEATURE_NAME_CONSTRAINTS)
	}

	// If the test doesn't have a NC violation but has a client hostname / SAN mismatch, then validation is a required feature to get the expected result
	if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_DNS && n.DnsSan != EXTVAL_VALID {
		requiredFeatures = append(requiredFeatures, FEATURE_VALIDATE_DNS)
	}
	if n.ClientHostnameType == CLIENT_HOSTNAME_TYPE_IP && n.IpSan != EXTVAL_VALID {
		requiredFeatures = append(requiredFeatures, FEATURE_VALIDATE_IP)
	}

	return requiredFeatures
}

func (n NameConstraintsTestCase) GetCertificates(rootCert *x509.Certificate, rootKey crypto.Signer) (*tls.Certificate, error) {
	localRootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	localRootBytes, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: certutil.RandomSerial(),
		Subject: pkix.Name{
			CommonName:   "local_root",
			Organization: []string{certutil.SUBJECT_ORGANIZATION},
			SerialNumber: certutil.RandomString(),
		},
		NotBefore:             certutil.GetNotBefore(),
		NotAfter:              certutil.GetNotAfter(false),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, rootCert, localRootKey.Public(), rootKey)
	if err != nil {
		return nil, err
	}
	localRoot, err := x509.ParseCertificate(localRootBytes)
	if err != nil {
		return nil, err
	}

	_, validIpRange, err := net.ParseCIDR(VALID_IP_RANGE)
	if err != nil {
		return nil, err
	}
	_, invalidIpRange, err := net.ParseCIDR(INVALID_IP_RANGE)
	if err != nil {
		return nil, err
	}

	icaTemplate := &x509.Certificate{
		SerialNumber: certutil.RandomSerial(),
		Subject: pkix.Name{
			CommonName:   "local_ica",
			Organization: []string{certutil.SUBJECT_ORGANIZATION},
			SerialNumber: certutil.RandomString(),
		},
		NotBefore:             certutil.GetNotBefore(),
		NotAfter:              certutil.GetNotAfter(false),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	if n.NameConstraintsDnsWhitelist != EXTVAL_NONE {
		icaTemplate.PermittedDNSDomainsCritical = true
		if n.NameConstraintsDnsWhitelist == EXTVAL_VALID {
			icaTemplate.PermittedDNSDomains = []string{VALID_DNS_TREE}
		} else if n.NameConstraintsDnsWhitelist == EXTVAL_INVALID {
			icaTemplate.PermittedDNSDomains = []string{INVALID_DNS_TREE}
		}
	}
	if n.NameConstraintsDnsBlacklist != EXTVAL_NONE {
		icaTemplate.PermittedDNSDomainsCritical = true
		if n.NameConstraintsDnsBlacklist == EXTVAL_VALID {
			icaTemplate.ExcludedDNSDomains = []string{VALID_DNS_TREE}
		} else if n.NameConstraintsDnsBlacklist == EXTVAL_INVALID {
			icaTemplate.ExcludedDNSDomains = []string{INVALID_DNS_NAME}
		}
	}
	if n.NameConstraintsIpWhitelist != EXTVAL_NONE {
		icaTemplate.PermittedDNSDomainsCritical = true
		if n.NameConstraintsIpWhitelist == EXTVAL_VALID {
			icaTemplate.PermittedIPRanges = []*net.IPNet{validIpRange}
		} else if n.NameConstraintsIpWhitelist == EXTVAL_INVALID {
			icaTemplate.PermittedIPRanges = []*net.IPNet{invalidIpRange}
		}
	}
	if n.NameConstraintsIpBlacklist != EXTVAL_NONE {
		icaTemplate.PermittedDNSDomainsCritical = true
		if n.NameConstraintsIpBlacklist == EXTVAL_VALID {
			icaTemplate.ExcludedIPRanges = []*net.IPNet{validIpRange}
		} else if n.NameConstraintsIpBlacklist == EXTVAL_INVALID {
			icaTemplate.ExcludedIPRanges = []*net.IPNet{invalidIpRange}
		}
	}

	localIcaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	localIcaBytes, err := x509.CreateCertificate(rand.Reader, icaTemplate, localRoot, localIcaKey.Public(), localRootKey)
	if err != nil {
		return nil, err
	}
	localIca, err := x509.ParseCertificate(localIcaBytes)
	if err != nil {
		return nil, err
	}

	var commonName string
	if n.CommonNameValue == EXTVAL_VALID {
		if n.CommonNameType == CN_TYPE_DNS {
			commonName = VALID_DNS_NAME
		} else if n.CommonNameType == CN_TYPE_IP {
			commonName = VALID_IP
		}
	} else if n.CommonNameValue == EXTVAL_INVALID {
		if n.CommonNameType == CN_TYPE_DNS {
			commonName = INVALID_DNS_NAME
		} else if n.CommonNameType == CN_TYPE_IP {
			commonName = INVALID_IP
		}
	}

	leafTemplate := &x509.Certificate{
		SerialNumber: certutil.RandomSerial(),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{certutil.SUBJECT_ORGANIZATION},
			SerialNumber: certutil.RandomString(),
		},
		NotBefore:             certutil.GetNotBefore(),
		NotAfter:              certutil.GetNotAfter(false),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	var sans []*ExtraSan
	if n.DnsSan == EXTVAL_VALID {
		sans = append(sans, &ExtraSan{tag: nameTypeDNS, value: []byte(VALID_DNS_NAME)})
	}
	if n.DnsSan == EXTVAL_INVALID {
		sans = append(sans, &ExtraSan{tag: nameTypeDNS, value: []byte(INVALID_DNS_NAME)})
	}
	if n.IpSan == EXTVAL_VALID {
		rawIp := net.ParseIP(VALID_IP)
		if ip := rawIp.To4(); ip != nil {
			rawIp = ip
		}
		sans = append(sans, &ExtraSan{tag: nameTypeIP, value: rawIp})
	}
	if n.IpSan == EXTVAL_INVALID {
		rawIp := net.ParseIP(INVALID_IP)
		if ip := rawIp.To4(); ip != nil {
			rawIp = ip
		}
		sans = append(sans, &ExtraSan{tag: nameTypeIP, value: rawIp})
	}
	if n.ExtraSan != nil {
		sans = append(sans, n.ExtraSan)
	}
	if len(sans) > 0 {
		sanExt, err := buildSanExtension(false, sans)
		if err != nil {
			return nil, err
		}
		leafTemplate.ExtraExtensions = append(leafTemplate.ExtraExtensions, sanExt)
	}

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	leafCertBytes, err := x509.CreateCertificate(rand.Reader, leafTemplate, localIca, leafKey.Public(), localIcaKey)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{leafCertBytes, localIca.Raw, localRoot.Raw, rootCert.Raw},
		PrivateKey:  leafKey,
	}, nil

}
