package impltests

import (
	"encoding/pem"
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type EnvoyRunner struct {
	version string
}

func (r *EnvoyRunner) Name() string {
	return "envoy"
}

func (r *EnvoyRunner) Initialize() error {
	version, err := execAndCapture("envoy", "--version")
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)

	r.version = version
	return nil
}

func (r *EnvoyRunner) Close() error {
	return nil
}

func (r *EnvoyRunner) GetVersion() string {
	return r.version
}

func (r *EnvoyRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	suites, err := test_executor.BuildTestSuites()
	if err != nil {
		return nil, err
	}

	rootCertPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: suites.GetRootCert().Raw,
	})
	pemString := strings.ReplaceAll(string(rootCertPem), "\n", "\\n")

	return test_executor.ExecuteAllTestsRemote(ctx, suites, func(hostname string, port uint) (bool, error) {
		sanType := "DNS"
		if net.ParseIP(hostname) != nil {
			sanType = "IP_ADDRESS"
		}

		var configYaml = fmt.Sprintf(`
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 127.0.0.1
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          access_log:
          - name: envoy.access_loggers.stdout
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: service_envoyproxy_io

  clusters:
  - name: service_envoyproxy_io
    type: LOGICAL_DNS
    # Comment out the following line to test on v6 networks
    dns_lookup_family: V4_ONLY
    load_assignment:
      cluster_name: service_envoyproxy_io
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: "%s"
                port_value: %d
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        common_tls_context:
          validation_context:
            match_typed_subject_alt_names:
            - san_type: %s
              matcher:
                exact: "%s"
            trusted_ca:
              inline_string: "%s"
`, hostname, port, sanType, hostname, pemString)

		cmd := exec.Command("envoy", "--config-yaml", configYaml)
		err = cmd.Start()
		if err != nil {
			return false, err
		}

		for {
			// Trial-and-error, 50ms is about enough to consistently have envoy startup
			time.Sleep(50 * time.Millisecond)
			c, err := net.Dial("tcp", "127.0.0.1:10000")
			if err == nil {
				c.Close()
				break
			}
		}

		resp, err := http.Get("http://127.0.0.1:10000/ok")
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		_ = cmd.Process.Signal(syscall.SIGTERM)
		_ = cmd.Wait()

		if err != nil || resp.StatusCode != http.StatusOK {
			return false, nil
		}
		return true, nil
	})
}
