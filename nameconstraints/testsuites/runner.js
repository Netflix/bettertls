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
const child_process = require('child_process');

function runSystem(cmd, andThen) {
  child_process.exec(cmd, {}, function(error, stdout, stderr) {
    if (error != null) {
      throw error;
    }
    andThen(stdout);
  });
}

function runTestsWithCmd(userAgent, cmdTemplate, resultsPath) {
  runTests(userAgent, function(dnsUrl, ipUrl, done) {
    child_process.exec(cmdTemplate.replace('${URL}', dnsUrl), {}, function(dnsError, dnsStdout, dnsStderr) {
      child_process.exec(cmdTemplate.replace('${URL}', ipUrl), {}, function(ipError, ipStdout, ipStderr) {
        done([dnsError == null, ipError == null]);
      });
    });
  }, resultsPath);
}

function runTests(userAgent, callback, resultsPath) {
  var config = JSON.parse(fs.readFileSync('../config.json'));
  var manifest = JSON.parse(fs.readFileSync('../../docs/nameconstraints/manifest.json'));
  var maxId = 1;
  for (var i=0; i < manifest.certManifest.length; i++) {
    maxId = Math.max(maxId, manifest.certManifest[i].id);
  }

  var testResults = {
    testVersion: config.testVersion,
    date: Date.now(),
    userAgent: userAgent,
    results: []
  };

  if (fs.existsSync(resultsPath)) {
    testResults = JSON.parse(fs.readFileSync(resultsPath));
  }

  function doTest(testId) {
    if (testId > maxId) {
      process.stdout.write("\n");
      fs.writeFileSync(resultsPath, JSON.stringify(testResults));
      return;
    }
    process.stdout.write("Running test " + testId + "/" + maxId + "\r");

    var testAlreadyRun = false;
    for (var i = 0; i < testResults.results.length; i++) {
      if (testResults.results[i].id == testId) {
        testAlreadyRun = true;
        break;
      }
    }
    if (testAlreadyRun) {
      doTest(testId+1);
    } else {
      var dnsUrl = 'https://' + config.hostname + ':' + (config.basePort+testId) + '/nameconstraints/well-known.txt';
      var ipUrl = 'https://' + config.ip + ':' + (config.basePort+testId) + '/nameconstraints/well-known.txt';

      callback(dnsUrl, ipUrl, function(result) {
        testResults.results.push({
          id: testId,
          dnsResult: result[0],
          ipResult: result[1]
        });
        fs.writeFileSync(resultsPath, JSON.stringify(testResults));
        doTest(testId+1);
      });
    }
  }

  runSystem('lsb_release -d -s', function(osVersion) {
    testResults.osVersion = osVersion;
    doTest(1);
  });
}

exports.runSystem = runSystem;
exports.runTestsWithCmd = runTestsWithCmd;
exports.runTests = runTests;

