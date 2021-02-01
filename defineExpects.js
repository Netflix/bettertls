#!/usr/bin/env node

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

const fs = require('fs');

const PASS = 0,
  WEAK_PASS = 1,
  FAIL = 2;

var config = JSON.parse(fs.readFileSync('config.json'));
var manifest = JSON.parse(fs.readFileSync('certificates/manifest.json'));
var expects = [];
for (var i=0; i < manifest.certManifest.length; i++) {
  var certDef = manifest.certManifest[i];

  var descriptions = [];

  var ncIpStatus = PASS;
  var ncDnsStatus = PASS;
  if (certDef.sans.length == 0) {
    if ((certDef.commonName == config.ip && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)
        || (certDef.commonName == config.ip && certDef.nameConstraints.blacklist.indexOf(config.ipSubtree) != -1)
        || (certDef.commonName == config.invalidIp && certDef.nameConstraints.whitelist.indexOf(config.ipSubtree) != -1)
        || (certDef.commonName == config.invalidIp && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)) {
      descriptions.push("The IP in the common name violates a name constraint.");
      ncIpStatus = FAIL;
    }
    if ((certDef.commonName == config.hostname && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)
        || (certDef.commonName == config.hostname && certDef.nameConstraints.blacklist.indexOf(config.hostSubtree) != -1)
        || (certDef.commonName == config.invalidHostname && certDef.nameConstraints.whitelist.indexOf(config.hostSubtree) != -1)
        || (certDef.commonName == config.invalidHostname && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)) {
      descriptions.push("The DNS name in the common name violates a name constraint.");
      ncDnsStatus = FAIL;
    }
  } else {
    // If a common name is defined, many implementions ignore it in favor of SAN. So any violation on the common name amounts to a weak pass.
    if ((certDef.commonName == config.ip && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)
        || (certDef.commonName == config.ip && certDef.nameConstraints.blacklist.indexOf(config.ipSubtree) != -1)
        || (certDef.commonName == config.invalidIp && certDef.nameConstraints.whitelist.indexOf(config.ipSubtree) != -1)
        || (certDef.commonName == config.invalidIp && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)) {
      descriptions.push("The IP in the common name violates a name constraint. Because there is a SAN extension, this might be ignored.");
      ncIpStatus = WEAK_PASS;
    }
    if ((certDef.commonName == config.hostname && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)
        || (certDef.commonName == config.hostname && certDef.nameConstraints.blacklist.indexOf(config.hostSubtree) != -1)
        || (certDef.commonName == config.invalidHostname && certDef.nameConstraints.whitelist.indexOf(config.hostSubtree) != -1)
        || (certDef.commonName == config.invalidHostname && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)) {
      descriptions.push("The DNS name in the common name violates a name constraint. Because there is a SAN extension, this might be ignored.");
      ncDnsStatus = WEAK_PASS;
    }

    // If a common name is an IP, it may be treated as a DNS name and have DNS name constraints applied against it.
    if ((certDef.commonName == config.ip || certDef.commonName == config.invalidIp)
        && (certDef.nameConstraints.whitelist.indexOf(config.hostSubtree) != -1
          || certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)) {
      descriptions.push("Although the common name is an IP, some implementations may apply DNS name constraints against it and thus fail validation.");
      ncDnsStatus = WEAK_PASS;
    }

    // Hard fail if there is a violation on the SAN values.
    if ((certDef.sans.indexOf(config.ip) != -1 && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)
        || (certDef.sans.indexOf(config.ip) != -1 && certDef.nameConstraints.blacklist.indexOf(config.ipSubtree) != -1)
        || (certDef.sans.indexOf(config.invalidIp) != -1 && certDef.nameConstraints.whitelist.indexOf(config.ipSubtree) != -1)
        || (certDef.sans.indexOf(config.invalidIp) != -1 && certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)) {
      descriptions.push("The IP in the SAN extension violates a name constraint.");
      ncIpStatus = FAIL;
    }
    if ((certDef.sans.indexOf(config.hostname) != -1 && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)
        || (certDef.sans.indexOf(config.hostname) != -1 && certDef.nameConstraints.blacklist.indexOf(config.hostSubtree) != -1)
        || (certDef.sans.indexOf(config.invalidHostname) != -1 && certDef.nameConstraints.whitelist.indexOf(config.hostSubtree) != -1)
        || (certDef.sans.indexOf(config.invalidHostname) != -1 && certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)) {
      descriptions.push("The DNS name in the SAN extension violates a name constraint.");
      ncDnsStatus = FAIL;
    }
  }

  var expect = {
    'ip': {
      'expect': null,
      'descriptions': []
    },
    'dns': {
      'expect': null,
      'descriptions': []
    }
  };
  if (certDef.commonName != config.ip && certDef.sans.indexOf(config.ip) == -1) {
    expect.ip.descriptions.push("The IP used as an origin is not listed in the CN or SAN extension.");
    expect.ip.expect = 'ERROR';
  } else if (ncIpStatus == FAIL) {
    expect.ip.expect = 'ERROR';
  } else {
    // Expect a pass unless one of the below checks weakens the expectation
    expect.ip.expect = 'OK';

    if (ncIpStatus == WEAK_PASS) {
      expect.ip.expect = 'WEAK-OK';
    }
    if (ncDnsStatus != PASS) {
      expect.ip.expect = 'WEAK-OK';
      expect.ip.descriptions.push("Although the DNS name is not the subject name in question, its name constraint violation may still cause this certificate to be rejected.");
    }

    // Weak-pass if the IP is in the CN but not in a SAN. Most browsers support this, but strictly it's against the RFC and some TLS stacks reject it.
    if (certDef.commonName == config.ip && certDef.sans.indexOf(config.ip) == -1) {
      expect.ip.expect = 'WEAK-OK';
      expect.ip.descriptions.push("The IP is only contained in the CN of this certificate, which isn't permitted by RFC but which many implementations support.");
    }

    // Weak-pass if there is a DNS name constraint and no DNS SAN
    if ((certDef.nameConstraints.whitelist.indexOf(config.hostSubtree) != -1
          || certDef.nameConstraints.whitelist.indexOf(config.invalidHostSubtree) != -1)
        && certDef.commonName != config.hostname
        && certDef.commonName != config.invalidHostname
        && certDef.sans.indexOf(config.hostname) == -1
        && certDef.sans.indexOf(config.invalidHostname) == -1) {
      expect.ip.expect = 'WEAK-OK';
      expect.ip.descriptions.push("There is a DNS name constraint but no DNS name in the certificate. This is allowed by the RFC, but some implementations will fail to validate the certificate.");
    }
  }

  if (certDef.commonName != config.hostname && certDef.sans.indexOf(config.hostname) == -1) {
    expect.dns.descriptions.push("The DNS hostname used as an origin is not listed in the CN or SAN extension.");
    expect.dns.expect = 'ERROR';
  } else if (ncDnsStatus == FAIL) {
    expect.dns.expect = 'ERROR';
  } else {
    // Expect a pass unless one of the below checks weakens the expectation
    expect.dns.expect = 'OK';

    if (ncDnsStatus == WEAK_PASS) {
      expect.dns.expect = 'WEAK-OK';
    }
    if (ncIpStatus != PASS) {
      expect.dns.expect = 'WEAK-OK';
      expect.dns.descriptions.push("Althought the IP address is not the subject name in question, its name constraint violation may still cause this certificate to be rejected.");
    }

    if (certDef.commonName == config.hostname && certDef.sans.indexOf(config.hostname) == -1) {
      expect.dns.expect = 'WEAK-OK';
      expect.dns.descriptions.push("The DNS name for this certificate only exists in the common name. Some browsers (such as Chrome) have deprecated using the CN entirely and only use names from SAN extensions.");
    }

    if (certDef.commonName == config.hostname && certDef.sans.length > 0 && certDef.sans.indexOf(config.hostname) == -1) {
      expect.dns.expect = 'WEAK-OK';
      expect.dns.descriptions.push("The DNS name for this certificate exists in the common name but not in the Subject Alternate Names extension even though the extension is specified. Most implementations will fail DNS-hostname validation on this certificate.");
    }

    // Weak-pass if there is a IP name constraint and no IP SAN
    if ((certDef.nameConstraints.whitelist.indexOf(config.ipSubtree) != -1
          || certDef.nameConstraints.whitelist.indexOf(config.invalidIpSubtree) != -1)
        && certDef.commonName != config.ip
        && certDef.commonName != config.invalidIp
        && certDef.sans.indexOf(config.ip) == -1
        && certDef.sans.indexOf(config.invalidIp) == -1) {
      expect.dns.expect = 'WEAK-OK';
      expect.dns.descriptions.push("There is a IP name constraint but no IP in the certificate. This isn't an explicit violation, but some implementations will fail to validate the certificate.");
    }
  }

  expects.push({
    'id': certDef.id,
    'ip': expect.ip,
    'dns': expect.dns,
    'descriptions': descriptions
  });
}

fs.writeFileSync('html/expects.json', JSON.stringify({'expects': expects}));
