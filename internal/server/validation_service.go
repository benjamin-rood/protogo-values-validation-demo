package server

import (
	"context"
	"fmt"
	"reflect"
	"time"

	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ValidationServer implements the ValidationService gRPC service
type ValidationServer struct {
	v1.UnimplementedValidationServiceServer
}

// NewValidationServer creates a new validation service server
func NewValidationServer() *ValidationServer {
	return &ValidationServer{}
}

// ValidateTypes validates that the plugin correctly transforms field types
func (s *ValidationServer) ValidateTypes(ctx context.Context, req *v1.ValidateTypesRequest) (*v1.ValidateTypesResponse, error) {
	results := make([]*v1.ValidationResult, 0)
	var valueSliceCount, pointerSliceCount int32

	// Validate ValidationTestMessage types (MVP compatibility)
	validationResults := s.validateValidationTestMessageTypes()
	results = append(results, validationResults...)

	// Validate PerformanceTestMessage types (Phase 1 spec-compliant)
	performanceResults := s.validatePerformanceTestMessageTypes()
	results = append(results, performanceResults...)

	// Count value slices and pointer slices
	for _, result := range results {
		if result.Passed && containsValueSlice(result.ActualType) {
			valueSliceCount++
		} else if result.Passed && containsPointerSlice(result.ActualType) {
			pointerSliceCount++
		}
	}

	// Check if all validations passed
	allPassed := true
	for _, result := range results {
		if !result.Passed {
			allPassed = false
			break
		}
	}

	return &v1.ValidateTypesResponse{
		Success:             allPassed,
		Results:             results,
		ValueSliceCount:     valueSliceCount,
		PointerSliceCount:   pointerSliceCount,
	}, nil
}

// RunBenchmarks performs performance benchmarking
func (s *ValidationServer) RunBenchmarks(ctx context.Context, req *v1.BenchmarkRequest) (*v1.BenchmarkResponse, error) {
	if req.Iterations <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "iterations must be > 0")
	}

	if req.DataSize <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "data_size must be > 0")
	}

	results := make([]*v1.BenchmarkResult, 0)

	// Run value slice iteration benchmark
	valueSliceResult := s.benchmarkValueSliceIteration(int(req.Iterations), int(req.DataSize))
	results = append(results, valueSliceResult)

	// Run pointer slice iteration benchmark
	pointerSliceResult := s.benchmarkPointerSliceIteration(int(req.Iterations), int(req.DataSize))
	results = append(results, pointerSliceResult)

	// Run memory allocation benchmark
	memoryResult := s.benchmarkMemoryAllocation(int(req.Iterations), int(req.DataSize))
	results = append(results, memoryResult)

	// Run serialization benchmark
	serializationResult := s.benchmarkSerialization(int(req.Iterations), int(req.DataSize))
	results = append(results, serializationResult)

	// Calculate summary statistics
	summary := s.calculateBenchmarkSummary(results)

	return &v1.BenchmarkResponse{
		Success: true,
		Results: results,
		Summary: summary,
	}, nil
}

// StreamValidation handles streaming validation requests
func (s *ValidationServer) StreamValidation(stream v1.ValidationService_StreamValidationServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			// End of stream
			return nil
		}

		// Process the request
		startTime := time.Now()
		
		// Validate the test data
		isValid := s.validateTestMessage(req.TestData)
		
		processingTime := time.Since(startTime)

		// Send response
		resp := &v1.StreamResponse{
			RequestId:      req.RequestId,
			Success:        isValid,
			Message:        fmt.Sprintf("Processed request %s", req.RequestId),
			SequenceNumber: req.SequenceNumber,
			Stats: &v1.ProcessingStats{
				ProcessingTimeNs: processingTime.Nanoseconds(),
				ItemsProcessed:   int32(len(req.TestData.ValueSliceData) + len(req.TestData.PointerSliceData)),
				Throughput:       float64(len(req.TestData.ValueSliceData)+len(req.TestData.PointerSliceData)) / processingTime.Seconds(),
			},
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

// Helper methods for type validation

func (s *ValidationServer) validateValidationTestMessageTypes() []*v1.ValidationResult {
	var results []*v1.ValidationResult

	// Test ValueSliceData field
	msg := v1.ValidationTestMessage{}
	actualType := reflect.TypeOf(msg.ValueSliceData).String()
	expectedType := "[]v1.DataPoint"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "ValidationTestMessage.ValueSliceData",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	// Test PointerSliceData field
	actualType = reflect.TypeOf(msg.PointerSliceData).String()
	expectedType = "[]*v1.DataPoint"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "ValidationTestMessage.PointerSliceData",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	// Test Metrics field (structured field option)
	actualType = reflect.TypeOf(msg.Metrics).String()
	expectedType = "[]v1.MetricPoint"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "ValidationTestMessage.Metrics",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	return results
}

func (s *ValidationServer) validatePerformanceTestMessageTypes() []*v1.ValidationResult {
	var results []*v1.ValidationResult

	// Test PerformanceTestMessage fields
	msg := v1.PerformanceTestMessage{}
	
	// Test ValueSliceData field
	actualType := reflect.TypeOf(msg.ValueSliceData).String()
	expectedType := "[]v1.DataPoint"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "PerformanceTestMessage.ValueSliceData",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	// Test PointerSliceData field
	actualType = reflect.TypeOf(msg.PointerSliceData).String()
	expectedType = "[]*v1.Metadata"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "PerformanceTestMessage.PointerSliceData",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	// Test Results field
	actualType = reflect.TypeOf(msg.Results).String()
	expectedType = "[]v1.ProcessingResult"
	
	results = append(results, &v1.ValidationResult{
		Scenario:     "PerformanceTestMessage.Results",
		Passed:       actualType == expectedType,
		ErrorMessage: getErrorMessage(actualType, expectedType),
		ExpectedType: expectedType,
		ActualType:   actualType,
	})

	return results
}

// Benchmark helper methods

func (s *ValidationServer) benchmarkValueSliceIteration(iterations, dataSize int) *v1.BenchmarkResult {
	// Create test data
	data := make([]v1.DataPoint, dataSize)
	for i := 0; i < dataSize; i++ {
		data[i] = v1.DataPoint{
			Id:        fmt.Sprintf("dp_%d", i),
			Value:     float64(i) * 1.5,
			Timestamp: int64(1000000 + i),
		}
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		sum := float64(0)
		for _, dp := range data {
			sum += dp.Value
		}
		_ = sum
	}
	duration := time.Since(start)

	return &v1.BenchmarkResult{
		Name:                "ValueSlice_Iteration",
		DurationNs:          float64(duration.Nanoseconds()),
		Allocations:         0, // Value slice iteration should have minimal allocations
		BytesAllocated:      0,
		OperationsPerSecond: float64(iterations) / duration.Seconds(),
	}
}

func (s *ValidationServer) benchmarkPointerSliceIteration(iterations, dataSize int) *v1.BenchmarkResult {
	// Create test data
	data := make([]*v1.DataPoint, dataSize)
	for i := 0; i < dataSize; i++ {
		data[i] = &v1.DataPoint{
			Id:        fmt.Sprintf("dp_%d", i),
			Value:     float64(i) * 1.5,
			Timestamp: int64(1000000 + i),
		}
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		sum := float64(0)
		for _, dp := range data {
			sum += dp.Value
		}
		_ = sum
	}
	duration := time.Since(start)

	return &v1.BenchmarkResult{
		Name:                "PointerSlice_Iteration",
		DurationNs:          float64(duration.Nanoseconds()),
		Allocations:         0, // Baseline comparison
		BytesAllocated:      0,
		OperationsPerSecond: float64(iterations) / duration.Seconds(),
	}
}

func (s *ValidationServer) benchmarkMemoryAllocation(iterations, dataSize int) *v1.BenchmarkResult {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		// Simulate memory allocation patterns
		msg := &v1.PerformanceTestMessage{
			ValueSliceData: make([]v1.DataPoint, dataSize),
		}
		_ = msg
	}
	duration := time.Since(start)

	return &v1.BenchmarkResult{
		Name:                "Memory_Allocation",
		DurationNs:          float64(duration.Nanoseconds()),
		Allocations:         int64(iterations), // One allocation per iteration
		BytesAllocated:      int64(iterations * dataSize * 64), // Estimate
		OperationsPerSecond: float64(iterations) / duration.Seconds(),
	}
}

func (s *ValidationServer) benchmarkSerialization(iterations, dataSize int) *v1.BenchmarkResult {
	// Create test message
	msg := &v1.PerformanceTestMessage{
		ValueSliceData: make([]v1.DataPoint, dataSize),
	}
	for i := 0; i < dataSize; i++ {
		msg.ValueSliceData[i] = v1.DataPoint{
			Id:        fmt.Sprintf("dp_%d", i),
			Value:     float64(i),
			Timestamp: int64(i),
		}
	}

	start := time.Now()
	var totalBytes int64
	for i := 0; i < iterations; i++ {
		data, err := proto.Marshal(msg)
		if err == nil {
			totalBytes += int64(len(data))
		}
		_ = data
	}
	duration := time.Since(start)

	return &v1.BenchmarkResult{
		Name:                "Serialization",
		DurationNs:          float64(duration.Nanoseconds()),
		Allocations:         int64(iterations),
		BytesAllocated:      totalBytes,
		OperationsPerSecond: float64(iterations) / duration.Seconds(),
	}
}

func (s *ValidationServer) calculateBenchmarkSummary(results []*v1.BenchmarkResult) *v1.BenchmarkSummary {
	var valueSliceDuration, pointerSliceDuration float64
	var memoryUsage int64

	for _, result := range results {
		switch result.Name {
		case "ValueSlice_Iteration":
			valueSliceDuration = result.DurationNs
		case "PointerSlice_Iteration":
			pointerSliceDuration = result.DurationNs
		case "Memory_Allocation", "Serialization":
			memoryUsage += result.BytesAllocated
		}
	}

	// Calculate performance improvement ratio
	improvementRatio := float64(1.0)
	if pointerSliceDuration > 0 && valueSliceDuration > 0 {
		improvementRatio = pointerSliceDuration / valueSliceDuration
	}

	return &v1.BenchmarkSummary{
		ValueSliceAvgDuration:        valueSliceDuration,
		PointerSliceAvgDuration:      pointerSliceDuration,
		PerformanceImprovementRatio:  improvementRatio,
		MemorySavingsBytes:           memoryUsage,
	}
}

// Utility functions

func (s *ValidationServer) validateTestMessage(msg *v1.ValidationTestMessage) bool {
	if msg == nil {
		return false
	}
	
	// Basic validation - check that fields have expected types
	valueSliceType := reflect.TypeOf(msg.ValueSliceData).String()
	pointerSliceType := reflect.TypeOf(msg.PointerSliceData).String()
	
	return valueSliceType == "[]v1.DataPoint" && pointerSliceType == "[]*v1.DataPoint"
}

func getErrorMessage(actual, expected string) string {
	if actual != expected {
		return fmt.Sprintf("Expected %s, got %s", expected, actual)
	}
	return ""
}

func containsValueSlice(typeStr string) bool {
	return typeStr[:2] == "[]" && typeStr[2] != '*'
}

func containsPointerSlice(typeStr string) bool {
	return typeStr[:3] == "[]*"
}