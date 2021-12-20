package main

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/Netflix/bettertls/test-suites/certutil"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func runServer(args []string) error {
	flagSet := flag.NewFlagSet("server", flag.ContinueOnError)
	var rootCa string
	flagSet.StringVar(&rootCa, "rootCa", "", "Use the given path as the root CA instead of generating an ephemeral root CA. If the file doesn't exist, a CA will generated and saved to the file.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	var rootCert *x509.Certificate
	var rootKey crypto.Signer
	if rootCa == "" {
		rootCert, rootKey, err = certutil.GenerateSelfSignedCert("bettertls_trust_root")
		if err != nil {
			return err
		}
	} else {
		if _, err := os.Stat(rootCa); os.IsNotExist(err) {
			rootCert, rootKey, err = certutil.GenerateSelfSignedCert("bettertls_trust_root")
			if err != nil {
				return err
			}
			f, err := os.OpenFile(rootCa, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			err = pem.Encode(f, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: rootCert.Raw,
			})
			if err != nil {
				return err
			}
			rootKeyBytes, err := x509.MarshalPKCS8PrivateKey(rootKey)
			if err != nil {
				return err
			}
			err = pem.Encode(f, &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: rootKeyBytes,
			})
			if err != nil {
				return err
			}
			f.Close()
		} else {
			data, err := ioutil.ReadFile(rootCa)
			if err != nil {
				return err
			}
			for len(data) > 0 && (rootCert == nil || rootKey == nil) {
				block, rest := pem.Decode(data)
				if block == nil {
					break
				}
				if block.Type == "CERTIFICATE" {
					rootCert, err = x509.ParseCertificate(block.Bytes)
					if err != nil {
						return err
					}
				}
				if block.Type == "PRIVATE KEY" {
					rootKeyPV, err := x509.ParsePKCS8PrivateKey(block.Bytes)
					if err != nil {
						return err
					}
					rootKey = rootKeyPV.(crypto.Signer)
				}
				data = rest
			}
			if rootCert == nil || rootKey == nil {
				return fmt.Errorf("rootCa file did not include certificate and key")
			}
		}
	}

	suites, err := test_executor.BuildTestSuitesWithRootCa(rootCert, rootKey)
	if err != nil {
		return err
	}

	server, err := test_executor.StartServer(suites,
		log.New(logrus.StandardLogger().WriterLevel(logrus.ErrorLevel), "", 0),
		8080, 8443)
	if err != nil {
		return err
	}
	defer server.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logrus.Infof("Now serving...")
	<-sigs
	logrus.Infof("Clean shutdown.")

	return nil
}
