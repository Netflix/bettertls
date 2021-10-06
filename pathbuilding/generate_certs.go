package pathbuilding

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
)

func GenerateCerts(rootCa *x509.Certificate, rootKey crypto.Signer, leafDnsName string, testCase *ExplicitTestCase) (*tls.Certificate, error) {

	// Generate self-signed certs and keys for all entities in the graph
	entityKeys := make(map[string]crypto.Signer)
	entitySelfSignedCerts := make(map[string]*x509.Certificate)
	for _, caName := range testCase.TrustGraph.NodeNames() {
		caCert, caKey, err := generateSelfSignedCert(caName)
		if err != nil {
			return nil, err
		}
		entityKeys[caName] = caKey
		entitySelfSignedCerts[caName] = caCert
	}
	// Replace the src node with the test suite root cert/key
	entitySelfSignedCerts[testCase.SrcNode] = rootCa
	entityKeys[testCase.SrcNode] = rootKey

	generateIntermediate := func(src string, dst string, isInvalid bool) (*x509.Certificate, error) {
		issuerCert := entitySelfSignedCerts[src]
		issuerKey := entityKeys[src]
		dstCert := entitySelfSignedCerts[dst]

		template := &x509.Certificate{
			SerialNumber:          randomSerial(),
			Subject:               dstCert.Subject,
			NotBefore:             getNotBefore(),
			NotAfter:              getNotAfter(false),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		if dst == testCase.DstNode {
			template.KeyUsage = x509.KeyUsageDigitalSignature
			template.IsCA = false
			template.Subject = pkix.Name{
				CommonName:   testCase.DstNode,
				Organization: []string{SUBJECT_ORGANIZATION},
				SerialNumber: randomString(),
			}
			template.DNSNames = []string{leafDnsName}
		}

		if isInvalid {
			switch testCase.InvalidReason {
			case INVALID_REASON_EXPIRED:
				template.NotAfter = getNotAfter(true)
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
	intermediates := make([][]byte, 0, testCase.TrustGraph.EdgeCount())
	for _, edge := range testCase.TrustGraph.GetAllEdges() {
		isInvalid := edge.MemberOf(testCase.InvalidEdges)
		cert, err := generateIntermediate(edge.Source, edge.Destination, isInvalid)
		if err != nil {
			return nil, err
		}

		if edge.Destination == testCase.DstNode {
			leafCerts = append(leafCerts, cert.Raw)
		} else {
			intermediates = append(intermediates, cert.Raw)
		}
	}

	return &tls.Certificate{
		Certificate: append(leafCerts, intermediates...),
		PrivateKey:  entityKeys[testCase.DstNode],
	}, nil
}
