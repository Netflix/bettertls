package impltests

import (
	"fmt"
	test_executor "github.com/Netflix/bettertls/test-suites/test-executor"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type JavaRunner struct {
	tmpDir  string
	version string
}

func (j *JavaRunner) Name() string {
	return "java"
}

func (j *JavaRunner) Initialize() error {
	version, err := execAndCapture("java", "-version")
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("failed to create a temporary directory: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "Curl.java"), []byte(`
import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SNIHostName;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLParameters;
import javax.net.ssl.SSLSocket;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManagerFactory;
import java.io.IOException;
import java.io.InputStream;
import java.net.InetAddress;
import java.net.Socket;
import java.net.URL;
import java.net.UnknownHostException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyStore;
import java.security.cert.Certificate;
import java.security.cert.CertificateFactory;
import java.util.Collections;

public class Curl {
    public static void main(String[] args) throws Exception {
        String caPath = args[0];
        String url = args[1];

        Certificate rootCert;
        try (InputStream inputStream = Files.newInputStream(Paths.get(caPath))) {
            rootCert = CertificateFactory.getInstance("X509").generateCertificate(inputStream);
        }

        KeyStore truststore = KeyStore.getInstance(KeyStore.getDefaultType());
        truststore.load(null, null);
        truststore.setCertificateEntry("1", rootCert);
        TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        tmf.init(truststore);

        SSLContext sslContext = SSLContext.getInstance("TLS");
        sslContext.init(null, tmf.getTrustManagers(), null);

        HttpsURLConnection connection = (HttpsURLConnection) new URL(url).openConnection();
        connection.setSSLSocketFactory(sslContext.getSocketFactory());
        if (connection.getResponseCode() != 200) {
            throw new AssertionError("Invalid response code: " + connection.getResponseCode());
        }
    }
}
`), 0644)
	if err != nil {
		return err
	}
	cmd := exec.Command("javac", "Curl.java")
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return err
	}

	j.tmpDir = tmpDir
	j.version = version
	return nil
}

func (j *JavaRunner) Close() error {
	if j.tmpDir != "" {
		return os.RemoveAll(j.tmpDir)
	}
	return nil
}

func (j *JavaRunner) GetVersion() string {
	return j.version
}

func (j *JavaRunner) RunTests(ctx *test_executor.ExecutionContext) (map[string]*test_executor.SuiteTestResults, error) {
	return testExec(ctx, func(caPath string, hostname string, tlsPort uint) []string {
		return []string{
			"java", "-Djdk.tls.maxCertificateChainLength=50", "-cp", j.tmpDir, "Curl", caPath, fmt.Sprintf("https://%s:%d/ok", hostname, tlsPort),
		}
	})
}
