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

package com.bettertls.nameconstraints;

import org.bouncycastle.asn1.x509.GeneralName;
import org.bouncycastle.asn1.x509.GeneralNames;
import org.bouncycastle.asn1.x509.GeneralSubtree;
import org.bouncycastle.asn1.x509.NameConstraints;
import org.bouncycastle.openssl.jcajce.JcaPEMWriter;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.IOException;
import java.io.OutputStream;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.UnrecoverableEntryException;
import java.security.cert.Certificate;
import java.security.cert.CertificateEncodingException;
import java.util.ArrayList;
import java.util.List;

public class CertificateGenerator {

    public static void main(String[] args) throws Exception {

        final Path outputDir = Paths.get("../certificates");
        if (!Files.exists(outputDir)) {
            Files.createDirectory(outputDir);
        }

        final JSONObject config = new JSONObject(new String(Files.readAllBytes(Paths.get("../config.json")), StandardCharsets.UTF_8));

        new CertificateGenerator(config, outputDir).generateCertificates();
    }

    private final Path outputDir;
    private final String hostname;
    private final String ip;
    private final String hostSubtree;
    private final String ipSubtree;
    private final String invalidHostname;
    private final String invalidIp;
    private final String invalidHostSubtree;
    private final String invalidIpSubtree;

    private final JSONArray certManifest = new JSONArray();
    private int nextCertId = 1;

    private CertificateGenerator(JSONObject config, Path outputDir) {
        this.outputDir = outputDir;

        this.hostname = config.getString("hostname");
        this.ip = config.getString("ip");
        this.hostSubtree = config.getString("hostSubtree");
        this.ipSubtree = config.getString("ipSubtree");

        this.invalidHostname = config.getString("invalidHostname");
        this.invalidIp = config.getString("invalidIp");
        this.invalidHostSubtree = config.getString("invalidHostSubtree");
        this.invalidIpSubtree = config.getString("invalidIpSubtree");
    }

    private void generateCertificates() throws Exception {

        KeyStore rootCa = new KeyStoreGenerator()
                .setCaKeyEntry(null)
                .setCommonName("Name Constraints Test Root CA")
                .setIsCa(true)
                .setNameConstraints(null)
                .setSubjectAlternateNames(null)
                .build();

        writeCertificate(rootCa.getCertificate(KeyStoreGenerator.DEFAULT_ALIAS), outputDir.resolve("root.crt"));

        for (String commonName : new String[]{null, hostname, ip, invalidHostname, invalidIp }) {
            for (String dnsSan : new String[]{null, hostname, invalidHostname}) {
                for (String ipSan : new String[]{null, ip, invalidIp}) {
                    generateCertificatesWithNames(rootCa, commonName, dnsSan, ipSan);
                }
            }
        }

        final JSONObject manifest = new JSONObject();
        manifest.put("certManifest", certManifest);
        Files.write(outputDir.resolve("manifest.json"), manifest.toString().getBytes(StandardCharsets.UTF_8));
    }

    private void generateCertificatesWithNames(KeyStore rootCa, String commonName, String dnsSan, String ipSan) throws Exception {

        GeneralNames sans = null;
        if (dnsSan != null || ipSan != null) {
            List<GeneralName> generalNames = new ArrayList<>();
            if (dnsSan != null) {
                generalNames.add(new GeneralName(GeneralName.dNSName, dnsSan));
            }
            if (ipSan != null) {
                generalNames.add(new GeneralName(GeneralName.iPAddress, ipSan));
            }
            sans = new GeneralNames(generalNames.toArray(new GeneralName[generalNames.size()]));
        }

        for (String ncIpWhitelist : new String[] { null, ipSubtree, invalidIpSubtree }) {
            for (String ncDnsWhitelist : new String[] { null, hostSubtree, invalidHostSubtree }) {

                List<GeneralSubtree> permittedWhitelist = new ArrayList<>();
                if (ncIpWhitelist != null) {
                    permittedWhitelist.add(new GeneralSubtree(new GeneralName(GeneralName.iPAddress, ncIpWhitelist)));
                }
                if (ncDnsWhitelist != null) {
                    permittedWhitelist.add(new GeneralSubtree(new GeneralName(GeneralName.dNSName, ncDnsWhitelist)));
                }

                for (String ncIpBlacklist : new String[] { null, ipSubtree, invalidIpSubtree }) {
                    for (String ncDnsBlacklist : new String[]{null, hostSubtree, invalidHostSubtree }) {

                        List<GeneralSubtree> permittedBlacklist = new ArrayList<>();
                        if (ncIpBlacklist != null) {
                            permittedBlacklist.add(new GeneralSubtree(new GeneralName(GeneralName.iPAddress, ncIpBlacklist)));
                        }
                        if (ncDnsBlacklist != null) {
                            permittedBlacklist.add(new GeneralSubtree(new GeneralName(GeneralName.dNSName, ncDnsBlacklist)));
                        }

                        NameConstraints nameConstraints = null;
                        if (permittedWhitelist.size() != 0 || permittedBlacklist.size() != 0) {
                            nameConstraints = new NameConstraints(
                                    permittedWhitelist.size() == 0 ? null : permittedWhitelist.toArray(new GeneralSubtree[permittedWhitelist.size()]),
                                    permittedBlacklist.size() == 0 ? null : permittedBlacklist.toArray(new GeneralSubtree[permittedBlacklist.size()]));
                        }

                        System.out.println("Generating certificate " + nextCertId + "...");
                        writeCertificateSet(makeTree(nextCertId, rootCa, nameConstraints, commonName, sans), outputDir, Integer.toString(nextCertId));

                        // Build a manifest JSON entry for the certificate
                        JSONArray manifestSans = new JSONArray();
                        if (dnsSan != null) {
                            manifestSans.put(dnsSan);
                        }
                        if (ipSan != null) {
                            manifestSans.put(ipSan);
                        }
                        JSONObject manifestNcs = new JSONObject();
                        JSONArray manifestNcWhitelist = new JSONArray();
                        if (ncDnsWhitelist != null) {
                            manifestNcWhitelist.put(ncDnsWhitelist);
                        }
                        if (ncIpWhitelist != null) {
                            manifestNcWhitelist.put(ncIpWhitelist);
                        }
                        JSONArray manifestNcBlacklist = new JSONArray();
                        if (ncDnsBlacklist != null) {
                            manifestNcBlacklist.put(ncDnsBlacklist);
                        }
                        if (ncIpBlacklist != null) {
                            manifestNcBlacklist.put(ncIpBlacklist);
                        }
                        manifestNcs.put("whitelist", manifestNcWhitelist);
                        manifestNcs.put("blacklist", manifestNcBlacklist);

                        certManifest.put(new JSONObject()
                                .put("id", nextCertId)
                                .put("commonName", commonName)
                                .put("sans", manifestSans)
                                .put("nameConstraints", manifestNcs)
                        );

                        nextCertId += 1;
                    }
                }
            }
        }
    }

    private static KeyStore makeTree(int certId, KeyStore rootCa, NameConstraints nameConstraints, String leafCommonName, GeneralNames leafSubjectAlternateNames) throws Exception {
        KeyStore localRoot = new KeyStoreGenerator()
                .setCaKeyEntry(getSignerPrivateKey(rootCa))
                .setCommonName("Local Root for " + certId)
                .setIsCa(true)
                .setNameConstraints(nameConstraints)
                .build();
        KeyStore localIntermediate = new KeyStoreGenerator()
                .setCaKeyEntry(getSignerPrivateKey(localRoot))
                .setCommonName("Intermediate CA for " + certId)
                .setIsCa(true)
                .build();
        KeyStore leafCert = new KeyStoreGenerator()
                .setCaKeyEntry(getSignerPrivateKey(localIntermediate))
                .setCommonName(leafCommonName)
                .setIsCa(false)
                .setSubjectAlternateNames(leafSubjectAlternateNames)
                .build();

        return leafCert;
    }

    private static void writeCertificate(Certificate certificate, Path path) throws CertificateEncodingException, IOException {
        try (OutputStream stream = Files.newOutputStream(path);
             OutputStreamWriter writer = new OutputStreamWriter(stream);
             JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            pemWriter.writeObject(certificate);
        }
    }

    private static void writeCertificateSet(KeyStore keyStore, Path outputDir, String name) throws IOException, CertificateEncodingException, UnrecoverableEntryException, NoSuchAlgorithmException, KeyStoreException {
        KeyStore.PrivateKeyEntry keyEntry = (KeyStore.PrivateKeyEntry) keyStore.getEntry(KeyStoreGenerator.DEFAULT_ALIAS, new KeyStore.PasswordProtection(KeyStoreGenerator.KEYSTORE_PASSWORD.toCharArray()));

        try (OutputStream stream = Files.newOutputStream(outputDir.resolve(name + ".key"));
             OutputStreamWriter writer = new OutputStreamWriter(stream);
             JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            pemWriter.writeObject(keyEntry.getPrivateKey());
        }

        try (OutputStream stream = Files.newOutputStream(outputDir.resolve(name + ".crt"));
             OutputStreamWriter writer = new OutputStreamWriter(stream);
             JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            pemWriter.writeObject(keyEntry.getCertificate());
        }

        try (OutputStream stream = Files.newOutputStream(outputDir.resolve(name + ".chain"));
             OutputStreamWriter writer = new OutputStreamWriter(stream);
             JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            Certificate[] chain = keyEntry.getCertificateChain();
            for (int i = 1; i < chain.length; i++) {
                pemWriter.writeObject(chain[i]);
            }
        }
    }

    private static KeyStore.PrivateKeyEntry getSignerPrivateKey(KeyStore keyStore) throws UnrecoverableEntryException, NoSuchAlgorithmException, KeyStoreException {
        return (KeyStore.PrivateKeyEntry) keyStore.getEntry(KeyStoreGenerator.DEFAULT_ALIAS, new KeyStore.PasswordProtection(KeyStoreGenerator.KEYSTORE_PASSWORD.toCharArray()));
    }
}
