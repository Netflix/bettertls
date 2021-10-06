package impltests

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestJava(t *testing.T) {
	version, err := execAndCapture("java", "-version")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(version)

	tmpDir := t.TempDir()
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
        String servername = args[1];
        String url = args[2];

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
        connection.setSSLSocketFactory(new SSLSocketFactoryWrapper(sslContext.getSocketFactory(), servername));
        connection.setHostnameVerifier((hostname, session) -> true);
        if (connection.getResponseCode() != 200) {
            throw new AssertionError("Invalid response code: " + connection.getResponseCode());
        }
    }

    private static class SSLSocketFactoryWrapper extends SSLSocketFactory {

        private final SSLSocketFactory wrappedFactory;
        private final String servername;

        public SSLSocketFactoryWrapper(SSLSocketFactory factory, String servername) {
            this.wrappedFactory = factory;
            this.servername = servername;
        }

        @Override
        public Socket createSocket(String host, int port) throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket(host, port);
            setParameters(socket);
            return socket;
        }

        @Override
        public Socket createSocket(String host, int port, InetAddress localHost, int localPort) throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket(host, port, localHost, localPort);
            setParameters(socket);
            return socket;
        }


        @Override
        public Socket createSocket(InetAddress host, int port) throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket(host, port);
            setParameters(socket);
            return socket;
        }

        @Override
        public Socket createSocket(InetAddress address, int port, InetAddress localAddress, int localPort) throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket(address, port, localAddress, localPort);
            setParameters(socket);
            return socket;

        }

        @Override
        public Socket createSocket() throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket();
            setParameters(socket);
            return socket;
        }

        @Override
        public String[] getDefaultCipherSuites() {
            return wrappedFactory.getDefaultCipherSuites();
        }

        @Override
        public String[] getSupportedCipherSuites() {
            return wrappedFactory.getSupportedCipherSuites();
        }

        @Override
        public Socket createSocket(Socket s, String host, int port, boolean autoClose) throws IOException {
            SSLSocket socket = (SSLSocket) wrappedFactory.createSocket(s, host, port, autoClose);
            setParameters(socket);
            return socket;
        }

        private void setParameters(SSLSocket socket) {
            SSLParameters parameters = socket.getSSLParameters();
            parameters.setServerNames(Collections.singletonList(new SNIHostName(servername)));
            socket.setSSLParameters(parameters);
        }
    }
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("javac", "Curl.java")
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	testExec(t, func(caPath string, testCaseName string, tlsPort int) []string {
		return []string{
			"java", "-Djdk.tls.maxCertificateChainLength=50", "-cp", tmpDir, "Curl", caPath, testCaseName + ".localhost", fmt.Sprintf("https://localhost:%d/ok", tlsPort),
		}
	})
}
