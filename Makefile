.PHONY: proto clean

# The 'proto' target will generate the Go files from the .proto definition.
proto:
	@echo "Generating protobuf files..."
	# Ensure the output directory exists
	mkdir -p proto/gen
	# Run the protoc compiler
	protoc --go_out=./proto/gen --go_opt=paths=source_relative \
	    --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
	    proto/account/v1/account.proto

clean:
	@echo "Cleaning up generated files..."
	rm -rf proto/gen
