# Pathbuilding

This directory contains code that creates a suite of path building test cases for TLS implementations.
Path building refers to the problem of building a valid, trusted chain from an end-entity's leaf certificate to a trust anchor.
For background and motivation, we highly recommend you first read [Ryan Sleevi's blog post](https://medium.com/@sleevi_/path-building-vs-path-verifying-the-chain-of-pain-9fbab861d7d6) on this topic.
The problem stated can be succinctly summarized with this excerpt:

> Often, when I talk to people who are responsible for configuring certificates on their servers, they often talk about _the_ certificate chain. ... [But] there are many chains, with [different chains are needed by different clients](https://blog.cloudflare.com/introducing-cfssl/), who have different root stores and different behaviors.

 Historically the TLS specifications have not required that TLS implementations support particularly robust certificate path building.
 In practice many of them (such as OpenSSL) don't support verifying anything other than a simple, non-branching sequence of certificates.
 This has changed in the [specification for TLS 1.3](https://datatracker.ietf.org/doc/html/rfc8446#section-4.4.2) which advises that implementations SHOULD support more robust certificate path building:
 
> Note: Prior to TLS 1.3, "certificate_list" ordering required each certificate to certify the one immediately preceding it; however, some implementations allowed some flexibility.  Servers sometimes
send both a current and deprecated intermediate for transitional purposes, and others are simply configured incorrectly, but these cases can nonetheless be validated properly.  For maximum compatibility, all implementations SHOULD be prepared to handle potentially extraneous certificates and arbitrary orderings from any TLS version, with the exception of the end-entity certificate which MUST be first.

This directory provides a suite of tests that can be used against TLS implementations to not only verify whether they satisfy the above provision, but whether they are doing so correctly.

Check out our [blog post](https://netflixtechblog.com/revisiting-bettertls-certificate-path-building-4c978b79843f) for a summary of results from common TLS implementations at the time this suite was created.

## Why does this matter?

The web PKI ecosystem is a shifting landscape.
Features like what certificate authorities are trusted, what algorithms should be used, and what X.509 certificate extensions can (and should) be enforced have been changing every year.
Service owners need to keep up with these changes while also ensuring that existing clients are able to reach their service.
For example, the [Let's Encrypt R3 CA](https://letsencrypt.org/certificates/) has two certificates; one signed by `DST Root CA X3` and one signed by `ISRG Root X1`.
Older clients only trusted the [self-signed DST Root CA X3 certificate](https://crt.sh/?id=8395), usually since they were built before the ISRG Root X1 CA had made its way into broadly distributed truststores.
However, newer clients only trust the ISRG Root X1 CA since the DST Root CA X3 self-signed certificate expired on September 30, 2021.
Ideally, service owners would be able to send _both_ the [DST Root CA X3 => R3](https://crt.sh/?id=3479778542) and [ISRG Root X1 => R3](https://crt.sh/?id=3334561879) certificate so that any client can verify their trust of the Let's Encrypt R3 CA and ultimately verify their trust in the end entity certificate.
In practice, many clients are not able to handle getting muiltple potential paths in a TLS response and this leaves both clients and service operators [subject to outages](https://status.catchpoint.com/incidents/f5yl89kygf12).  

In general, having clients with a robust certificate path building allows the community to be more agile and make changes to the web PKI ecosystem over time while reducing risk of breaking older clients.
Here are just a few examples of these sorts of changes in the past:

* [Distrust of the Symantec CA](https://scotthelme.co.uk/are-you-ready-for-the-symantec-distrust/)
* [Deprecation of signing algorithms using SHA-1](https://www.venafi.com/education-center/ssl/sha-1-deprecation)
* [Expiration of the AddTrust CA](https://www.agwa.name/blog/post/fixing_the_addtrust_root_expiration)
* [Apple's restriction of certificate validity to 398 days](https://support.apple.com/en-us/HT211025)

At a minimum, this test suite can help inform service owners about how much path building different TLS client implementations support so that they can determine how clients will be impacted by changes to their service's TLS configuration and certificates.

# Quickstart

Run `go build ./cmd/pathbuilder` to build the binary.

Run `./pathbuilder server` to start a server which will dynamically generate the test cases as requested.

By default, the server has a plaintext listener on port 8080 and hosts a TLS listener on port 8443.
You can browse to `http://localhost:8080` which will run some javascript to run the test suite in your browser.

All tests use the same trust anchor which is randomly generated every time the server is started.
You can request the root certificate used for test cases from the server: `curl -O http://localhost:8080/root.pem`.
If you are running tests in your browser, you will need to visit this URL to download the root certificate and import into your browser's trust store.

The server provides a list of all test available in the suite via the `/testcase` endpoint:

```
$ curl -s http://localhost:8080/testcase | jq . | head
{
  "testCases": [
    "BAARAAQ",
    "BAARAARCAEASQAI",
    "BAARAARCAEASQAQ",
    "BAARAARCAEASQAY",
    "BAARAARCAEASQBA",
    "BAARAARCAEASQBI",
    "BAARAARCAEASQBQ",
    "BAARAAJCAEASQAI",
```

You can get information about a specific test via the `/testcase/{testName}` endpoint:

```
$ curl -s http://localhost:8080/testcase/BAARAAQ | jq .
{
  "name": "BAARAAQ",
  "trustGraph": "LINEAR_TRUST_GRAPH",
  "srcNode": "Trust Anchor",
  "dstNode": "EE",
  "invalidEdges": [],
  "invalidReason": "UNSPECIFIED",
  "expectedPath": [
    "Trust Anchor",
    "ICA",
    "EE"
  ],
  "certificates": [
    "MIICAT...",
    "MIICHz..."
  ]
}
```

To evaluate a test case, you can either supply the above certificates to your TLS implementation, or direct your TLS implementation to the running server which will present the certificates based on the SNI servername in the handshake.
For example:

```
$ curl --CAcert root.pem https://BAARAAQ.localhost:8443/ok
OK
$ curl --CAcert root.pem https://BAARAARCAEASQAI.localhost:8443/ok
curl: (60) SSL certificate problem: certificate has expired
More details here: https://curl.haxx.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.
```

Many TLS implementations can be tested using the provided Go test harness, which can `exec` a program or script for each test case.
See [impltests/curl_test.go](impltests/curl_test.go) as an example.

# About the tests

## Trust Graphs

All of the tests in this suite utilize a _Trust Graph_.
A Trust Graph is a directed graph where the nodes represent potentially trusted entities (a Distinguished Name and a public key) and the edges represent a certificate.
Consider the following example which defines the [TWO_ROOTS](trust_graph.go) trust graph:

```
+--------+
| Root 1 |======v
+--------+      +-----+         +----+
                | ICA | ======> | EE |
+--------+      +-----+         +----+
| Root 2 |======^
+--------+
```

There are two root CAs, both of which have created a certificate for an intermediate CA (ICA), which ultimately creates certificates for the end entity (EE).
The end entity (a service owner) doesn't know in general which root is trusted by clients, so it will always send _three_ certificates to clients: the leaf certificate for the end entity signed by ICA (`ICA => EE`), a certificate for the ICA signed by Root 1 (`Root 1 => ICA`), and a certificate for the ICA signed by Root 2 (`Root 2 => ICA`).
Clients which support path building should be able to verify trust in EE whether they trust Root 1 or Root 2 as a trust anchor.
Put another way, the presence of a certificate which leads to a non-trusted root should not prohibit the client's ability to find and verify a chain.

## Invalid Edges

To verify that clients support _robust_ path building, test cases can mark an edge in the Trust Graph as _invalid_.
The test suite supports several reasons for an edge being invalid (see the next section) such as the certificate being expired.
When there are multiple paths to a trust anchor, a robust client should be able to find a path to the trust anchor despite invalid edges.
Put another way, clients should not just find any path and then verify it; they must check that each edge is valid as they build a path.

Consider the following trust graph copied from [RFC 4158 figure 7](https://datatracker.ietf.org/doc/html/rfc4158#section-2.3):

```
     +---------+
     |  Trust  |
     | Anchor  |
     +---------+
      |       |
      v       v
   +---+    +---+
   | A |<-->| C |
   +---+    +---+
    |         |
    |  +---+  |
    +->| B |<-+
       +---+
         |
         v
       +----+
       | EE |
       +----+
```

If the certificate `Trust Anchor => A` is expired a client should still be able to find a path from `Trust Anchor` down to `EE` (and similarly if the `Trust Achor => C` certificate is expired).

## Invalid Edge Reasons

This test suite supports several reasons for an edge being invalid as described below.
Some of these reasons MUST be supported by any TLS implementation, such as a certificate being expired.
Other reasons are not strictly required, such as the certificate using a deprecated signature algorithm (i.e. using SHA-1).

### EXPIRED

The certificate's NotAfter time is before now. All TLS implementations must support this.

### NAME_CONSTRAINTS

The certificate has a [name constraints](https://nameconstraints.bettertls.com/) extension that is at odds with the end-entity certificate.
This extension is marked as critical by this test suite, so all implementations should support this.

### BAD_EKU

The certificate has an [Extended Key Usage](https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.12) extension that is incompatible with the end-entity's use of the certificate for TLS server authentication.
The [Mozilla Certificate Policy FAQ](https://wiki.mozilla.org/CA:CertificatePolicyV2.1#Frequently_Asked_Questions) states:

> Inclusion of EKU in CA certificates is generally allowed. NSS and CryptoAPI both treat the EKU extension in intermediate certificates as a constraint on the permitted EKU OIDs in end-entity certificates. Browsers and certificate client software have been using EKU in intermediate certificates, and it has been common for enterprise subordinate CAs in Windows environments to use EKU in their intermediate certificates to constrain certificate issuance.

So while most implementations support the semantics of an incompatible EKU in CAs as a reason to treat a certificate as invalid, RFCs do not require this behavior so this might not be supported by all TLS implementations.

### MISSING_BASIC_CONSTRAINTS

The certificate is missing the [Basic Constraints](https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.9) extension.
RFC 5280 requires that all CAs have this extension, so all implementations should support this.

### NOT_A_CA

The certificate has a basic constraints extension, but it does not identify the certificate as being a CA.
Similarly to the above, all implementations should support this.

### DEPRECATED_CRYPTO

The certificate is signed with an algorithm that has been considered deprecated (i.e. using SHA-1).
Enforcement of SHA-1 deprecation is not universally present in all TLS implementations.

# Interpreting test results

The [Go test executor](test_executor.go) (and javascript test executor) first evaluate whether a given TLS implementation supports branching certificate chains at all (using the TWO_ROOTS trust graph described above).
If it doesn't, most tests in the suite will be skipped.
As noted above, according to RFC 8446 any TLS implementation supporting TLS 1.3 _SHOULD_ support branching certificate chains.

The test executor also tests whether each Invalid Reason described above is supported by the TLS implementation.
It is reasonable for some of them (such as BAD_EKU and DEPRECATED_CRYPTO) to not be supported and subsequent test cases that use those Invalid Reasons will be skipped.

Thus, any skipped tests indicates that the TLS implementation does not support a particular behavior.
Any _failed_ tests are likely to indicate that the TLS implementation does not correctly implement a feature it _intends_ to support.
