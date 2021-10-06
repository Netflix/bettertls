// You can build this file with `vgo build go_http.go`
package main

import (
  "flag"
  "os"
  "io"
  "net/http"
  "fmt"
  "runtime"
  "crypto/x509"
  "crypto/tls"
  "encoding/pem"
  "io/ioutil"
)

func main() {
  version := flag.Bool("version", false, "Display Go version information and exit.")
  cacert := flag.String("cacert", "", "Path to the CA cert to trust.")
  flag.Parse()

  if *version {
    fmt.Printf("%s\n", runtime.Version())
    return
  }

  url := flag.Args()[0]

  var caStore *x509.CertPool
  if *cacert != "" {
    caStore = x509.NewCertPool()

    data, err := ioutil.ReadFile(*cacert)
    if err != nil {
      panic(err)
    }
    for data != nil {
      block, rest := pem.Decode(data)
      if block != nil && block.Type == "CERTIFICATE" {
        cert, err := x509.ParseCertificate(block.Bytes)
        if err != nil {
          panic(err)
        }
        caStore.AddCert(cert)
      }
      if block == nil {
        break
      }
      data = rest
    }
  }

  tr := &http.Transport{
    TLSClientConfig: &tls.Config{
      RootCAs: caStore,
    },
  }
  client := &http.Client{Transport: tr}
  resp, err := client.Get(url)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    panic("Bad status code")
  }
  io.Copy(os.Stdout, resp.Body)
}
