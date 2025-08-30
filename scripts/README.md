# Validation Scripts

These scripts demonstrate the protoc-gen-go-values plugin behavior and validate the architectural limitations discovered during development.

## Scripts Overview

### `validate-plugin.sh`
**Comprehensive validation script** that tests the plugin through multiple phases:

```bash
./scripts/validate-plugin.sh
```

**Phases:**
1. **Code Generation**: Verifies buf generate works and code compiles
2. **Plugin Behavior**: Confirms only fields with options are transformed  
3. **Runtime Marshaling**: Demonstrates the marshaling panic with value slices
4. **Summary**: Provides conclusion about plugin limitations

### `demonstrate-limitation.go`
**Interactive demonstration** showing plugin correctness and marshaling failure:

```bash
go run scripts/demonstrate-limitation.go
```

**Features:**
- Inspects generated types using Go reflection
- Validates plugin transformation logic
- Triggers the marshaling panic in a controlled way
- Provides detailed root cause analysis

## Expected Output

### Plugin Behavior (Working Correctly)
```
ValidationTestMessage field transformations:
  ValueSliceData:   []v1.DataPoint ✅ (has field option → transformed)
  PointerSliceData: []*v1.DataPoint ✅ (no field option → unchanged)
  Metrics:          []v1.MetricPoint ✅ (has field option → transformed)
  ValidateTypesResponse.Results: []*v1.ValidationResult ✅ (no field option → unchanged)
```

### Runtime Marshaling (Critical Failure)
```
❌ PANIC (as expected): reflect: Elem of invalid type v1.DataPoint

🔍 Root Cause Analysis:
   - Protobuf reflection calls .Elem() on slice types
   - Expects pointers that can be dereferenced: []*Type
   - Value slices don't support .Elem(): []Type
   - Runtime panic: 'reflect: Elem of invalid type'
```

## Educational Purpose

These scripts serve as **educational case studies** demonstrating:

1. **Plugin Implementation Success**: The plugin works exactly as designed
2. **Architectural Incompatibility**: Protobuf marshaling requires pointer slices for message types
3. **System Constraints**: Sometimes implementation success reveals deeper architectural limitations

## Integration with Discovery

These validation scripts complement the comprehensive analysis in:
- `../protogo-values/README.md` - Complete project failure analysis
- `../protogo-values/specs/plugin-transformation-bug.spec.md` - Detailed bug investigation  
- Integration tests in `internal/validation/` - Real-world gRPC failure scenarios

## Running the Scripts

### Prerequisites
- Go 1.24+
- buf CLI
- Generated protobuf code (run `buf generate` first)

### Quick Validation
```bash
# Full validation suite
./scripts/validate-plugin.sh

# Interactive demonstration  
go run scripts/demonstrate-limitation.go

# Integration tests (will panic as expected)
go test -v ./internal/validation
```

## Key Insights

1. **Plugin Correctness**: The plugin implementation is technically sound and works as specified
2. **Fundamental Limitation**: Protobuf's architecture cannot support value slices for message types
3. **Discovery Process**: Comprehensive testing revealed limitations that weren't apparent during initial design
4. **Architectural Coupling**: Protobuf's type system and marshaling are tightly integrated - you can't change one without the other

These scripts validate the decision to discontinue the plugin project due to insurmountable architectural constraints.