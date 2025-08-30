.PHONY: install-plugin generate test benchmark clean help

# Build and install the plugin from the adjacent directory
install-plugin:
	cd ../protogo-values && make build && make install

# Generate code using protoc directly to handle include paths
generate: install-plugin
	mkdir -p gen
	protoc \
		--proto_path=. \
		--proto_path=../protogo-values/proto \
		--go-values_out=gen \
		--go-values_opt=paths=source_relative \
		api/validation/v1/types.proto

# Run validation tests
test: generate
	go test -v ./internal/validation -run Test

# Run performance benchmarks
benchmark: generate
	go test -bench=. -benchmem ./internal/validation

# Clean generated files
clean:
	rm -rf gen/

# Show available commands
help:
	@echo "Available commands:"
	@echo "  install-plugin - Install protoc-gen-go-values from ../protogo-values/"
	@echo "  generate       - Generate Go code from protobuf definitions using protoc"
	@echo "  test          - Run validation tests"
	@echo "  benchmark     - Run performance benchmarks"
	@echo "  clean         - Remove generated files"