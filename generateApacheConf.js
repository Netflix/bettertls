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

var config = JSON.parse(fs.readFileSync('config.json'));
var manifest = JSON.parse(fs.readFileSync('certificates/manifest.json'));
var maxId = 1;
for (var i=0; i < manifest.certManifest.length; i++) {
  maxId = Math.max(maxId, manifest.certManifest[i].id);
}

process.stdout.write("ServerName " + config.hostname + "\n");
process.stdout.write("Listen " + config.basePort + "\n"
    + "<VirtualHost *:" + config.basePort + ">\n"
    + "  DocumentRoot /apps/bettertls/test_html\n"
    + "</VirtualHost>\n");

for (var i=1; i<=maxId; i++) {
  var port = config.basePort + i;
  process.stdout.write("Listen " + port + "\n"
      + "<VirtualHost *:" + port + ">\n"
      + "  DocumentRoot /apps/bettertls/test_html\n"
      + "  Header set Access-Control-Allow-Origin \"*\"\n"
      + "  SSLEngine on\n"
      + "  SSLCertificateFile /apps/bettertls/certificates/" + i + ".crt\n"
      + "  SSLCertificateKeyFile /apps/bettertls/certificates/" + i + ".key\n"
      + "  SSLCertificateChainFile /apps/bettertls/certificates/" + i + ".chain\n"
      + "</VirtualHost>\n");
}

