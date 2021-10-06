package pathbuilding

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/google/uuid"
	"math/big"
	"time"
)

const SUBJECT_ORGANIZATION = "bettertls.com"

func getNotBefore() time.Time {
	return time.Now().Add(-24 * time.Hour)
}
func getNotAfter(expired bool) time.Time {
	if expired {
		return time.Now().Add(-1 * time.Hour)
	} else {
		// About 30 years
		return time.Now().Add(366 * 24 * time.Hour)
	}
}

func randomSerial() *big.Int {
	// We want 136 bits of random number, plus an 8-bit instance id prefix (which we always set to 0x01 here).
	randomBytes := make([]byte, 18)
	randomBytes[0] = 0x01
	_, err := rand.Read(randomBytes[1:])
	if err != nil {
		panic(err)
	}
	x := big.NewInt(0)
	x.SetBytes(randomBytes)
	return x
}

func randomString() string {
	s, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return s.String()
}

func generateSelfSignedCert(commonName string) (*x509.Certificate, crypto.Signer, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{SUBJECT_ORGANIZATION},
			SerialNumber: randomString(),
		},
		NotBefore:             getNotBefore(),
		NotAfter:              getNotAfter(false),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, template, template, caKey.Public(), caKey)
	if err != nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}
