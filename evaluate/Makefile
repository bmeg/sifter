
# pip install grpcio-tools
proto:
	@protoc \
		-I ./ \
		--lint_out=. \
		--go_out=Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		exec.proto
	@python -m grpc_tools.protoc -I ./ --python_out=./ --grpc_python_out=./ exec.proto

proto-depends:
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@go install github.com/golang/protobuf/protoc-gen-go
	@go install github.com/ckaznocha/protoc-gen-lint

bindata:
	go-bindata -pkg evaluate ./*.py

bindata-depends:
	go get -u github.com/shuLhan/go-bindata/...
