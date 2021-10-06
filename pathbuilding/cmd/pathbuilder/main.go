package main

import (
	"encoding/json"
	"fmt"
	"github.com/Netflix/bettertls/pathbuilding"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <server|showtest> ...\n", os.Args[0])
		os.Exit(1)
	}

	subcommand := os.Args[1]
	var err error
	if subcommand == "server" {
		err = runServer(os.Args[2:])
	} else if subcommand == "showtest" {
		err = showTest(os.Args[2:])
	} else {
		fmt.Printf("Invalid subcommand: %s\n", subcommand)
		os.Exit(1)
	}

	if err != nil {
		panic(err)
	}
}

func runServer(args []string) error {
	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		return err
	}
	server, err := pathbuilding.StartServer(provider, log.New(
		logrus.StandardLogger().WriterLevel(logrus.ErrorLevel), "", 0),
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

func showTest(args []string) error {
	if len(args) == 0 {
		fmt.Printf("Usage: %s showtest test_name\n", os.Args[0])
		return nil
	}

	testName := args[0]
	if idx := strings.LastIndex(testName, "/"); idx >= 0 {
		testName = testName[idx+1:]
	}
	provider, err := pathbuilding.NewTestCaseProvider()
	if err != nil {
		return err
	}
	manifest, err := provider.GetManifest()
	if err != nil {
		return err
	}

	testCase, err := provider.GetTestCase(testName)
	if err != nil {
		return err
	}

	// For the purposes of this utility, also include the self-signed trust anchor
	testCase.Certificates = append(testCase.Certificates, manifest.Root)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "   ")
	err = encoder.Encode(testCase)
	if err != nil {
		return err
	}

	return nil
}
