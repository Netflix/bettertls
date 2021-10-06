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

function formatDescriptions(data) {
  var desc = "<ul>";
  for (var i = 0; i < data.descriptions.length; i++) {
    var li = document.createElement('li');
    li.appendChild(document.createTextNode(data.descriptions[i]));
    desc += '<li>' + li.innerHTML + '</li>';
  }
  desc += '</ul>';
  return desc;
}

function linkRenderer(data, type, row, meta) {
  var host = (row.type == 'IP' ? sessionData.config.ip : sessionData.config.hostname);
  return "<a href=\"https://" + host + ":" + (sessionData.config.basePort + row.id) + "/nameconstraints/well-known.txt\">" + data + "</a>";
}

function browserTest() {
  if (window.location.hostname.endsWith("bettertls.com")) {
    $('#browserTestsDisabled').show();
    return;
  }

  var config = sessionData.config;
  var testId = 0;
  for (var id in sessionData.testMap) {
    var t = sessionData.testMap[id];
    if (t.commonName == config.hostname
        && t.sans.length == 2 && t.sans.indexOf(config.hostname) != -1 && t.sans.indexOf(config.ip) != -1
        && t.nameConstraints.whitelist.length == 0
        && t.nameConstraints.blacklist.length == 0) {
      testId = sessionData.testMap[id].id;
      break;
    }
  }
  if (testId == 0) {
    alert("Could not find test certificate set to use.");
    return;
  }
  $.get('https://' + sessionData.config.hostname + ':' + (sessionData.config.basePort+testId) + '/nameconstraints/well-known.txt').then(
    function(data) {
      runBrowserTest();
    }, function(err) {
      $('#installRoot').show();
    }
  );
}

function runBrowserTest() {

  var testResults = {
    testVersion: sessionData.config.testVersion,
    date: Date.now(),
    userAgent: navigator.userAgent,
    results: []
  };

  var displayDiv = $('#testBrowserTab .testResults')[0];

  renderLiveTestResults(displayDiv, testResults, function(testId, type) {
    var myPromise = $.Deferred();
    var host = (type == 'DNS' ? sessionData.config.hostname : sessionData.config.ip);
    var targetUrl = 'https://' + host + ':' + (sessionData.config.basePort+testId) + '/nameconstraints/well-known.txt';

    $.get(targetUrl).then(
      function() {
	myPromise.resolve(true);
      }, function() {
	myPromise.resolve(false);
      }
    );

    return myPromise;
  }, function(testResults, table) {
    var a = document.createElement('a');
    a.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(JSON.stringify(testResults)));
    a.setAttribute('download', 'testResults.json');
    var button = document.createElement('button');
    button.appendChild(document.createTextNode("Download Test Results"));
    a.appendChild(button);
    displayDiv.insertBefore(a, displayDiv.firstChild);
  });
}

function buildResultRow(testData, status, expect, type, stats) {
  var testPassed = true;
  var resultText = null;
  if (status) {
    if (expect.expect == 'OK' || expect.expect == 'WEAK-OK') {
      // Pass
    } else {
      testPassed = false;
      resultText = 'False Negative';
    }
  } else {
    if (expect.expect == 'OK') {
      testPassed = false;
      resultText = 'False Positive';
    } else if (expect.expect == 'WEAK-OK') {
      resultText = 'False Positive (OK)';
    } else {
      // Pass
    }
  }

  stats.numRun += 1;
  if (testPassed) {
    stats.numPassed += 1;
  } else {
    stats.numFailed += 1;
  }

  return {
    id: testData.id,
    type: type,
    commonName: testData.commonName == null ? '' : testData.commonName,
    sans: testData.sans.join(', '),
    ncWhitelist: testData.nameConstraints.whitelist.join(', '),
    ncBlacklist: testData.nameConstraints.blacklist.join(', '),
    expect: expect.expect,
    status: status ? 'OK' : 'ERROR',
    testResult: (resultText === null ? 'OK' : resultText),
    descriptions: testData.descriptions.concat(expect.descriptions),
    testPassed: testPassed
  };
}

function refreshStats(testDiv, stats) {
  var target = $('.stats', testDiv)[0];
  target.innerHTML = 'Tests run: ' + stats.numRun + '/' + stats.numTests
    + ', Num Passed: <span class="passed">' + stats.numPassed + '</span>'
    + ', Num Failed: <span class="failed">' + stats.numFailed + '</span>';
}

function generateTestTable(displayDiv, testResults) {
  var testDiv = document.createElement('div');
  testDiv.innerHTML = '<div class="divTable">'
    + '<div><div>User Agent:</div><div class="useragent"></div></div>'
    + '<div class="osVersion"><div>OS Version:</div><div></div></div>'
    + '<div><div>Test Date:</div><div class="testDate"></div></div></div>'
    + '<div class="stats"></div>'
    + '<table class="testResultsTable display"><thead><tr><th></th><th>Id</th><th>Type</th><th>Common Name</th><th>Subject Alternate Names</th><th>Name Constraints Whitelist</th><th>Name Constraints Blacklist</th><th>Expected Status<th>Status</th><th>Test Result</th></tr></thead><tbody></tbody></table>';
  testDiv.querySelector('.useragent').appendChild(document.createTextNode(testResults.userAgent));
  testDiv.querySelector('.testDate').appendChild(document.createTextNode(new Date(testResults.date).toString()));
  if (testResults.osVersion != null) {
    testDiv.querySelector('.osVersion > div:nth-child(2)').appendChild(document.createTextNode(testResults.osVersion));
  } else {
    testDiv.querySelector('.osVersion').style = 'display:none';
  }

  $(displayDiv).empty().append(testDiv);

  var filters = document.createElement('div');
  filters.className = 'filters';
  filters.innerHTML = '<label><input type="checkbox" name="hide_passing" checked> Hide Passing</label>';

  var table = $('table', testDiv).DataTable({
    filters: {
      hidePassing: filters.querySelector('input[name="hide_passing"]')
    },
    columns: [
      {
        "className": 'details-control',
        "orderable": false,
        "data": null,
        "defaultContent": ''
      },
      { data: 'id' },
      { data: 'type' },
      { data: 'commonName' },
      { data: 'sans' },
      { data: 'ncWhitelist' },
      { data: 'ncBlacklist' },
      { data: 'expect', render: linkRenderer },
      { data: 'status' },
      { data: 'testResult' }
    ],
    "order": [[1, 'asc']],
    "lengthMenu": [ [10, 25, 50, 100, 1000, -1], [10, 25, 50, 100, 1000, "All"] ],
    "createdRow": function( tr, data, dataIndex ) {
      if (!data.testPassed) {
        $(tr).addClass('fail');
        if (data.testResult.indexOf('False Positive') !== -1) {
          $(tr).addClass('false-positive');
        }
        if (data.testResult.indexOf('False Negative') !== -1) {
          $(tr).addClass('false-negative');
        }
      }
      $('.details-control', tr).click(function() {
        var row = table.row(tr);
        if ( row.child.isShown() ) {
          // This row is already open - close it
          row.child.hide();
          $(tr).removeClass('shown');
        } else {
          // Open this row
          row.child( formatDescriptions(data) ).show();
          $(tr).addClass('shown');
        }
      });
    }
  });
  $('.dataTables_filter', testDiv).empty().append(filters);
  $('input[type="checkbox"]', filters).change(table.draw);

  var stats = {
    numTests: Object.keys(sessionData.testMap).length * 2,
    numRun: 0,
    numPassed: 0,
    numFailed: 0
  };

  return {
    'testDiv': testDiv,
    'table': table,
    'stats': stats
  };
}

function renderLiveTestResults(displayDiv, testResults, runTestCallback, doneCallback) {

  var testTable = generateTestTable(displayDiv, testResults);
  var testDiv = testTable.testDiv;
  var table = testTable.table;
  var stats = testTable.stats;

  function runTest(id) {
    var t = sessionData.testMap[id];
    if (t == null) {
      if (doneCallback != null) {
        doneCallback(testResults, table);
      }
      return;
    }

    var subtestResult = function(status, expect, type) {
      var resultRow = buildResultRow(t, status, expect, type, stats);

      table.row.add(resultRow);
      table.draw();
    }

    runTestCallback(id, 'DNS').then(function(dnsResult) {
      subtestResult(dnsResult, t.dnsExpect, 'DNS');
      refreshStats(testDiv, stats);

      runTestCallback(id, 'IP').then(function(ipResult) {
        subtestResult(ipResult, t.ipExpect, 'IP');

        testResults.results.push({
          id: t.id,
          dnsResult: dnsResult,
          ipResult: ipResult
        });

        refreshStats(testDiv, stats);
        runTest(id+1);

      }, function(err) { alert("Failed to run test; try refreshing."); });
    }, function(err) { alert("Failed to run test; try refreshing."); });

  }

  runTest(1);
}

function renderSavedTestResults(displayDiv, testResults) {
  var resultsMap = {};
  for (var i = 0; i < testResults.results.length; i++) {
    var result = testResults.results[i];
    resultsMap[result.id] = result;
  }

  var testTable = generateTestTable(displayDiv, testResults);

  var allRows = [];
  for (var id = 1; sessionData.testMap[id] != null; id++) {
    var t = sessionData.testMap[id];
    var result = resultsMap[t.id];
    if (result == null) {
      console.error("Missing saved result for " + t.id);
      continue;
    }
    allRows.push(buildResultRow(t, result.dnsResult, t.dnsExpect, 'DNS', testTable.stats));
    allRows.push(buildResultRow(t, result.ipResult, t.ipExpect, 'IP', testTable.stats));
  }

  refreshStats(testTable.testDiv, testTable.stats);
  testTable.table.rows.add(allRows);
  testTable.table.draw();
}

