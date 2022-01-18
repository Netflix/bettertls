BetterTLS
===============

BetterTLS is a collection of test suites for TLS clients.
Find out more at [bettertls.com](https://bettertls.com).

This Repository
===============

The `docs/` directory contains the pages hosted at [bettertls.com](https://bettertls.com).
These pages contain most of the detailed information about what these test suites are and what their results mean.

Inside the `test-suites` directory you'll find code for the tests themselves and a harness for running those tests.
Check out the sections below for information on running those tests yourself and extending the BetterTLS code to run the tests against additional TLS implementations.

# Running tests

Tests are built and run using [Go](https://go.dev/), with minimum version 1.16.

Inside the `test-suites` directory of this repo you can run `go run ./cmd/bettertls` as the entrypoint to most functionality.
Use `--help` to learn more about supported sub-commands and use `--help` on each sub-command to learn more about supported parameters.

```
$ go run ./cmd/bettertls --help
Usage: /tmp/go-build1035531161/b001/exe/bettertls <server|get-test> ...
Supported sub-commands: server, get-test, run-tests, generate-manifests, show-results

$ go run ./cmd/bettertls run-tests --help
Usage of run-tests:
  -implementation string
    	Implementation to test.
  -outputDir string
    	Directory to which test results will be written. (default ".")
  -suite string
    	Run only the given suite instead of all suites.
  -testCase int
    	Run only the given test case in the suite instead of all tests. Requires --suite to be sepecified as well. (default -1)
```

To run all tests against all implementations and all test suites, use the `run-tests` subcommand, e.g. `go run ./cmd/bettertls run-tests`.
You can run tests for a single implementation or single test suite with flags.
For example:

```
$ go run ./cmd/bettertls run-tests --implementation curl --suite pathbuilding
curl/pathbuilding 100% |█████████████████████████████████████████████████████████████████████████████████████████████████████████| (81/81, 227 tests/s)        
Implementation: curl
Version: curl 7.68.0 (x86_64-pc-linux-gnu) libcurl/7.68.0 OpenSSL/1.1.1f zlib/1.2.11 brotli/1.0.7 libidn2/2.2.0 libpsl/0.21.0 (+libidn2/2.2.0) libssh/0.9.3/openssl/zlib nghttp2/1.40.0 librtmp/2.3
Release-Date: 2020-01-08
Protocols: dict file ftp ftps gopher http https imap imaps ldap ldaps pop3 pop3s rtmp rtsp scp sftp smb smbs smtp smtps telnet tftp 
Features: AsynchDNS brotli GSS-API HTTP2 HTTPS-proxy IDN IPv6 Kerberos Largefile libz NTLM NTLM_WB PSL SPNEGO SSL TLS-SRP UnixSockets

Suite: pathbuilding
  Supported Features: INVALID_REASON_EXPIRED, INVALID_REASON_NAME_CONSTRAINTS, INVALID_REASON_BAD_EKU, INVALID_REASON_MISSING_BASIC_CONSTRAINTS, INVALID_REASON_NOT_A_CA, INVALID_REASON_DEPRECATED_CRYPTO, 
  Unsupported Features: BRANCHING, 
  Passed: 20
  Skipped: 61
  Failures: 0
```

The test results are saved to `curl_results.json`.

# Running tests in a browser

Browsers can be tested by running the test server:

```
go run ./cmd/bettertls server
```

You can then browse to `http://localhost:8080` which will start the test suites.
When it is complete, the textbox will display a JSON dump of the test results.
You can save this output to a file and use the commands below to interpret the results.

# Viewing test results

The summary of test results is printed when tests are run.
You can also print the summary using the `show-results` subcommand:

```
$ go run ./cmd/bettertls show-results --resultsFile ../docs/results/go_results.json
Implementation: go
Version: go1.18beta1
Suite: pathbuilding
  Supported Features: BRANCHING, INVALID_REASON_EXPIRED, INVALID_REASON_NAME_CONSTRAINTS, INVALID_REASON_BAD_EKU, INVALID_REASON_MISSING_BASIC_CONSTRAINTS, INVALID_REASON_NOT_A_CA, INVALID_REASON_DEPRECATED_CRYPTO, 
  Unsupported Features: 
  Passed: 80
  Skipped: 0
  Failures: 1
    False positives: 1

Suite: nameconstraints
  Supported Features: NAME_CONSTRAINTS, 
  Unsupported Features: 
  Passed: 8750
  Skipped: 0
  Failures: 0
```

You can pass the `--json` flag to this subcommand to get a JSON-formatted summary, which includes a listing of specifically which tests passed/failed.

```
$ go run ./cmd/bettertls show-results --resultsFile ../docs/results/go_results.json --json | jq .suiteSummary.pathbuilding.falsePositiveTests
[
  57
]
```

To get more information about a test (and in particular the certificates used in the test), use the `get-test` subcommand:

```
$ go run ./cmd/bettertls get-test --suite pathbuilding --testId 57 | jq .
{
  "suite": "pathbuilding",
  "testId": 57,
  "definition": {
    "ExplicitTestCase": {
      "TrustGraph": {},
      "SrcNode": "Trust Anchor",
      "DstNode": "EE",
      "InvalidEdges": [
        {
          "Source": "Trust Anchor",
          "Destination": "A"
        },
        {
          "Source": "C",
          "Destination": "B"
        }
      ],
      "InvalidReason": 0,
      "ExpectFailure": false,
      "Comment": "Should be able to find an alternate path through a more complicated tree."
    },
    "InvalidReason": 3
  },
  "expectedResult": "PASS",
  "certificates": [
    "MIIB/jCCAaSgAwIBAgISAe3sTVQGhDGUb5yFSU1QARO3MAoGCCqGSM49BAMCMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUIxLTArBgNVBAUTJDVhZjFjMDNhLWQ2ZWItNGQ5MC1iZmFjLWIyZGE0NDE2MDA2NTAeFw0yMTEyMjIxOTM2MTNaFw0yMjEyMjQxOTM2MTNaMFQxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCzAJBgNVBAMTAkVFMS0wKwYDVQQFEyQyMjExYzAyNy01MmY5LTQ1NjktYWQ0Ni05OTQ2ZjE5ZGI4OGMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQEUN7X7St/PzihRAif8DjMQEM+Jbh0+020VWFeFfU38QIGXO2O3rjkGP2QBicGL+qduHc/bo5FsTt9qjc0JPazo1cwVTAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAfBgNVHSMEGDAWgBSPwhAydqu1J69CeZe5AERko29waTAUBgNVHREEDTALgglsb2NhbGhvc3QwCgYIKoZIzj0EAwIDSAAwRQIhAMa64jdSBVsOJD4SKlRdiWBFLWvovO3T80110uy+otaDAiAbGBp/ceHV+z2Hh7J4FFLoVafxjRMj5PxVyNlGyZo2rQ==",
    "MIICHTCCAcSgAwIBAgISARUMpj8BRbiS3WfikFBlv03wMAoGCCqGSM49BAMCMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUMxLTArBgNVBAUTJDA3YWEzOGFlLTM0ZmMtNGU3ZC05ZWZmLTU1NDlhNjFiMDc3MTAeFw0yMTEyMjIxOTM2MTNaFw0yMjEyMjQxOTM2MTNaMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUIxLTArBgNVBAUTJDVhZjFjMDNhLWQ2ZWItNGQ5MC1iZmFjLWIyZGE0NDE2MDA2NTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI77YnZUOCwd4NVCEC8f/E1JgNCtv+rWs0P3xt0ePuANHecJ0vU360K9cS3n2cXQNQu4BJMVPJYE0j25DVB+6vSjeDB2MA4GA1UdDwEB/wQEAwICBDATBgNVHSUEDDAKBggrBgEFBQcDBDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSPwhAydqu1J69CeZe5AERko29waTAfBgNVHSMEGDAWgBSogEGB8+QJnjAeqTXPo7mcXrMauDAKBggqhkjOPQQDAgNHADBEAiBAlVqspaw6Rw5OcHy+Z2Rv55Llo0jmt+Qy4F9Jw2hSuQIgciD+WXGF7ImvRrjcKovFvfzKBS0TYj6AY/gJPRtr8q8=",
    "MIICCTCCAa+gAwIBAgISAWxXScUOantLm8vX9PNp6JhqMAoGCCqGSM49BAMCMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUExLTArBgNVBAUTJGRjNjI0ZTIwLWE2YmItNGYyYS1hYmU2LTBmMzMwYzQ0NzQ1MTAeFw0yMTEyMjIxOTM2MTNaFw0yMjEyMjQxOTM2MTNaMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUIxLTArBgNVBAUTJDVhZjFjMDNhLWQ2ZWItNGQ5MC1iZmFjLWIyZGE0NDE2MDA2NTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABI77YnZUOCwd4NVCEC8f/E1JgNCtv+rWs0P3xt0ePuANHecJ0vU360K9cS3n2cXQNQu4BJMVPJYE0j25DVB+6vSjYzBhMA4GA1UdDwEB/wQEAwICBDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSPwhAydqu1J69CeZe5AERko29waTAfBgNVHSMEGDAWgBRjFLJd+LqXOY7GAfWH+rx5a98ZsTAKBggqhkjOPQQDAgNIADBFAiAiO1PupPIWNf/CJWKVztr9phAVunJWMHLkoBYqQdf2BgIhAI23WPRyRkDUx7t5lj4/MJF7dflVk52bzEM6im4++Tez",
    "MIICCTCCAa+gAwIBAgISASXF6lU+pfqmo9xcL4fYakJUMAoGCCqGSM49BAMCMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUMxLTArBgNVBAUTJDA3YWEzOGFlLTM0ZmMtNGU3ZC05ZWZmLTU1NDlhNjFiMDc3MTAeFw0yMTEyMjIxOTM2MTNaFw0yMjEyMjQxOTM2MTNaMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUExLTArBgNVBAUTJGRjNjI0ZTIwLWE2YmItNGYyYS1hYmU2LTBmMzMwYzQ0NzQ1MTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABFshv7uX/SqP0SsksuPZMYD/Lb1s1ogupVcMEOAXBOwHeSbXZryJaxzJAf+VcUAlDEY5a+MMsJ87/8BHg/Mahf6jYzBhMA4GA1UdDwEB/wQEAwICBDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRjFLJd+LqXOY7GAfWH+rx5a98ZsTAfBgNVHSMEGDAWgBSogEGB8+QJnjAeqTXPo7mcXrMauDAKBggqhkjOPQQDAgNIADBFAiEAvJKYp/xd6dcoiZ9+uRsnScvlkXeB8s5lbXsKafFDLtACIEi0WeS1zL2rNW9vDPeEPkiSkOD874s3HV243KUK/+PF",
    "MIICCTCCAa+gAwIBAgISAbi1OlomU/bxnXq6G55rh/sjMAoGCCqGSM49BAMCMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUExLTArBgNVBAUTJGRjNjI0ZTIwLWE2YmItNGYyYS1hYmU2LTBmMzMwYzQ0NzQ1MTAeFw0yMTEyMjIxOTM2MTNaFw0yMjEyMjQxOTM2MTNaMFMxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xCjAIBgNVBAMTAUMxLTArBgNVBAUTJDA3YWEzOGFlLTM0ZmMtNGU3ZC05ZWZmLTU1NDlhNjFiMDc3MTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABBu+gxhOrTmBOGaZq6CPs+pLjTNsIXrd30ELK7UGl1ZOBJ5Ftg+/eSEyIh9GtOQX1yMAEyVH4GO3ljF4Sa/PO5ujYzBhMA4GA1UdDwEB/wQEAwICBDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSogEGB8+QJnjAeqTXPo7mcXrMauDAfBgNVHSMEGDAWgBRjFLJd+LqXOY7GAfWH+rx5a98ZsTAKBggqhkjOPQQDAgNIADBFAiEA146AMDqjYp1puF30fVpn4qcpP/dAftX/OTVJKlj54nQCIAXHwg5fxMvGod7TBpCO2Y3y7JTIMSI5kbyLjfFAXSh7",
    "MIICHDCCAcKgAwIBAgISAQ3vLs0jPq8mLEIoaumsxRx5MAoGCCqGSM49BAMCMGYxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xHTAbBgNVBAMMFGJldHRlcnRsc190cnVzdF9yb290MS0wKwYDVQQFEyQxODA3NWM2YS0zMzM0LTQ1NTQtOWZhNS01MmFiM2Q0OGU3NmMwHhcNMjExMjIyMTkzNjEzWhcNMjIxMjI0MTkzNjEzWjBTMRYwFAYDVQQKEw1iZXR0ZXJ0bHMuY29tMQowCAYDVQQDEwFDMS0wKwYDVQQFEyQwN2FhMzhhZS0zNGZjLTRlN2QtOWVmZi01NTQ5YTYxYjA3NzEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQbvoMYTq05gThmmaugj7PqS40zbCF63d9BCyu1BpdWTgSeRbYPv3khMiIfRrTkF9cjABMlR+Bjt5YxeEmvzzubo2MwYTAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUqIBBgfPkCZ4wHqk1z6O5nF6zGrgwHwYDVR0jBBgwFoAUQDQ9YJNBnlpN0M+q9JORgwvYIJMwCgYIKoZIzj0EAwIDSAAwRQIhAPnY2hcjD4iEgAgU45Gv5nRTtBbEsgnPbavaKI1fh4+EAiAIdjM6WOD33Fny1PN7G4KuLi+Goi9DTGDw9zfqKNYoKg==",
    "MIICMDCCAdegAwIBAgISAUwmw4h/CoyZb6RI9Ori4Yv5MAoGCCqGSM49BAMCMGYxFjAUBgNVBAoTDWJldHRlcnRscy5jb20xHTAbBgNVBAMMFGJldHRlcnRsc190cnVzdF9yb290MS0wKwYDVQQFEyQxODA3NWM2YS0zMzM0LTQ1NTQtOWZhNS01MmFiM2Q0OGU3NmMwHhcNMjExMjIyMTkzNjEzWhcNMjIxMjI0MTkzNjEzWjBTMRYwFAYDVQQKEw1iZXR0ZXJ0bHMuY29tMQowCAYDVQQDEwFBMS0wKwYDVQQFEyRkYzYyNGUyMC1hNmJiLTRmMmEtYWJlNi0wZjMzMGM0NDc0NTEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARbIb+7l/0qj9ErJLLj2TGA/y29bNaILqVXDBDgFwTsB3km12a8iWscyQH/lXFAJQxGOWvjDLCfO//AR4PzGoX+o3gwdjAOBgNVHQ8BAf8EBAMCAgQwEwYDVR0lBAwwCgYIKwYBBQUHAwQwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUYxSyXfi6lzmOxgH1h/q8eWvfGbEwHwYDVR0jBBgwFoAUQDQ9YJNBnlpN0M+q9JORgwvYIJMwCgYIKoZIzj0EAwIDRwAwRAIgKuW4kS+VsCkoQ2jgK9yEGh0vYzO94PgJEGBrrt+aC2kCICgRZPqpVQBK1KM35mo/JsfCW1MGaNjy4q4mdLzIblol"
  ]
}
```

# Testing additional TLS implementations

## Execting tests with the embedded test runner

To add a new implementation to be tested, you will need to implement the [impltests.ImplementationRunner](test-suites/impltests/runner.go) interface.
The easiest way to implement this interface to follow the pattern of one of the existing implementations, which invokes the TLS implementation as a separate process for each test case.

When running tests, the Go executable runs an HTTPS server that will present certificates for each test case.
The easiest way to test a new implementation is to create a script or executable that will attempt to establish a TLS connection to the server, given a port/hostname and trusted CA as parameters.
Check out the [curl](test-suites/impltests/curl.go) implementation as an example.

If it makes more sense to test your implementation by passing in the certificates (rather than establishing a TLS connection to a remote service), check out the [PKI.js](test-suites/impltests/pkijs.go) for an example of how certificates get passed to that executable.

Once you have create your implementation, add it to the [`Runners` map](test-suites/impltests/runner.go).
You will then be able to run `go run ./cmd/bettertls run-tests --implementation my_impl`.

## Exporting tests to run outside of the BetterTLS executor

For many use cases it might be easier to export tests from BetterTLS so that you can run tests in your implementations own test harness.
The bettertls executable has an `export-tests` command which will write all tests in a JSON format that you can use in your own project.

If you have checked out the bettertls repository then you can run:
```
go run ./cmd/bettertls export-tests --out tests.json
```

Or without checking out the repository:
```
GOBIN=$PWD go install github.com/Netflix/bettertls/test-suites/cmd/bettertls@latest
./bettertls export-tests --out tests.json
```

The following is an abbreviated example of what gets exported:

```
{
  "betterTlsRevision": "939077295c05d36c53f1f386fc3ec55167360f3a",
  "trustRoot": "MII...",
  "suites": {
    "nameconstraints": {
      "features": ["NAME_CONSTRAINTS", "VALIDATE_DNS", "VALIDATE_IP"],
      "sanityCheckTestCase": 0,
      "featureTestCases": {
        "NAME_CONSTRAINTS": [1],
        "VALIDATE_DNS": [2,3],
        "VALIDATE_IP": [4,5]
      },
      "testCases": [
        {
          "id": 0,
          "suite": "nameconstraints",
          "certificates": [
            "MII...",
            ...
          ],
          "hostname": "test.localhost",
          "requiredFeatures": [],
          "expected": "ACCEPT",
          "failureIsWarning": false
        },
        ...
      ]
    }
  }
}
```

**Top-Level Export**

| Field Name | Description |
| --- | --- |
| betterTlsRevision | The BetterTLS git revision used when building/running the binary. Tests may change over time, so this can help determine exactly how a test case was generated. |
| trustRoot | A certificate that should be used as the (only) trust anchor for all tests. Base64-encoded DER format. |
| suites | An object of the exported test suites. Keys are the suite name and values are described in the TestSuite table below. |

**TestSuite**

| Field Name | Description |
| --- | --- |
| features | The various implementation features for the test suite. Refer to the features section below for what each feature means. |
| sanityCheckTestCase | The index of a test that does not require any features and should pass. This is mostly useful to ensure your test harness and trust store has been set up properly. |
| featureTestCases | A map from each feature to an array of test indices. To determine if an implementation supports a given feature, all the given test cases should pass. |
| testCases | An array of the test cases, as described below. |

**TestCase**

| Field Name | Description                                                                                                                                                  |
| --- |--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| id | The id of the test case. This id (along with "suite" below) can be passed to the `bettertls get-test` command. Note that ids are only unique within a suite. |
| suite | The name of the test suite this test case belongs to.                                                                                                        |
| certificates | The array of certificates for the test case, leaf first. Certificates are Base64-encoded DER format.                                                         |
| hostname | The hostname that should be used by the client for subject name verification. This may be a DNS name or a stringified IP address.                            |
| requiredFeatures | An array of features that the TLS implementation needs in order to run this test. The test should be skipped if any feature is not supported.                |
| expected | The expected behavior of the TLS implementation. Either "ACCEPT" or "REJECT"                                                                                 |
| failureIsWarning | If true, getting an unexpected result on this test should just be considered a warning. See the note below.                                                  |

### A note about the "failureIsWarning" tests

There are a number of tests, especially in the nameconstraints test suite, which have "failureIsWarning" on tests with ambiguous expected results.
This mostly stems from inconsistency about whether 1) implementations use the subject Common Name (CN) for hostname verification (and therefore apply name constraints checks against it), and 2) whether implementations only check nameconstraints against the hostname being verified (they _should_ be checking any SANs in the certificate for violations, but many implementations do not).

For example, a test might have the common name `bad.example.com` as the subject CN and `test.localhost` as the only DNS SAN.
If the issuing CA has `localhost` as the whitelisted DNS tree, implementations would reject this chain if they apply name constraints checks against the CN.
This test (with `hostname="test.localhost"`) has `expected="ACCEPT"` but `failureIsWarning=true` for this reason.

Modern implementations should not be using the subject CN for hostname verification, so test cases where the expected hostname is not forbidden by name constraints but is only present in the CN (and no SANs) have `expected="REJECT"` but `failureIsWarning=true`.

Finally, a test might have SANs which violate name constraints but are irrelevant to the hostname being verified.
For example, a leaf certificate with a DNS SAN of `test.localhost` and an IP SAN of `192.168.0.1` under a issuer CA with name constraints `DNS=localhost` and `IP=127.0.0.0/24` would have an expected `FAILURE` result because of the IP address name constraint violation.
However, this test will have `failureIsWarning=true` when verifying `hostname="test.localhost"`.

In short, unexpected results on such tests suggest the implementation does not have what we consider to be the "most correct" outcome, but it is not necessarily wrong.

### Feature descriptions

When an implementation is run with the bettertls test executor, supported features are detected during testing.
Tests which require unsupported features are then automatically skipped.

If you are exporting tests from BetterTLS for use in your project, you can use the `featureTestCases` tests to dynamically detect supported features in your project, but it will probably make more sense to look at the descriptions below and determine which features you expect to be supported.
You can then skip any tests which require features not in your list.

### nameconstraints

* **NAME_CONSTRAINTS**: Whether the name constraints extension is supported at all (i.e. are intermediate CA certificates with a critical name constraints extension accepted).
* **VALIDATE_DNS**: Does the implementation perform hostname verification when `hostname` is a DNS name?
* **VALIDATE_IP**: Does the implementation perform hostname verification (that is, against IP address SANs) when `hostname` is a stringified IP address?

### pathbuilding

* **BRANCHING**: Does the implementation support path building at all? That is, when presented with an array of certificates that are not a linear chain, does the implementation discover a chain within it to a trust anchor?
* **INVALID_REASON_xxx**: Does the implementation reject certificates for a given reason. Check out the "Path Building" section on the [bettertls website](https://bettertls.com) for more information on each of the "invalid reasons".
