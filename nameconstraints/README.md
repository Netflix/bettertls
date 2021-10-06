Name Constraints
================

Check out [bettertls.com/nameconstraints](https://bettertls.com/nameconstraints) for more information about what the name constraints extension is and what the test suite covers.
This directory contains tools for creating and running the name constraints test suite.

The [config.json](config.json) file defines the hostname and IP used when generating certificates for the test suite and when running the test suite itself. If you intend to run the test suite locally, this is the first thing you should update. For example, to run locally you might setup `localhost.local` to resolve to your localhost and configure `config.json` with

    "ip": "127.0.0.1",
    "ipSubtree": "127.0.0.0/8",
    "hostname": "localhost.local",
    "hostSubtree": "local",

The certificates used for the test suite are generated using the code in the [generator](generator) subfolder. It's built with gradle and can be used with `cd generator; gradle run`. This involves generating a lot of RSA keys, so it can take about an hour to run.

The `defineExpects.js` script generates the `../docs/nameconstraints/expects.json` file which contains expected test results and descriptions for their expected behavior. You should run this after generating certificates. `node defineExpects.js`

The `generateApacheConf.js` script generates an Apache configuration using your test suite's certificates. You may need to update the paths in this script as appropriate for your system. You can then generate an apache config by running it, e.g. `node generateApacheConf.js > /etc/apache2/sites-enabled/001-bettertls.conf`.

The website and javascript for running the in-browser test suite is in the [docs](../docs) directory. If you have done the above to configure for running locally and you have setup Apache, you should be able to browse to http://localhost:8000.

The [testsuites](testsuites) directory contains scripts for running the BetterTLS test suite for non-browser clients. Take a look at [runcurl.js](testsuites/runcurl.js) for a simple example.
