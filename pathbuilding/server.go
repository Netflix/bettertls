package pathbuilding

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	listeners []net.Listener
	server    *http.Server
	wg        *sync.WaitGroup

	plaintextPort int
	tlsPort       int
}

func StartServer(provider *TestCaseProvider, serverLogger *log.Logger, plaintextPort uint16, tlsPort uint16) (*Server, error) {
	manifest, err := provider.GetManifest()
	if err != nil {
		return nil, err
	}

	ptListener, err := net.Listen("tcp", fmt.Sprintf(":%d", plaintextPort))
	if err != nil {
		return nil, err
	}

	tlsListener, err := tls.Listen("tcp", fmt.Sprintf(":%d", tlsPort), &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			testCaseName := strings.ToUpper(info.ServerName[:strings.Index(info.ServerName, ".")])
			var testCase *ExplicitTestCase
			if testCaseName == "VERIFY" {
				testCase = &ExplicitTestCase{
					TrustGraph: LINEAR_TRUST_GRAPH,
					SrcNode:    "Trust Anchor",
					DstNode:    "ICA",
				}
			} else {
				var err error
				testCase, err = DecodeHostname(testCaseName)
				if err != nil {
					return nil, err
				}
			}
			return provider.GetCertificatesForTestCase(testCase, info.ServerName)
		},
	})
	if err != nil {
		ptListener.Close()
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/root.crt", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-x509-ca-cert")
		writer.Write(manifest.Root)
	})
	mux.HandleFunc("/root.pem", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-pem-file")
		pem.Encode(writer, &pem.Block{Type: "CERTIFICATE", Bytes: manifest.Root})
	})
	mux.HandleFunc("/manifest.json", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(manifest)
	})
	mux.HandleFunc("/testcase", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			testCaseList, err := provider.GetTestCases()
			if err != nil {
				http.Error(writer, fmt.Sprintf("Failed to generate test case list: %v", err), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(writer).Encode(testCaseList)
		} else if request.Method == http.MethodPost {
			testCaseManifest := new(TestCaseManifest)
			err := json.NewDecoder(request.Body).Decode(testCaseManifest)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
				return
			}

			var trustGraph *TrustGraph
			for _, tg := range ALL_TRUST_GRAPHS {
				if tg.Name() == testCaseManifest.TrustGraph {
					trustGraph = tg
					break
				}
			}
			if trustGraph == nil {
				http.Error(writer, fmt.Sprintf("Invalid trust graph: %s", testCaseManifest.TrustGraph), http.StatusBadRequest)
				return
			}
			invalidEdges := make([]Edge, 0, len(testCaseManifest.InvalidEdges))
			for _, edge := range testCaseManifest.InvalidEdges {
				if len(edge) != 2 {
					http.Error(writer, fmt.Sprintf("Invalid edge: %v", edge), http.StatusBadRequest)
					return
				}
				invalidEdges = append(invalidEdges, Edge{edge[0], edge[1]})
			}
			var invalidReason InvalidReason
			if testCaseManifest.InvalidReason != "" {
				invalidReason = InvalidReasonFromString(testCaseManifest.InvalidReason)
				if invalidReason == INVALID_REASON_UNSPECIFIED {
					http.Error(writer, fmt.Sprintf("Invalid invalidReason: %s", testCaseManifest.InvalidReason), http.StatusBadRequest)
					return
				}
			}

			testCase := &ExplicitTestCase{
				TrustGraph:    trustGraph,
				SrcNode:       testCaseManifest.SrcNode,
				DstNode:       testCaseManifest.DstNode,
				InvalidEdges:  invalidEdges,
				InvalidReason: InvalidReasonFromString(testCaseManifest.InvalidReason),
			}
			testCaseName, err := EncodeHostname(testCase)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Failed to encode test case hostname: %v", err), http.StatusInternalServerError)
				return
			}

			testCaseManifest, err = provider.GetTestCase(testCaseName)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Failed to generate test case manifest: %v", err), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(writer).Encode(testCaseManifest)
		} else {
			http.Error(writer, fmt.Sprintf("Invalid method: %s", request.Method), http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/testcase/", func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		if !strings.HasPrefix(path, "/testcase/") {
			http.NotFound(writer, request)
			return
		}
		testCaseString := path[len("/testcase/"):]

		testCaseManifest, err := provider.GetTestCase(testCaseString)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Failed to generate test case manifest: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(writer).Encode(testCaseManifest)
	})
	mux.HandleFunc("/ok", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		_, err := writer.Write([]byte("OK"))
		if err != nil {
			logrus.Errorf("Error writing response: %v", err)
		}
	})
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	server := &http.Server{Handler: mux, ErrorLog: serverLogger}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	for _, listener := range []net.Listener{ptListener, tlsListener} {
		go func(listener net.Listener) {
			err := server.Serve(listener)
			if err != http.ErrServerClosed {
				logrus.Errorf("Error: %v", err)
			}
			wg.Done()
		}(listener)
	}

	return &Server{
		listeners:     []net.Listener{ptListener, tlsListener},
		server:        server,
		wg:            wg,
		plaintextPort: ptListener.Addr().(*net.TCPAddr).Port,
		tlsPort:       tlsListener.Addr().(*net.TCPAddr).Port,
	}, nil
}

func (s *Server) PlaintextPort() int {
	return s.plaintextPort
}

func (s *Server) TlsPort() int {
	return s.tlsPort
}

func (s *Server) Stop() {
	s.server.Close()
	for _, listener := range s.listeners {
		listener.Close()
	}
	s.wg.Wait()
}
