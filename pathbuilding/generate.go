package pathbuilding

//go:generate go build -o ./build/protoc-gen-go github.com/golang/protobuf/protoc-gen-go
//go:generate protoc --plugin=./build/protoc-gen-go --go_out=paths=source_relative:./ -I . ./hostname_encoded_test_case.proto
