#!/usr/bin/ruby

require 'net/http'
require 'net/https'

if ARGV[0] == '--useragent'
  puts "Ruby: #{RUBY_VERSION}-p#{RUBY_PATCHLEVEL}, #{OpenSSL::OPENSSL_VERSION}"
  exit(0)
end

url = URI(ARGV[0])
http = Net::HTTP.new(url.host, url.port)
http.use_ssl = true
http.ca_file = '../docs/root.crt'
http.verify_mode = OpenSSL::SSL::VERIFY_PEER
begin
  http.get(url.path)
rescue => e
  exit(1)
end

exit(0)
