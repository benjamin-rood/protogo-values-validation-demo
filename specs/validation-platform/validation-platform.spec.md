# Validation Platform Sub-Repository Specification

## Overview

This specification defines a separate sub-repository that serves as a comprehensive validation platform for the `protoc-gen-go-values` plugin. The platform provides end-to-end validation of plugin functionality through realistic usage scenarios, performance benchmarks, and production-ready deployment patterns.

## Repository Structure

```
protogo-values-validator/
├── plugin/                          # Plugin integration
│   ├── buf.yaml                     # Buf configuration for plugin
│   └── buf.gen.yaml                 # Code generation configuration
├── service/                         # Validation service
│   ├── api/
│   │   └── v1/
│   │       ├── validation.proto     # Validation service definition
│   │       └── types.proto          # Test message types
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # Service implementation
│   ├── internal/
│   │   ├── server/
│   │   │   └── handlers.go         # gRPC handlers
│   │   └── validation/
│   │       ├── types_test.go       # Type validation tests
│   │       ├── performance_test.go # Performance benchmarks
│   │       └── integration_test.go # End-to-end tests
│   ├── go.mod
│   └── Dockerfile
├── deployment/                      # Kubernetes manifests
│   ├── base/
│   │   ├── namespace.yaml
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   ├── jobs/
│   │   ├── validation-job.yaml
│   │   └── benchmark-job.yaml
│   └── monitoring/
│       ├── prometheus.yaml
│       └── grafana-dashboard.json
├── scripts/                         # Validation scripts
│   ├── validate-types.sh
│   ├── run-benchmarks.sh
│   └── integration-test.sh
├── .github/
│   └── workflows/
│       ├── validate.yaml
│       └── performance.yaml
└── README.md
```

## Core Requirements

### R1: Plugin Integration Validation
**Event:** Plugin generates Go code from protobuf definitions with field options
**Condition:** WHEN field options `(protogo_values.value_slice) = true` are present
**Response:** System SHALL generate value slices (`[]Type`) instead of pointer slices (`[]*Type`)

### R2: Type System Validation
**Event:** Generated code compilation and type checking
**Condition:** WHEN code is compiled with Go compiler
**Response:** System SHALL:
- Generate syntactically correct Go code
- Produce expected slice types for annotated fields
- Maintain pointer slices for non-annotated fields
- Pass static type validation

### R3: Performance Validation
**Event:** Performance comparison between value and pointer slices
**Condition:** WHEN processing equivalent data structures
**Response:** System SHALL demonstrate:
- Reduced memory allocations for value slice access
- Improved cache locality for iteration operations
- Measurable performance improvements in benchmarks
- Zero allocations for slice length operations

### R4: Integration Testing
**Event:** Real-world gRPC service operation
**Condition:** WHEN service handles actual requests with generated types
**Response:** System SHALL:
- Successfully serialize and deserialize messages
- Maintain compatibility with standard protobuf tooling
- Support gRPC streaming operations
- Handle large message volumes without degradation

### R5: Deployment Validation
**Event:** Kubernetes deployment of validation service
**Condition:** WHEN deployed to production-like environment
**Response:** System SHALL:
- Successfully deploy with health checks
- Serve validation endpoints
- Pass readiness and liveness probes
- Maintain service availability under load

## Technical Specifications

### 2.1 Protobuf Definitions

#### validation.proto
```proto
syntax = "proto3";

package validation.v1;

import "google/api/annotations.proto";
import "protogo_values/options.proto";

option go_package = "github.com/benjamin-rood/protogo-values-validator/gen/api/v1";

service ValidationService {
  // Validates type generation correctness
  rpc ValidateTypes(ValidateTypesRequest) returns (ValidateTypesResponse) {
    option (google.api.http) = {
      post: "/v1/validate/types"
      body: "*"
    };
  }

  // Benchmarks performance characteristics
  rpc RunBenchmarks(BenchmarkRequest) returns (BenchmarkResponse) {
    option (google.api.http) = {
      post: "/v1/validate/benchmark"
      body: "*"
    };
  }

  // Stream processing validation
  rpc StreamValidation(stream StreamRequest) returns (stream StreamResponse);
}
```

#### types.proto
```proto
syntax = "proto3";

package validation.v1;

import "protogo_values/options.proto";

// Message with value slices for performance testing
message PerformanceTestMessage {
  // Should generate []DataPoint, not []*DataPoint
  repeated DataPoint value_slice_data = 1 [(protogo_values.value_slice) = true];
  
  // Should remain []*Metadata (control group)
  repeated Metadata pointer_slice_data = 2;
  
  // Nested value slices
  repeated ProcessingResult results = 3 [(protogo_values.value_slice) = true];
}

message DataPoint {
  string id = 1;
  double value = 2;
  int64 timestamp = 3;
  repeated string tags = 4;
}

message Metadata {
  string key = 1;
  string value = 2;
  map<string, string> attributes = 3;
}

message ProcessingResult {
  string operation_id = 1;
  bool success = 2;
  double duration_ms = 3;
  repeated string error_messages = 4;
}
```

### 2.2 Validation Tests

#### Type Validation
```go
func TestGeneratedTypeValidation(t *testing.T) {
    tests := []struct {
        name       string
        fieldPath  string
        expectType string
        actualType any
    }{
        {
            name:       "value_slice_data should be value slice",
            fieldPath:  "PerformanceTestMessage.ValueSliceData",
            expectType: "[]v1.DataPoint",
            actualType: v1.PerformanceTestMessage{}.ValueSliceData,
        },
        {
            name:       "pointer_slice_data should remain pointer slice",
            fieldPath:  "PerformanceTestMessage.PointerSliceData",
            expectType: "[]*v1.Metadata",
            actualType: v1.PerformanceTestMessage{}.PointerSliceData,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actualType := reflect.TypeOf(tt.actualType).String()
            assert.Equal(t, tt.expectType, actualType,
                "Field %s has incorrect type", tt.fieldPath)
        })
    }
}
```

#### Performance Benchmarks
```go
func BenchmarkValueSliceVsPointerSlice(b *testing.B) {
    const dataSize = 10000
    
    b.Run("ValueSlice/Iteration", func(b *testing.B) {
        msg := createTestMessageWithValueSlices(dataSize)
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            sum := benchmarkValueSliceIteration(msg.ValueSliceData)
            _ = sum
        }
    })
    
    b.Run("PointerSlice/Iteration", func(b *testing.B) {
        msg := createTestMessageWithPointerSlices(dataSize)
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            sum := benchmarkPointerSliceIteration(msg.PointerSliceData)
            _ = sum
        }
    })
    
    b.Run("ValueSlice/Access", func(b *testing.B) {
        msg := createTestMessageWithValueSlices(dataSize)
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            _ = len(msg.ValueSliceData)
        }
    })
}
```

### 2.3 Kubernetes Validation Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: validation-job
  namespace: protogo-validation
spec:
  template:
    spec:
      containers:
      - name: validator
        image: protogo/validator:latest
        command: ["/scripts/validate-types.sh"]
        env:
        - name: VALIDATION_SERVICE_URL
          value: "http://validation-service:8080"
        - name: EXPECTED_VALUE_SLICE_COUNT
          value: "2"
        - name: EXPECTED_POINTER_SLICE_COUNT
          value: "1"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
      restartPolicy: Never
  backoffLimit: 3
```

### 2.4 CI/CD Pipeline

#### Validation Workflow
```yaml
name: Plugin Validation

on:
  schedule:
    - cron: '0 2 * * *'  # Daily validation
  workflow_dispatch:
    inputs:
      plugin_version:
        description: 'Plugin version to validate'
        required: true
        default: 'latest'

jobs:
  validate-plugin:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.20', '1.21']
        
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Install Dependencies
      run: |
        go install github.com/bufbuild/buf/cmd/buf@latest
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    - name: Install Plugin
      run: |
        go install github.com/benjamin-rood/protogo-values/cmd/protoc-gen-go-values@${{ inputs.plugin_version || 'latest' }}
    
    - name: Generate Code
      run: |
        cd service
        buf generate
    
    - name: Run Type Validation
      run: |
        cd service
        go test -v ./internal/validation -run TestGeneratedTypeValidation
    
    - name: Run Performance Benchmarks
      run: |
        cd service
        go test -bench=. -benchmem ./internal/validation
    
    - name: Build Service
      run: |
        cd service
        docker build -t validation-service:test .
    
    - name: Integration Test
      run: |
        docker run -d --name validation-test -p 8080:8080 validation-service:test
        sleep 10
        ./scripts/integration-test.sh
        docker logs validation-test
        docker stop validation-test
```

### 2.5 Validation Scripts

#### validate-types.sh
```bash
#!/bin/bash
set -e

echo "Validating generated types..."

# Check service health
curl -f "${VALIDATION_SERVICE_URL}/health" || {
    echo "ERROR: Validation service not healthy"
    exit 1
}

# Validate generated code structure
echo "Checking generated protobuf code..."

GENERATED_DIR="/app/gen/api/v1"
if [[ ! -d "$GENERATED_DIR" ]]; then
    echo "ERROR: Generated code directory not found"
    exit 1
fi

# Count value slices (should be []Type, not []*Type)
VALUE_SLICE_COUNT=$(grep -c "^\s*\w\+\s\+\[\]\w\+" "$GENERATED_DIR"/*.pb.go || true)
echo "Found $VALUE_SLICE_COUNT value slices"

# Count pointer slices (should be []*Type)
POINTER_SLICE_COUNT=$(grep -c "^\s*\w\+\s\+\[\]\*\w\+" "$GENERATED_DIR"/*.pb.go || true)
echo "Found $POINTER_SLICE_COUNT pointer slices"

# Validate expected counts
if [[ "$VALUE_SLICE_COUNT" != "$EXPECTED_VALUE_SLICE_COUNT" ]]; then
    echo "ERROR: Expected $EXPECTED_VALUE_SLICE_COUNT value slices, found $VALUE_SLICE_COUNT"
    exit 1
fi

if [[ "$POINTER_SLICE_COUNT" != "$EXPECTED_POINTER_SLICE_COUNT" ]]; then
    echo "ERROR: Expected $EXPECTED_POINTER_SLICE_COUNT pointer slices, found $POINTER_SLICE_COUNT"
    exit 1
fi

# Run type validation via service
echo "Running service-based type validation..."
curl -X POST "${VALIDATION_SERVICE_URL}/v1/validate/types" \
    -H "Content-Type: application/json" \
    -d '{}' \
    --fail-with-body || {
    echo "ERROR: Type validation service call failed"
    exit 1
}

echo "Type validation completed successfully!"
```

## Success Criteria

### 3.1 Functional Validation
- [ ] Plugin generates correct Go types for all test scenarios
- [ ] Generated code compiles without errors or warnings
- [ ] Type assertions pass for all field configurations
- [ ] Integration tests pass with real gRPC communication

### 3.2 Performance Validation
- [ ] Value slice iteration shows measurable performance improvement
- [ ] Memory allocation tests demonstrate reduced allocations
- [ ] Benchmark results show consistent performance gains
- [ ] Load testing validates performance under scale

### 3.3 Deployment Validation
- [ ] Kubernetes deployment succeeds in multiple environments
- [ ] Health checks pass consistently
- [ ] Service remains available under load
- [ ] Monitoring and alerting function correctly

### 3.4 Compatibility Validation
- [ ] Works with standard protobuf toolchain
- [ ] Compatible with Buf build system
- [ ] Integrates with gRPC ecosystem
- [ ] Supports multiple Go versions

## Implementation Timeline

### Phase 1: Core Infrastructure (Week 1-2)
- Set up repository structure
- Implement basic protobuf definitions
- Create foundational validation tests
- Set up CI/CD pipeline

### Phase 2: Validation Logic (Week 3-4)
- Implement comprehensive type validation
- Create performance benchmarking suite
- Add integration testing framework
- Develop validation service

### Phase 3: Deployment & Monitoring (Week 5-6)
- Create Kubernetes manifests
- Implement monitoring and alerting
- Add load testing capabilities
- Complete documentation

### Phase 4: Production Readiness (Week 7-8)
- Security review and hardening
- Performance optimization
- Comprehensive testing
- Final validation and documentation

## Maintenance Strategy

### Automated Validation
- Daily automated validation runs
- Performance regression detection
- Compatibility testing with new Go versions
- Plugin version compatibility matrix

### Monitoring & Alerting
- Service health monitoring
- Performance metrics tracking
- Error rate and latency monitoring
- Automated incident response

### Documentation & Training
- Comprehensive usage documentation
- Troubleshooting guides
- Performance tuning recommendations
- Team training materials

This specification provides a comprehensive foundation for creating a production-ready validation platform that thoroughly validates the protoc-gen-go-values plugin functionality, performance characteristics, and real-world compatibility.