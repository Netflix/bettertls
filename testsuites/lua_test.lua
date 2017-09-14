#!/usr/bin/env lua

--[[

    Copyright 2017 Daurnimator

       Licensed under the Apache License, Version 2.0 (the "License");
       you may not use this file except in compliance with the License.
       You may obtain a copy of the License at

           http://www.apache.org/licenses/LICENSE-2.0

       Unless required by applicable law or agreed to in writing, software
       distributed under the License is distributed on an "AS IS" BASIS,
       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
       See the License for the specific language governing permissions and
       limitations under the License.

]]


if arg[1] == "--useragent" then
	local http_version = require "http.version"
	local luaossl = require "openssl"

	print(string.format("%s, %s %s, %s",
		_VERSION,
		http_version.name,
		http_version.version,
		luaossl.version(luaossl.SSLEAY_VERSION)
	))
	return os.exit(0)
end

local http_request = require "http.request"
local http_tls = require "http.tls"
local ossl_store = require "openssl.x509.store"

local uri = assert(arg[1], "missing argument (expected URI)")

local request = http_request.new_from_uri(uri)
-- Create TLS context with certificate store containing a single trusted root cert
request.ctx = http_tls.new_client_context()
request.ctx:setStore(ossl_store.new():add("../certificates/root.crt"))
local headers = request:go()
if not headers or headers:get ":status" ~= "200" then
	return os.exit(1)
end
return os.exit(0)
