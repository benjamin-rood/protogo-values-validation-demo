# Protogo Values Validation Platform

A comprehensive validation platform for the `protoc-gen-go-values` plugin that validates plugin functionality through protobuf integration, type verification, performance testing, and gRPC service validation.

## Overview

This platform validates that the `protoc-gen-go-values` plugin correctly transforms repeated protobuf fields marked with field options from pointer slices (`[]*Type`) to value slices (`[]Type`). It demonstrates seamless integration with the buf ecosystem and supports both simple and structured field option formats.

**Phase 1: Core Infrastructure** - ✅ Complete  
**Phase 2: Validation Logic** - In Progress

See the [validation platform specification](specs/validation-platform/validation-platform.spec.md) for detailed requirements and implementation phases.

## Project Structure

```
protogo-values-validation-demo/
├── api/validation/v1/           # Protobuf definitions with field options
│   ├── types.proto             # Test messages using plugin field options  
│   └── validation.proto        # ValidationService gRPC definition
├── internal/validation/         # Test implementations
│   ├── types_test.go           # Type validation tests
│   └── benchmark_test.go       # Performance benchmarks
├── gen/api/validation/v1/       # Generated Go code 
│   └── types.pb.go             # Generated types with plugin transformations
├── .github/workflows/           # CI/CD Pipeline
│   └── validate.yaml           # Automated plugin validation
├── specs/                      # Specifications
│   ├── mvp-validation/         # MVP specification
│   └── validation-platform/    # Full platform specification  
├── buf.gen.yaml                # Code generation config with plugin
├── buf.yaml                    # Buf module configuration
├── buf.lock                    # Dependency lock file
├── go.mod                      # Go module with dependencies
├── Makefile                    # Build automation
└── .gitignore                  # Excludes generated files
```

## Quick Start

### Prerequisites
- Go 1.20+
- [Buf CLI](https://buf.build/docs/installation) for protobuf management
- Local plugin from `../protogo-values/`

**Buf Workspace**: This project uses buf workspace configuration to resolve cross-module imports between the validation platform and the plugin's protobuf definitions.

### Usage

```bash
# Generate code using buf with plugin workspace integration
buf generate --template buf.gen.yaml

# Run type validation tests  
go test -v ./internal/validation/

# Run performance benchmarks
go test -bench=. -benchmem ./internal/validation/

# Run all tests and benchmarks
make test && make benchmark

# Show all available commands
make help
```

## Buf Workspace Configuration

This project demonstrates proper buf ecosystem integration:

**Root Workspace** (`/protoc-plugin-go-values/buf.work.yaml`):
```yaml
version: v1
directories:
  - protogo-values/proto        # Plugin protobuf definitions
  - protogo-values-validation-demo  # Validation platform
```

**Cross-Module Import Resolution**:
- `protogo_values/options.proto` → Resolved via workspace  
- Plugin field options work seamlessly with buf generate
- No manual proto copying or path manipulation required

## What Gets Validated

### Type Transformations Validated

**ValidationTestMessage (MVP compatibility)**:
- ✅ `repeated DataPoint value_slice_data = 1 [(protogo_values.value_slice) = true]` → `[]DataPoint`
- ✅ `repeated DataPoint pointer_slice_data = 2` → `[]*DataPoint` (unchanged)  
- ✅ `repeated MetricPoint metrics = 3 [(protogo_values.field_opts).value_slice = true]` → `[]MetricPoint`

**PerformanceTestMessage (Phase 1 spec-compliant)**:
- ✅ `repeated DataPoint value_slice_data = 1 [(protogo_values.value_slice) = true]` → `[]DataPoint`
- ✅ `repeated Metadata pointer_slice_data = 2` → `[]*Metadata` (control group)
- ✅ `repeated ProcessingResult results = 3 [(protogo_values.value_slice) = true]` → `[]ProcessingResult`

### Performance Comparisons
- Memory allocation differences during iteration
- Cache locality improvements for value slices  
- Zero-allocation access for slice operations
- Benchmarks comparing value vs pointer slice performance

### Buf Ecosystem Integration
- ✅ **Buf workspace cross-module imports working**
- ✅ **Plugin integration with buf generate**
- ✅ **Field options resolved via workspace**
- ✅ **CI/CD pipeline with automated validation**
- ✅ **Multi-Go version compatibility testing**

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

## Implementation Phases

**✅ Phase 1: Core Infrastructure (Complete)**
- Repository structure and buf workspace setup
- Protobuf definitions with ValidationService
- Plugin integration and cross-module imports
- Foundational validation tests
- CI/CD pipeline with GitHub Actions

**🚀 Phase 2: Validation Logic (Next)**
- Comprehensive type validation service
- Performance benchmarking suite  
- Integration testing framework
- Validation service implementation

**📋 Phase 3: Deployment & Monitoring** 
- Kubernetes manifests and deployment
- Monitoring and alerting integration
- Load testing capabilities
- Production-ready documentation

**🔒 Phase 4: Production Readiness**
- Security review and hardening
- Performance optimization
- Comprehensive testing suite
- Final validation and deployment

See the [complete specification](specs/validation-platform/validation-platform.spec.md) for the full platform roadmap.