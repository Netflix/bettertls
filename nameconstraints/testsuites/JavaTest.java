/**
 *
 *  Copyright 2017 Netflix, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

import javax.net.ssl.HttpsURLConnection;
import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManagerFactory;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyStore;
import java.security.cert.Certificate;
import java.security.cert.CertificateFactory;

public class JavaTest {

    public static void main(String[] args) throws Exception {
        if (args.length != 2) {
            throw new IllegalArgumentException("Expects two args: dnsUrl ipUrl");
        }
        String dnsUrl = args[0];
        String ipUrl = args[1];

        CertificateFactory certificateFactory = CertificateFactory.getInstance("X509");
        Certificate rootCrt;
        try (InputStream inputStream = Files.newInputStream(Paths.get("../../docs/root.crt"))) {
            rootCrt = certificateFactory.generateCertificate(inputStream);
        }
        KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());
        trustStore.load(null, null);
        trustStore.setCertificateEntry("1", rootCrt);
        TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        tmf.init(trustStore);

        SSLContext sslContext = SSLContext.getInstance("TLS");
        sslContext.init(null, tmf.getTrustManagers(), null);

        boolean dnsResult = testUrl(sslContext, dnsUrl);
        boolean ipResult = testUrl(sslContext, ipUrl);

        System.out.println(String.format("{\"dnsResult\":%s, \"ipResult\":%s}", Boolean.toString(dnsResult), Boolean.toString(ipResult)));
    }

    private static boolean testUrl(SSLContext sslContext, String url) {
        try {
            HttpsURLConnection urlConnection = (HttpsURLConnection) new URL(url).openConnection();
            urlConnection.setSSLSocketFactory(sslContext.getSocketFactory());

            int bytesRead = 0;
            try (InputStream inputStream = urlConnection.getInputStream()) {
                byte[] buffer = new byte[4096];
                bytesRead = inputStream.read(buffer);
            }

            return bytesRead > 0;
        } catch (IOException e) {
            return false;
        }
    }
}
