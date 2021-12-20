(function() {
    var TestCaseResult = protobuf.roots.default.test_executor.TestCaseResult;
    var SuiteTestResults = protobuf.roots.default.test_executor.SuiteTestResults;
    var content = document.getElementById('view_test_results');

    function base64decode(s) {
        var raw = window.atob(s);
        var rawLength = raw.length;
        var array = new Uint8Array(new ArrayBuffer(rawLength));
        var i;
        for(i = 0; i < rawLength; i++) {
            array[i] = raw.charCodeAt(i);
        }
        return array;
    }

    function td(txt) {
        var td = document.createElement('td');
        if (txt) {
            td.appendChild(document.createTextNode(txt));
        }
        return td;
    }

    function tdRowSpan(txt, rowSpan) {
        var td = document.createElement('td');
        td.rowSpan = rowSpan;
        td.appendChild(document.createTextNode(txt));
        return td;
    }

    var loadedManifests = {};
    var loadedSuiteResults = {};
    var currentManifest;

    function loadTestResults(resultsPath) {
        fetch(resultsPath).then(function(res) {
            return res.json()
        }).then(function (results) {
            content.getElementsByClassName('implementation')[0].textContent = results.implementation;
            content.getElementsByClassName('version')[0].textContent = results.version;
            content.getElementsByClassName('date')[0].textContent = results.date;

            loadManifest(results.betterTlsRevision).then(function() {
                currentManifest = loadedManifests[results.betterTlsRevision];
                for (suiteName in results.suites) {
                    var rawProto = pako.inflate(base64decode(results.suites[suiteName]));
                    loadedSuiteResults[suiteName] = SuiteTestResults.decode(rawProto);
                }
            }).then(function() {
                renderLoadedResults();
            })
        });
    }

    function loadManifest(revision) {
        if (loadedManifests[revision]) {
            return Promise.resolve();
        }
        return fetch('results/manifests/' + revision + '.manifest')
            .then(function(res) { return res.json()})
            .then(function(manifest) {
                loadedManifests[revision] = manifest;
            });
    }

    function appendRow() {
        var tbody = arguments[0];
        var tr = document.createElement('tr');
        for (var i = 1; i < arguments.length; i++) {
            tr.appendChild(arguments[i]);
        }
        tbody.appendChild(tr);
    }

    function setChild(parent, child) {
        parent.textContent = '';
        parent.appendChild(child);
    }

    function renderLoadedResults() {
        var tableData = [];


        var supportedFeaturesTable = document.createElement('table');
        setChild(content.getElementsByClassName('supportedFeatures')[0], supportedFeaturesTable);
        supportedFeaturesTable.className = 'ui definition table';
        var supportedFeaturesTbody = document.createElement('tbody');
        supportedFeaturesTable.appendChild(supportedFeaturesTbody);
        var unsupportedFeaturesTable = document.createElement('table');
        setChild(content.getElementsByClassName('unsupportedFeatures')[0], unsupportedFeaturesTable);
        unsupportedFeaturesTable.className = 'ui definition table';
        unsupportedFeaturesTable.textContent = '';
        var unsupportedFeaturesTbody = document.createElement('tbody');
        unsupportedFeaturesTable.appendChild(unsupportedFeaturesTbody);

        var testResultSummaryTbody = content.getElementsByClassName('test_result_summary')[0];
        testResultSummaryTbody.textContent = '';

        var suiteName;
        for (suiteName in loadedSuiteResults) {
            var results = loadedSuiteResults[suiteName];
            var manifest = currentManifest.suiteManifests[suiteName];

            var feature;
            var supportedFeatures = [];
            var unsupportedFeatures = [];
            for (feature of results.supportedFeatures) {
                supportedFeatures.push(manifest.features[feature]);
            }
            for (feature of results.unsupportedFeatures) {
                unsupportedFeatures.push(manifest.features[feature]);
            }
            appendRow(supportedFeaturesTbody, td(suiteName), td(supportedFeatures.join(', ')));
            appendRow(unsupportedFeaturesTbody, td(suiteName), td(unsupportedFeatures.join(', ')));

            var summary = {
                'PASS': 0,
                'WARN': 0,
                'SKIP': 0,
                'FAIL_FP': 0,
                'FAIL_FN': 0,
            }

            var testId;
            for (testId = 0; testId < results.testCaseResults.length; testId++) {
                var testResult = results.testCaseResults[testId];
                var expectedResult = manifest.expectedResults[testId];

                var actual;
                var tResult;
                if (testResult === TestCaseResult.SKIPPED) {
                    actual = 'SKIPPED';
                    tResult = 'SKIP';
                } else if (testResult === TestCaseResult.ACCEPTED) {
                    actual = 'ACCEPTED';
                    if (expectedResult === 'FAIL') {
                        tResult = 'FAIL_FN';
                    } else if (expectedResult === 'PASS' || expectedResult === 'SOFT_PASS') {
                        tResult = 'PASS';
                    } else {
                        tResult = 'WARN';
                    }
                } else if (testResult === TestCaseResult.REJECTED) {
                    actual = 'REJECTED';
                    if (expectedResult === 'PASS') {
                        tResult = 'FAIL_FP';
                    } else if (expectedResult === 'FAIL' || expectedResult === 'SOFT_FAIL') {
                        tResult = 'PASS';
                    } else {
                        tResult = 'WARN';
                    }
                }

                summary[tResult] += 1;
                tableData.push([suiteName, testId, expectedResult, actual, tResult]);
            }

            appendRow(testResultSummaryTbody, tdRowSpan(suiteName, 5), td('PASSED'), td('' + (summary['PASS'] + summary['WARN'])));
            appendRow(testResultSummaryTbody, td('PASSED (with warning)'), td('' + summary['WARN']));
            appendRow(testResultSummaryTbody, td('SKIPPED'), td('' + summary['SKIP']));
            appendRow(testResultSummaryTbody, td('FAILED (false positive)'), td('' + summary['FAIL_FP']));
            appendRow(testResultSummaryTbody, td('FAILED (false negative)'), td('' + summary['FAIL_FN']));
        }

        theDataTable.clear();
        theDataTable.rows.add(tableData).draw();
    }

    $('#view_test_results .results_select').dropdown({
        onChange: function(val) {
            if (val) {
                loadTestResults('results/' + val + '.json');
            }
        }
    });

    var theDataTable;
    $(document).ready(function() {
        theDataTable = $('#view_test_results .test_results').DataTable({
            columns: [
                {title: 'Suite'},
                {title: 'Test ID'},
                {title: 'Expected'},
                {title: 'Actual'},
                {title: 'Test Status'}
            ]
        });
    });
})();
