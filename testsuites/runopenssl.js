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
const https = require('https');
const url = require('url');
const runner = require('./runner.js');

var urlPattern = new RegExp(/^https:\/\/([^:\/]*):([0-9]+)\//);
var verifyPattern = new RegExp(/Verify return code: ([0-9]+)/);

var config = JSON.parse(fs.readFileSync('../config.json'));

runner.runSystem('openssl version', function(version) {
  console.log("UserAgent: " + version);
  runner.runTests(version, function(dnsUrl, ipUrl, done) {
    var matches = dnsUrl.match(urlPattern);
    if (!matches) {
      console.error("Failed to match url: " + dnsUrl);
      return;
    }

    runner.runSystem('openssl s_client -CAfile ../certificates/root.crt -verify_hostname ' + matches[1] + ' -connect ' + matches[1] + ':' + matches[2] + ' </dev/null', function(dnsOutput) {
      var dnsVerifyMatch = dnsOutput.match(verifyPattern);
      if (!dnsVerifyMatch) {
        console.error("Failed to get verify result: " + dnsOutput);
        return;
      }
      var dnsResult = dnsVerifyMatch[1] == "0";

      var matches = ipUrl.match(urlPattern);
      if (!matches) {
        console.error("Failed to match url: " + ipUrl);
        return;
      }

      runner.runSystem('openssl s_client -CAfile ../certificates/root.crt -verify_ip ' + matches[1] + ' -connect ' + matches[1] + ':' + matches[2] + ' </dev/null', function(ipOutput) {
        var ipVerifyMatch = ipOutput.match(verifyPattern);
        if (!ipVerifyMatch) {
          console.error("Failed to get verify result: " + ipOutput);
          return;
        }
        var ipResult = ipVerifyMatch[1] == "0";

        done([dnsResult, ipResult]);
      });
    });
  }, '../docs/results/openssl_1.1.0f_linux.json');
});

