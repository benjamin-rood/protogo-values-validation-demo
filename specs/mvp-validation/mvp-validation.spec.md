# MVP Validation Platform Specification

## Overview

This specification defines a Minimum Viable Product (MVP) for validating the `protoc-gen-go-values` plugin. The MVP focuses on core functionality validation through basic protobuf integration, type verification, and plugin integration testing.

**Reference**: This MVP is derived from the complete [validation-platform.spec.md](../validation-platform/validation-platform.spec.md) specification, implementing the essential components needed to validate plugin functionality.

## MVP Scope

### Core Validation Requirements
- **R1**: Plugin integration with local binary from `../protogo-values/`
- **R2**: Type transformation verification (`[]*Type` → `[]Type`)
- **R3**: Basic performance comparison (value vs pointer slices)
- **R4**: Buf integration for code generation

### Out of Scope (Full Platform Features)
- gRPC service implementation
- Kubernetes deployment
- CI/CD pipeline
- Comprehensive monitoring
- Production deployment scripts

## MVP Architecture

```
protogo-values-validation-demo/
├── api/                          # Protobuf definitions
│   └── validation/
│       └── v1/
│           ├── types.proto       # Test message types with field options
│           └── buf.yaml          # Buf module configuration
├── internal/                     # Implementation
│   └── validation/
│       ├── types_test.go         # Type validation tests
│       └── benchmark_test.go     # Basic performance comparison
├── buf.gen.yaml                  # Code generation configuration
├── go.mod                        # Go module
└── Makefile                      # Build automation
```

## Core Requirements (EARS Format)

### R1: Local Plugin Integration
**Event:** Code generation from protobuf definitions  
**Condition:** WHEN using buf generate with local plugin binary  
**Response:** System SHALL:
- Use `protoc-gen-go-values` from `../protogo-values/cmd/protoc-gen-go-values`
- Generate Go code with field option transformations
- Maintain compatibility with standard protoc-gen-go output

### R2: Type System Validation
**Event:** Generated code compilation and type verification  
**Condition:** WHEN fields are marked with `(protogo_values.value_slice) = true`  
**Response:** System SHALL:
- Generate value slices (`[]Type`) for annotated repeated fields
- Preserve pointer slices (`[]*Type`) for non-annotated fields
- Produce compilable Go code without errors

### R3: Basic Performance Validation  
**Event:** Performance comparison between slice types  
**Condition:** WHEN benchmarking equivalent operations  
**Response:** System SHALL demonstrate:
- Measurable performance difference in iteration operations
- Memory allocation differences between slice types
- Zero-allocation access for slice length operations

## Technical Specifications

### Protobuf Definitions

#### api/validation/v1/types.proto
```proto
syntax = "proto3";

package validation.v1;

import "proto/protogo_values/options.proto";

option go_package = "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1";

// Test message for MVP validation
message ValidationTestMessage {
  // Should generate []DataPoint (value slice)
  repeated DataPoint value_slice_data = 1 [(protogo_values.value_slice) = true];
  
  // Should remain []*DataPoint (pointer slice - control group)
  repeated DataPoint pointer_slice_data = 2;
  
  // Test structured field option format
  repeated MetricPoint metrics = 3 [(protogo_values.field_opts).value_slice = true];
}

message DataPoint {
  string id = 1;
  double value = 2;
  int64 timestamp = 3;
}

message MetricPoint {
  string name = 1;
  double measurement = 2;
  map<string, string> labels = 3;
}
```

### Build Configuration

#### buf.gen.yaml
```yaml
version: v1
plugins:
  # Standard Go protobuf generation
  - plugin: go
    out: gen
    opt:
      - paths=source_relative
  # Our custom plugin from local build
  - plugin: go-values
    path: ../protogo-values/protoc-gen-go-values
    out: gen
    opt:
      - paths=source_relative
```

#### api/validation/v1/buf.yaml
```yaml
version: v1
name: buf.build/validation-demo/api
deps:
  - buf.build/protogo-values/options
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

### Validation Tests

#### internal/validation/types_test.go
```go
package validation

import (
    "reflect"
    "testing"
    
    v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
)

func TestPluginTypeTransformation(t *testing.T) {
    tests := []struct {
        name           string
        fieldName      string
        expectedType   string
        getActualType  func() interface{}
    }{
        {
            name:         "value_slice_data should be value slice",
            fieldName:    "ValueSliceData", 
            expectedType: "[]v1.DataPoint",
            getActualType: func() interface{} {
                return v1.ValidationTestMessage{}.ValueSliceData
            },
        },
        {
            name:         "pointer_slice_data should remain pointer slice",
            fieldName:    "PointerSliceData",
            expectedType: "[]*v1.DataPoint", 
            getActualType: func() interface{} {
                return v1.ValidationTestMessage{}.PointerSliceData
            },
        },
        {
            name:         "metrics should be value slice (structured option)",
            fieldName:    "Metrics",
            expectedType: "[]v1.MetricPoint",
            getActualType: func() interface{} {
                return v1.ValidationTestMessage{}.Metrics
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actualType := reflect.TypeOf(tt.getActualType()).String()
            if actualType != tt.expectedType {
                t.Errorf("Field %s has type %s, expected %s", 
                    tt.fieldName, actualType, tt.expectedType)
            }
        })
    }
}

func TestPluginIntegrationWorking(t *testing.T) {
    // Verify our test message can be created and used
    msg := &v1.ValidationTestMessage{
        ValueSliceData: []v1.DataPoint{
            {Id: "test1", Value: 1.0, Timestamp: 123},
            {Id: "test2", Value: 2.0, Timestamp: 124},
        },
        PointerSliceData: []*v1.DataPoint{
            {Id: "ptr1", Value: 3.0, Timestamp: 125},
        },
        Metrics: []v1.MetricPoint{
            {Name: "cpu", Measurement: 0.85},
        },
    }
    
    if len(msg.ValueSliceData) != 2 {
        t.Errorf("Expected 2 value slice items, got %d", len(msg.ValueSliceData))
    }
    
    if len(msg.PointerSliceData) != 1 {
        t.Errorf("Expected 1 pointer slice item, got %d", len(msg.PointerSliceData))
    }
    
    if len(msg.Metrics) != 1 {
        t.Errorf("Expected 1 metric item, got %d", len(msg.Metrics))
    }
}
```

#### internal/validation/benchmark_test.go
```go
package validation

import (
    "testing"
    
    v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
)

const benchmarkDataSize = 1000

func BenchmarkValueSliceIteration(b *testing.B) {
    // Create test data with value slices
    valueSlices := make([]v1.DataPoint, benchmarkDataSize)
    for i := 0; i < benchmarkDataSize; i++ {
        valueSlices[i] = v1.DataPoint{
            Id: "test",
            Value: float64(i),
            Timestamp: int64(i),
        }
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        sum := 0.0
        for _, dp := range valueSlices {
            sum += dp.Value
        }
        _ = sum
    }
}

func BenchmarkPointerSliceIteration(b *testing.B) {
    // Create test data with pointer slices
    pointerSlices := make([]*v1.DataPoint, benchmarkDataSize)
    for i := 0; i < benchmarkDataSize; i++ {
        pointerSlices[i] = &v1.DataPoint{
            Id: "test",
            Value: float64(i), 
            Timestamp: int64(i),
        }
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        sum := 0.0
        for _, dp := range pointerSlices {
            sum += dp.Value
        }
        _ = sum
    }
}

func BenchmarkSliceLengthAccess(b *testing.B) {
    valueSlices := make([]v1.DataPoint, benchmarkDataSize)
    
    b.Run("ValueSlice", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _ = len(valueSlices)
        }
    })
    
    pointerSlices := make([]*v1.DataPoint, benchmarkDataSize)
    
    b.Run("PointerSlice", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _ = len(pointerSlices)
        }
    })
}
```

### Build Automation

#### Makefile
```make
.PHONY: install-plugin generate test benchmark clean help

# Install the plugin from the adjacent directory
install-plugin:
	cd ../protogo-values && make install

# Generate code using buf
generate: install-plugin
	buf generate api/validation/v1

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
	@echo "  generate       - Generate Go code from protobuf definitions"
	@echo "  test          - Run validation tests"
	@echo "  benchmark     - Run performance benchmarks"
	@echo "  clean         - Remove generated files"
```

## Success Criteria

### MVP Validation Goals
- [ ] Plugin successfully transforms annotated fields to value slices
- [ ] Non-annotated fields remain as pointer slices
- [ ] Generated code compiles without errors
- [ ] Tests pass and demonstrate type correctness
- [ ] Benchmarks show measurable performance difference
- [ ] Both field option formats work (simple and structured)

### MVP Deliverables
1. **Working Plugin Integration** - Code generation using local plugin
2. **Type Validation Tests** - Automated verification of transformations
3. **Performance Benchmarks** - Quantitative comparison of slice types
4. **Build Automation** - Simple Makefile for development workflow

## Implementation Plan

### Phase 1: Foundation (Day 1)
1. Create directory structure
2. Set up Go module and dependencies
3. Create basic protobuf definitions
4. Configure buf for local plugin

### Phase 2: Integration (Day 2)  
1. Implement plugin installation and code generation
2. Create type validation tests
3. Verify plugin transformation works correctly
4. Debug any integration issues

### Phase 3: Validation (Day 3)
1. Implement performance benchmarks
2. Add comprehensive test coverage
3. Create build automation
4. Document usage and results

## Extension Path

This MVP provides the foundation for extending to the full validation platform:
- **Service Layer**: Add gRPC service using validated types
- **Deployment**: Add containerization and Kubernetes manifests
- **CI/CD**: Implement automated testing pipeline
- **Monitoring**: Add performance tracking and alerting

The MVP validates the core plugin functionality and provides a working reference for building the complete validation platform.