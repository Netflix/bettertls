#!/usr/bin/env python

"""

    Copyright 2017 Netflix, Inc.

       Licensed under the Apache License, Version 2.0 (the "License");
       you may not use this file except in compliance with the License.
       You may obtain a copy of the License at

           http://www.apache.org/licenses/LICENSE-2.0

       Unless required by applicable law or agreed to in writing, software
       distributed under the License is distributed on an "AS IS" BASIS,
       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
       See the License for the specific language governing permissions and
       limitations under the License.

"""

import requests
import sys

if sys.argv[1] == '--useragent':

    from requests.packages.urllib3 import util
    if util.IS_PYOPENSSL:
        import OpenSSL.SSL
        openssl_version = OpenSSL.SSL.SSLeay_version(OpenSSL.SSL.SSLEAY_VERSION)
    else:
        import ssl
        openssl_version = ssl.OPENSSL_VERSION

    print("Python %s, requests %s, %s" % (sys.version.replace("\n", ' ').strip(), requests.__version__, openssl_version))
    sys.exit(0)

try:
    r = requests.get(sys.argv[1], verify='../../docs/root.crt')
    exit_code = 0 if r.status_code == 200 else 1
    sys.exit(exit_code)
except Exception:
    sys.exit(1)

