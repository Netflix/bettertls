package test_executor

//go:generate go build -o protoc-gen-go github.com/golang/protobuf/protoc-gen-go
//go:generate protoc --plugin=./protoc-gen-go --go_out=. --go_opt=paths=source_relative test_results.proto
//go:generate rm -f protoc-gen-go
