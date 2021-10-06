all: reports/boringssl_20211005.html reports/curl_7.68.0.html reports/gnutls_3.6.13.html reports/golang_1.17.1.html reports/java_11.0.11.html reports/libressl_3.4.0.html reports/openssl_1.1.1f.html reports/pkijs_2.1.97.html reports/rustls_0.20.0.html

go_test_report:
	go build github.com/vakenbolt/go-test-report

reports/boringssl_20211005.html: go_test_report
	go test -json ./ -run 'TestBoringSSL' | ./go-test-report -t "BoringSSL 20211005" -o reports/boringssl_20211005.html

reports/curl_7.68.0.html: go_test_report
	go test -json ./ -run 'TestCurl' | ./go-test-report -t "Curl 7.68.0" -o reports/curl_7.68.0.html

reports/gnutls_3.6.13.html: go_test_report
	go test -json ./ -run 'TestGnuTls' | ./go-test-report -t "GNU TLS 3.6.13" -o reports/gnutls_3.6.13.html

reports/golang_1.17.1.html: go_test_report
	go test -json ./ -run 'TestX509Verify|TestHttpClient' | ./go-test-report -t "Golang 1.17.1" -o reports/golang_1.17.1.html

reports/java_11.0.11.html: go_test_report
	go test -json ./ -run 'TestJava' | ./go-test-report -t "Java 11.0.11" -o reports/java_11.0.11.html

reports/libressl_3.4.0.html: go_test_report
	go test -json ./ -run 'TestLibreSSL' | ./go-test-report -t "LibreSSL 3.4.0" -o reports/libressl_3.4.0.html

reports/openssl_1.1.1f.html: go_test_report
	go test -json ./ -run 'TestOpenSSL' | ./go-test-report -t "OpenSSL 1.1.1f" -o reports/openssl_1.1.1f.html

reports/pkijs_2.1.97.html: go_test_report
	go test -json ./ -run 'TestPkiJs' | ./go-test-report -t "PKI.js 2.1.97" -o reports/pkijs_2.1.97.html

reports/rustls_0.20.0.html: go_test_report
	go test -json ./ -run 'TestRustls' | ./go-test-report -t "Rustls 0.20.0" -o reports/rustls_0.20.0.html

clean:
	rm -f reports/*.html
