package main

import (
	"crypto"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Netflix/bettertls/test-suites/certutil"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"github.com/sirupsen/logrus"
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
		rootCert, rootKey, err = certutil.LoadCert(rootCa)
		if err != nil {
			return err
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
