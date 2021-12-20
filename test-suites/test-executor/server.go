package test_executor

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	test_case "github.com/Netflix/bettertls/test-suites/test-case"
	"github.com/Netflix/bettertls/test-suites/test-executor/web"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	listeners []net.Listener
	server    *http.Server
	wg        *sync.WaitGroup

	plaintextPort int
	tlsPort       int

	providerName string
	testIndex    uint
}

func (s *Server) SetTest(provider string, testIndex uint) {
	s.providerName = provider
	s.testIndex = testIndex
}

func StartServer(suites *TestSuites, serverLogger *log.Logger, plaintextPort uint16, tlsPort uint16) (*Server, error) {

	ptListener, err := net.Listen("tcp", fmt.Sprintf(":%d", plaintextPort))
	if err != nil {
		return nil, err
	}

	var server *Server
	tlsConfig := &tls.Config{
		// To make sure clients always do a full TLS handshake in order to check the cert verification, do not allow session tickets
		SessionTicketsDisabled: true,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			provider := suites.GetProvider(server.providerName)
			if provider == nil {
				return nil, fmt.Errorf("invalid provider: %s", server.providerName)
			}
			testCase, err := provider.GetTestCase(server.testIndex)
			if err != nil {
				return nil, fmt.Errorf("invalid test case %d: %v", server.testIndex, err)
			}
			return testCase.GetCertificates(suites.rootCert, suites.rootKey)
		},
	}
	tlsListener, err := tls.Listen("tcp", fmt.Sprintf(":%d", tlsPort), tlsConfig)
	if err != nil {
		ptListener.Close()
		return nil, err
	}

	router := http.NewServeMux()
	router.HandleFunc("/root.crt", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-x509-ca-cert")
		writer.Write(suites.rootCert.Raw)
	})
	router.HandleFunc("/root.pem", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-pem-file")
		pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: suites.rootCert.Raw})
	})
	router.HandleFunc("/suites", func(writer http.ResponseWriter, request *http.Request) {
		var resp struct {
			BetterTlsRevision string                 `json:"betterTlsRevision"`
			Suites            map[string]interface{} `json:"suites"`
		}
		resp.BetterTlsRevision = GetBuildRevision()
		resp.Suites = make(map[string]interface{})

		for _, providerName := range suites.GetProviderNames() {
			provider := suites.GetProvider(providerName)
			var suite struct {
				TestCount           uint                         `json:"testCount"`
				SanityCheckTestCase uint                         `json:"sanityCheckTestCase"`
				FeatureTestCases    map[test_case.Feature][]uint `json:"featureTestCases"`
			}
			suite.FeatureTestCases = make(map[test_case.Feature][]uint)

			testCount, err := provider.GetTestCaseCount()
			if err != nil {
				http.Error(writer, fmt.Sprintf("failed to get test count: %v", err), http.StatusInternalServerError)
				return
			}
			suite.TestCount = testCount
			suite.SanityCheckTestCase, err = provider.GetSanityCheckTestCase()
			if err != nil {
				http.Error(writer, fmt.Sprintf("failed to get sanity check test case: %v", err), http.StatusInternalServerError)
				return
			}
			for _, feature := range provider.GetFeatures() {
				testCases, err := provider.GetTestCasesForFeature(feature)
				if err != nil {
					http.Error(writer, fmt.Sprintf("failed to get feature test cases: %v", err), http.StatusInternalServerError)
					return
				}
				suite.FeatureTestCases[feature] = testCases
			}

			resp.Suites[provider.Name()] = &suite
		}

		json.NewEncoder(writer).Encode(&resp)
	})
	router.HandleFunc("/setTest", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, fmt.Sprintf("Invalid request method for this endpoint: %s", request.Method), http.StatusBadRequest)
			return
		}

		var reqBody struct {
			Provider string `json:"suite"`
			TestCase uint   `json:"testCase"`
		}
		err = json.NewDecoder(request.Body).Decode(&reqBody)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
			return
		}

		provider := suites.GetProvider(reqBody.Provider)
		if provider == nil {
			http.Error(writer, fmt.Sprintf("Invalid suite: %s", reqBody.Provider), http.StatusBadRequest)
			return
		}
		_, err := provider.GetTestCase(reqBody.TestCase)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Invalid test case: %d", reqBody.TestCase), http.StatusBadRequest)
			return
		}
		server.SetTest(reqBody.Provider, reqBody.TestCase)
	})
	router.HandleFunc("/getTest", func(writer http.ResponseWriter, request *http.Request) {
		q := request.URL.Query()
		providerName := q.Get("suite")
		testId, err := strconv.Atoi(q.Get("testCase"))
		if err != nil {
			http.Error(writer, fmt.Sprintf("Invalid test case: %s", q.Get("testCase")), http.StatusBadRequest)
			return
		}

		provider := suites.GetProvider(providerName)
		if provider == nil {
			http.Error(writer, fmt.Sprintf("Invalid suite: %s", providerName), http.StatusBadRequest)
			return
		}
		testCase, err := provider.GetTestCase(uint(testId))
		if err != nil {
			http.Error(writer, fmt.Sprintf("Invalid test case: %d", testId), http.StatusBadRequest)
			return
		}

		var respBody struct {
			Suite            string                   `json:"suite"`
			TestId           uint                     `json:"testId"`
			Hostname         string                   `json:"hostname"`
			ExpectedResult   test_case.ExpectedResult `json:"expectedResult"`
			RequiredFeatures []test_case.Feature      `json:"requiredFeatures"`
		}
		respBody.Suite = provider.Name()
		respBody.TestId = uint(testId)
		respBody.Hostname = testCase.GetHostname()
		respBody.ExpectedResult = testCase.ExpectedResult()
		respBody.RequiredFeatures = testCase.RequiredFeatures()

		json.NewEncoder(writer).Encode(&respBody)
	})
	router.HandleFunc("/ok", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		_, err := writer.Write([]byte("OK"))
		if err != nil {
			logrus.Errorf("Error writing response: %v", err)
		}
	})
	router.Handle("/", http.FileServer(http.FS(web.Content)))

	httpServer := &http.Server{Handler: router, ErrorLog: serverLogger}
	// Do not allow keep-alives since we want client testing to always have a do a new TLS handshake.
	httpServer.SetKeepAlivesEnabled(false)

	allListeners := []net.Listener{ptListener, tlsListener}
	wg := &sync.WaitGroup{}
	wg.Add(len(allListeners))
	for _, listener := range allListeners {
		go func(listener net.Listener) {
			err := httpServer.Serve(listener)
			if err != http.ErrServerClosed {
				logrus.Errorf("Error: %v", err)
			}
			wg.Done()
		}(listener)
	}

	server = &Server{
		listeners:     allListeners,
		server:        httpServer,
		wg:            wg,
		plaintextPort: ptListener.Addr().(*net.TCPAddr).Port,
		tlsPort:       tlsListener.Addr().(*net.TCPAddr).Port,
	}

	return server, nil
}

func (s *Server) PlaintextPort() int {
	return s.plaintextPort
}

func (s *Server) Stop() {
	s.server.Close()
	for _, listener := range s.listeners {
		listener.Close()
	}
	s.wg.Wait()
}
