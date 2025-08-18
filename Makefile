.PHONY: proto clean

# The 'proto' target will generate the Go files from the .proto definition.
proto:
	@echo "Generating protobuf files..."
	# Ensure the output directory exists
	mkdir -p gen
	# Run the protoc compiler
	find proto -name '*.proto' -exec protoc \
		--proto_path=proto \
		--go_out=./gen --go_opt=paths=source_relative \
		--go-grpc_out=./gen --go-grpc_opt=paths=source_relative {} +

clean:
	@echo "Cleaning up generated files..."
	rm -rf proto/gen
