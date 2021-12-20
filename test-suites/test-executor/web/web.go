package web

import "embed"

//go:generate npm install protobufjs
//go:generate node_modules/.bin/pbjs -t static-module -w closure -o test_results.js ../test_results.proto
//go:generate cp test_results.js ../../../docs
//go:generate rm -rf node_modules package-lock.json

//go:embed index.html test_results.js
var Content embed.FS
