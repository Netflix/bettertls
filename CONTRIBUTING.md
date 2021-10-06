Contributing
============

We welcome contributions big and small! The goal of BetterTLS is to add anything you think can of to improve the [bettertls.com](https://bettertls.com) site or the ecosystem of TLS as a whole.

One of the easiest things you could do to get started is add another HTTPS client to our growing collections of test suites. We've added several clients (such as [lua](testsuites/runlua.js), [openssl](testsuites/runopenssl.js), and [ruby](testsuites/runruby.js)) since launching, and as you can see from those examples it doesn't take much to do it. These testsuites run against the public website, so all you need to do is download the test [root certificate](https://nameconstraints.bettertls.com/root.crt) to `docs/root.crt`. You don't even need to worry about generating certificates for the test suite to get started!

If you'd like to extend BetterTLS with additional tests for name constraints (or even some other TLS specification like [HSTS](https://tools.ietf.org/html/rfc6797) or [HPKP](https://tools.ietf.org/html/rfc7469)), instructions to checkout and modify this repository can be found in the [README](README.md). You will need [Node.JS](https://nodejs.org/en/download) to run most of the tools and Java to generate certificates. The scripts produce an an [Apache](https://httpd.apache.org/) config.



Creating PRs and Issues
=======================

We don't have any strict requirements or procedures for contributing, so don't hesitate to open up a PR or issue!



Contact Us
==========

If you would rather get in touch with us directly, feel free to email us at bettertls@netflix.com.
