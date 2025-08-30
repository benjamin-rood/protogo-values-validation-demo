# Protogo Values Validation Demo

A comprehensive validation platform for the `protoc-gen-go-values` plugin that provides end-to-end testing of plugin functionality through realistic usage scenarios, performance benchmarks, and production-ready deployment patterns.

## Purpose

This repository validates the core functionality of the `protoc-gen-go-values` plugin, which generates value slices (`[]Type`) instead of pointer slices (`[]*Type`) for protobuf repeated fields when annotated with `(protogo_values.value_slice) = true`.

## Key Validation Areas

### Plugin Integration
- Validates correct Go code generation from protobuf definitions with field options
- Ensures `(protogo_values.value_slice) = true` annotations produce value slices
- Maintains pointer slices for non-annotated fields

### Type System Verification
- Generated code compiles successfully with Go compiler
- Produces expected slice types for annotated fields
- Passes static type validation and maintains protobuf compatibility

### Performance Testing
- Demonstrates reduced memory allocations for value slice access
- Shows improved cache locality for iteration operations
- Measures performance improvements through comprehensive benchmarks
- Validates zero allocations for slice length operations

### Integration Testing
- Real-world gRPC service operation with generated types
- Message serialization and deserialization compatibility
- gRPC streaming operation support
- Large message volume handling without degradation

### Deployment Validation
- Kubernetes deployment in production-like environments
- Health check validation and service availability
- Load testing and monitoring integration

## Architecture

The validation platform includes:

- **Validation Service**: gRPC service with REST endpoints for type validation and benchmarking
- **Test Definitions**: Comprehensive protobuf definitions with mixed value/pointer slice configurations
- **Performance Benchmarks**: Comparative testing between value and pointer slice performance
- **Kubernetes Deployment**: Production-ready manifests with monitoring and health checks
- **CI/CD Pipeline**: Automated validation across multiple Go versions
- **Integration Scripts**: Automated validation and testing scripts

## Success Criteria

- [ ] Plugin generates correct Go types for all test scenarios
- [ ] Generated code compiles without errors or warnings
- [ ] Type assertions pass for all field configurations
- [ ] Performance benchmarks show measurable improvements
- [ ] Integration tests pass with real gRPC communication
- [ ] Kubernetes deployment succeeds in multiple environments
- [ ] Compatible with standard protobuf toolchain and Buf build system

## Getting Started

1. Install dependencies (Go, Buf, protoc-gen-go-values plugin)
2. Generate code: `buf generate`
3. Run validation tests: `go test -v ./internal/validation`
4. Run performance benchmarks: `go test -bench=. -benchmem ./internal/validation`
5. Deploy to Kubernetes: `kubectl apply -f deployment/`

This platform serves as both a validation tool and reference implementation for using the protoc-gen-go-values plugin in production environments.