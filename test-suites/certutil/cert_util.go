package certutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
)

const SUBJECT_ORGANIZATION = "bettertls.com"

func GetNotBefore() time.Time {
	return time.Now().Add(-24 * time.Hour)
}
func GetNotAfter(expired bool) time.Time {
	if expired {
		return time.Now().Add(-1 * time.Hour)
	} else {
		// About 30 years
		return time.Now().Add(366 * 24 * time.Hour)
	}
}

func RandomSerial() *big.Int {
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

func RandomString() string {
	s, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return s.String()
}

func GenerateSelfSignedCert(commonName string) (*x509.Certificate, crypto.Signer, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: RandomSerial(),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{SUBJECT_ORGANIZATION},
			SerialNumber: RandomString(),
		},
		NotBefore:             GetNotBefore(),
		NotAfter:              GetNotAfter(false),
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

func LoadCert(rootCa string) (*x509.Certificate, crypto.Signer, error) {
	var rootCert *x509.Certificate
	var rootKey crypto.Signer

	if _, err := os.Stat(rootCa); os.IsNotExist(err) {
		rootCert, rootKey, err = GenerateSelfSignedCert("bettertls_trust_root")
		if err != nil {
			return nil, nil, err
		}
		f, err := os.OpenFile(rootCa, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()
		err = pem.Encode(f, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: rootCert.Raw,
		})
		if err != nil {
			return nil, nil, err
		}
		rootKeyBytes, err := x509.MarshalPKCS8PrivateKey(rootKey)
		if err != nil {
			return nil, nil, err
		}
		err = pem.Encode(f, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: rootKeyBytes,
		})
		if err != nil {
			return nil, nil, err
		}
		f.Close()
	} else {
		data, err := ioutil.ReadFile(rootCa)
		if err != nil {
			return nil, nil, err
		}
		for len(data) > 0 && (rootCert == nil || rootKey == nil) {
			block, rest := pem.Decode(data)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				rootCert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, nil, err
				}
			}
			if block.Type == "PRIVATE KEY" {
				rootKeyPV, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return nil, nil, err
				}
				rootKey = rootKeyPV.(crypto.Signer)
			}
			data = rest
		}
		if rootCert == nil || rootKey == nil {
			return nil, nil, fmt.Errorf("rootCa file did not include certificate and key")
		}
	}

	return rootCert, rootKey, nil
}
