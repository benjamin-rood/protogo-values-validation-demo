# Protogo Values Validation Demo - MVP

A **Minimum Viable Product** validation platform for the `protoc-gen-go-values` plugin that validates core plugin functionality through protobuf integration, type verification, and performance testing.

## Overview

This MVP validates that the `protoc-gen-go-values` plugin correctly transforms repeated protobuf fields marked with `(protogo_values.value_slice) = true` from pointer slices (`[]*Type`) to value slices (`[]Type`).

**Full Platform**: This MVP implements the essential components from the complete [validation platform specification](specs/validation-platform/validation-platform.spec.md). See [MVP specification](specs/mvp-validation/mvp-validation.spec.md) for detailed requirements.

## Project Structure

```
protogo-values-validation-demo/
├── api/validation/v1/           # Protobuf definitions with field options
│   ├── types.proto             # Test messages using plugin field options  
│   └── buf.yaml                # Buf module configuration
├── internal/validation/         # Test implementations
│   ├── types_test.go           # Type validation tests
│   └── benchmark_test.go       # Performance benchmarks
├── specs/                      # Specifications
│   ├── mvp-validation/         # MVP specification
│   └── validation-platform/    # Full platform specification  
├── buf.gen.yaml                # Code generation config (uses local plugin)
├── buf.yaml                    # Root buf configuration
├── go.mod                      # Go module
└── Makefile                    # Build automation
```

## Quick Start

### Prerequisites
- Go 1.24+
- [Buf CLI](https://docs.buf.build/installation)
- Local plugin from `../protogo-values/`

### Usage

```bash
# Install plugin from adjacent directory and generate code
make generate

# Run type validation tests  
make test

# Run performance benchmarks
make benchmark

# Clean generated files
make clean
```

## What Gets Validated

### Type Transformations
- ✅ `repeated DataPoint value_slice_data = 1 [(protogo_values.value_slice) = true]` → `[]DataPoint`
- ✅ `repeated DataPoint pointer_slice_data = 2` → `[]*DataPoint` (unchanged)  
- ✅ `repeated MetricPoint metrics = 3 [(protogo_values.field_opts).value_slice = true]` → `[]MetricPoint`

### Performance Comparisons
- Memory allocation differences during iteration
- Cache locality improvements for value slices
- Zero-allocation access for slice operations

### Plugin Integration
- Local plugin binary integration via Buf
- Code generation from protobuf with field options
- Compilation verification of generated Go code

## Example Test Output

```go
// Generated types validation
func TestPluginTypeTransformation(t *testing.T) {
    // Validates that ValueSliceData is []v1.DataPoint (not []*v1.DataPoint)
    // Validates that PointerSliceData remains []*v1.DataPoint  
    // Validates that structured field options work correctly
}

// Performance comparison
func BenchmarkValueSliceIteration(b *testing.B) {
    // Demonstrates performance differences between slice types
}
```

## Extension Path

This MVP provides the foundation for the full validation platform:
- **Service Layer**: Add gRPC validation service
- **Deployment**: Add Kubernetes manifests and CI/CD  
- **Monitoring**: Add performance tracking and alerting
- **Scale Testing**: Add load testing and production scenarios

See the [complete specification](specs/validation-platform/validation-platform.spec.md) for the full platform roadmap.