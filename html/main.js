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
window.sessionData = {};
$(document).ready(function() {

  // Setup navigation
  function hashNav() {
    var target = window.location.hash;
    if (target.indexOf('?') != -1) {
      target = target.substring(0, target.indexOf('?'));
    }
    if (target == '' || target == '#') {
      target = '#!about';
    }

    $('#mainTabs > .nav > li').removeClass('active');
    $('#mainTabs > div').hide();
    if (target == '#!test') {
      $('#mainTabs > .nav > li[data-target="testBrowserTab"]').addClass('active');
      $('#mainTabs #testBrowserTab').show();
    } else if (target == '#!view') {
      $('#mainTabs > .nav > li[data-target="viewResultsTab"]').addClass('active');
      $('#mainTabs #viewResultsTab').show();

      var params = window.location.hash;
      if (params.indexOf('?') == -1) {
        params = '';
      } else {
        params = params.substring(params.indexOf('?')+1);
      }
      params = params.split('&');
      var paramsMap = {};
      for (var i = 0; i < params.length; i++) {
        var p = params[i].split('=', 2);
        paramsMap[decodeURIComponent(p[0])] = decodeURIComponent(p[1]);
      }

      if (paramsMap['results']) {
        $('#resultsSelect').val(paramsMap['results']);
        showArchiveResults(paramsMap['results']);
      }

    } else if (target == '#!about') {
      $('#mainTabs > .nav > li[data-target="aboutTab"]').addClass('active');
      $('#mainTabs #aboutTab').show();
    }

  }
  window.onhashchange = hashNav;

  $('#resultsSelect').change(function() {
    window.location.hash = '#!view?results=' + encodeURIComponent($(this).val());
  });
  $('#resultsUpload').change(function() {
    var f = this.files[0]; 

    if (f) {
      var r = new FileReader();
      r.onload = function(e) {
        showResults(JSON.parse(e.target.result), $("#viewResultsTab .testResults")[0]);
      };
      r.readAsText(f);
    }
  });

  function showArchiveResults(name) {
    $.getJSON('results/' + name + '.json').then(function(results) {
      var displayDiv = $("#viewResultsTab .testResults");
      showResults(results, displayDiv[0]);
    }, function(err) {
      alert("There was an error fetching the archived test results");
    });
  }

  function showResults(results, displayDiv) {
    var tableDiv = $('.testResultsOutput', displayDiv).empty().hide();
    var loadingDiv = $('.testResultsLoading', displayDiv).show();

    setTimeout(function() {
      renderSavedTestResults(tableDiv[0], results);
      loadingDiv.hide();
      tableDiv.show();
    }, 100);
  }

  $('#initTestButton').click(function() {
    $(this).remove();
    browserTest();
  });

  /* Custom filtering function which will search data in column four between two values */
  $.fn.dataTable.ext.search.push(
    function( settings, data, dataIndex ) {
      var filters = settings.oInit.filters;
      var hidePassing = filters.hidePassing.checked;
      if (hidePassing && (data[9] === 'OK' || data[9] == 'False Positive (OK)')) {
        return false;
      }
      return true;
    }
  );

  var loaderTimeout = setTimeout(function() {
    $('#mainLoader').show();
  }, 500);

  // Pre-load session data
  $.when( $.getJSON('config.json'), $.getJSON('manifest.json'), $.getJSON('expects.json') )
    .then(function(configRaw, manifest, expects) {
      sessionData.config = configRaw[0];
      sessionData.testMap = {};
      var testMap = sessionData.testMap;
      for (var i=0; i < manifest[0].certManifest.length; i++) {
        testMap[manifest[0].certManifest[i].id] = manifest[0].certManifest[i];
      }
      for (var i=0; i < expects[0].expects.length; i++) {
        var expect = expects[0].expects[i]
        testMap[expect.id].descriptions = expect.descriptions;
        testMap[expect.id].ipExpect = expect.ip;
        testMap[expect.id].dnsExpect = expect.dns;
      }

      // Hide loading and show main tab
      clearTimeout(loaderTimeout);
      $('#mainLoader').hide();
      $('#mainTabs').show();

      // Perform the initial navigation
      hashNav();
    }, function(err) {
      alert("Could not load configuration files. Try refreshing.");
    });
});
