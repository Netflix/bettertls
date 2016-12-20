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

import org.bouncycastle.asn1.x500.X500Name;
import org.bouncycastle.asn1.x509.*;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.X509v3CertificateBuilder;
import org.bouncycastle.operator.jcajce.JcaContentSignerBuilder;

import java.io.ByteArrayInputStream;
import java.math.BigInteger;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.KeyStore;
import java.security.cert.CertificateFactory;
import java.util.Calendar;
import java.util.Date;

class KeyStoreGenerator {

    public static final String DEFAULT_ALIAS = "1";
    public static final String KEYSTORE_PASSWORD = "changeit";

    private KeyStore.PrivateKeyEntry caKeyEntry;
    private String commonName;
    private boolean isCa;
    private NameConstraints nameConstraints;
    private GeneralNames sans;

    public KeyStoreGenerator setCaKeyEntry(KeyStore.PrivateKeyEntry caKeyEntry) {
        this.caKeyEntry = caKeyEntry;
        return this;
    }

    public KeyStoreGenerator setCommonName(String commonName) {
        this.commonName = commonName;
        return this;
    }

    public KeyStoreGenerator setIsCa(boolean isCa) {
        this.isCa = isCa;
        return this;
    }

    public KeyStoreGenerator setNameConstraints(NameConstraints nameConstraints) {
        this.nameConstraints = nameConstraints;
        return this;
    }

    public KeyStoreGenerator setSubjectAlternateNames(GeneralNames sans) {
        this.sans = sans;
        return this;
    }

    public KeyStore build() throws Exception {
        KeyPairGenerator rsa = KeyPairGenerator.getInstance("RSA");
        rsa.initialize(2048);
        KeyPair kp = rsa.generateKeyPair();

        X509CertificateHolder caCertHolder;
        if (caKeyEntry != null) {
            caCertHolder = new X509CertificateHolder(caKeyEntry.getCertificate().getEncoded());
        } else {
            caCertHolder = null;
        }

        Calendar cal = Calendar.getInstance();
        cal.add(Calendar.MONTH, 12);
        if (caCertHolder != null && cal.getTime().after(caCertHolder.getNotAfter())) {
            cal.setTime(caCertHolder.getNotAfter());
        }

        byte[] pk = kp.getPublic().getEncoded();
        SubjectPublicKeyInfo bcPk = SubjectPublicKeyInfo.getInstance(pk);

        String subjectNameStr = "C=US, ST=California, L=Los Gatos, O=Netflix Inc, OU=Platform Security (" + System.nanoTime() + ")";
        if (commonName != null) {
            subjectNameStr += ", CN=" + commonName;
        }
        X500Name subjectName = new X500Name(subjectNameStr);
        X509v3CertificateBuilder certGen = new X509v3CertificateBuilder(
                caCertHolder == null ? subjectName : caCertHolder.getSubject(),
                BigInteger.valueOf(System.nanoTime()),
                new Date(),
                cal.getTime(),
                subjectName,
                bcPk
        );
        certGen.addExtension(Extension.basicConstraints, true, new BasicConstraints(isCa));
        if (nameConstraints != null) {
            certGen.addExtension(Extension.nameConstraints, false, nameConstraints);
        }
        if (sans != null) {
            certGen.addExtension(Extension.subjectAlternativeName, false, sans);
        }

        X509CertificateHolder certHolder = certGen
                .build(new JcaContentSignerBuilder("SHA256withRSA").build(caKeyEntry == null ? kp.getPrivate() : caKeyEntry.getPrivateKey()));

        java.security.cert.Certificate certificate;
        try (ByteArrayInputStream bais = new ByteArrayInputStream(certHolder.getEncoded())) {
            certificate = CertificateFactory.getInstance("X.509").generateCertificate(bais);
        }

        java.security.cert.Certificate[] certificateChain;
        if (caKeyEntry == null) {
            certificateChain = new java.security.cert.Certificate[]{certificate};
        } else {
            certificateChain = new java.security.cert.Certificate[caKeyEntry.getCertificateChain().length + 1];
            certificateChain[0] = certificate;
            System.arraycopy(caKeyEntry.getCertificateChain(), 0, certificateChain, 1, caKeyEntry.getCertificateChain().length);
        }

        KeyStore keyStore = KeyStore.getInstance(KeyStore.getDefaultType());
        keyStore.load(null, null);
        keyStore.setKeyEntry(DEFAULT_ALIAS, kp.getPrivate(), KEYSTORE_PASSWORD.toCharArray(), certificateChain);
        return keyStore;
    }
}
