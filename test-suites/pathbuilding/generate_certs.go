package pathbuilding

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"

	"github.com/Netflix/bettertls/test-suites/certutil"
)

func GenerateCerts(rootCa *x509.Certificate, rootKey crypto.Signer, leafDnsName string, testCase *TestCaseImpl) (*tls.Certificate, error) {

	// Generate self-signed certs and keys for all entities in the graph
	entityKeys := make(map[string]crypto.Signer)
	entitySelfSignedCerts := make(map[string]*x509.Certificate)
	for _, caName := range testCase.ExplicitTestCase.TrustGraph.NodeNames() {
		caCert, caKey, err := certutil.GenerateSelfSignedCert(caName)
		if err != nil {
			return nil, err
		}
		entityKeys[caName] = caKey
		entitySelfSignedCerts[caName] = caCert
	}
	// Replace the src node with the test suite root cert/key
	entitySelfSignedCerts[testCase.ExplicitTestCase.SrcNode] = rootCa
	entityKeys[testCase.ExplicitTestCase.SrcNode] = rootKey

	generateIntermediate := func(src string, dst string, isInvalid bool) (*x509.Certificate, error) {
		issuerCert := entitySelfSignedCerts[src]
		issuerKey := entityKeys[src]
		dstCert := entitySelfSignedCerts[dst]

		template := &x509.Certificate{
			SerialNumber:          certutil.RandomSerial(),
			Subject:               dstCert.Subject,
			NotBefore:             certutil.GetNotBefore(),
			NotAfter:              certutil.GetNotAfter(false),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		if dst == testCase.ExplicitTestCase.DstNode {
			template.KeyUsage = x509.KeyUsageDigitalSignature
			template.IsCA = false
			template.Subject = pkix.Name{
				CommonName:   testCase.ExplicitTestCase.DstNode,
				Organization: []string{certutil.SUBJECT_ORGANIZATION},
				SerialNumber: certutil.RandomString(),
			}
			template.DNSNames = []string{leafDnsName}
			template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		}

		if isInvalid {
			switch testCase.InvalidReason {
			case INVALID_REASON_EXPIRED:
				template.NotAfter = certutil.GetNotAfter(true)
			case INVALID_REASON_NAME_CONSTRAINTS:
				template.PermittedDNSDomainsCritical = true
				template.PermittedDNSDomains = []string{"bad.example.com"}
			case INVALID_REASON_BAD_EKU:
				template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection}
			case INVALID_REASON_MISSING_BASIC_CONSTRAINTS:
				template.BasicConstraintsValid = false
			case INVALID_REASON_NOT_A_CA:
				template.IsCA = false
			case INVALID_REASON_DEPRECATED_CRYPTO:
				template.SignatureAlgorithm = x509.ECDSAWithSHA1
			default:
				return nil, fmt.Errorf("Unhandled invalid reason: %s", testCase.InvalidReason.String())
			}
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, template, issuerCert, entityKeys[dst].Public(), issuerKey)
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, err
		}
		return cert, nil
	}

	leafCerts := make([][]byte, 0, 1)
	intermediates := make([][]byte, 0, testCase.ExplicitTestCase.TrustGraph.EdgeCount())
	for _, edge := range testCase.ExplicitTestCase.TrustGraph.GetAllEdges() {
		isInvalid := edge.MemberOf(testCase.ExplicitTestCase.InvalidEdges)
		cert, err := generateIntermediate(edge.Source, edge.Destination, isInvalid)
		if err != nil {
			return nil, err
		}

		if edge.Destination == testCase.ExplicitTestCase.DstNode {
			leafCerts = append(leafCerts, cert.Raw)
		} else {
			intermediates = append(intermediates, cert.Raw)
		}
	}

	return &tls.Certificate{
		Certificate: append(leafCerts, intermediates...),
		PrivateKey:  entityKeys[testCase.ExplicitTestCase.DstNode],
	}, nil
}
