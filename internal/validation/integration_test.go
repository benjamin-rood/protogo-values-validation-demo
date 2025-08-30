package validation

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/benjamin-rood/protogo-values-validation-demo/internal/server"
	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// Phase 2: Integration Testing Framework
// This suite validates R4: Integration Testing requirements:
// - Real-world gRPC service operation
// - Successful serialization and deserialization
// - Compatibility with standard protobuf tooling
// - Support for gRPC streaming operations

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// setupTestServer creates an in-memory gRPC server for testing
func setupTestServer() func() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	
	validationServer := server.NewValidationServer()
	v1.RegisterValidationServiceServer(s, validationServer)
	
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	
	return func() {
		s.Stop()
		lis.Close()
	}
}

// bufDialer creates a dialer for the in-memory test server
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// createTestClient creates a gRPC client for testing
func createTestClient(t testing.TB) (v1.ValidationServiceClient, func()) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(bufDialer), 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	
	client := v1.NewValidationServiceClient(conn)
	
	return client, func() { conn.Close() }
}

// TestValidationServiceIntegration tests the complete validation service
func TestValidationServiceIntegration(t *testing.T) {
	cleanup := setupTestServer()
	defer cleanup()
	
	client, closeConn := createTestClient(t)
	defer closeConn()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	t.Run("ValidateTypes_Success", func(t *testing.T) {
		req := &v1.ValidateTypesRequest{
			TestScenarios:  []string{"basic", "performance"},
			DeepValidation: true,
		}
		
		resp, err := client.ValidateTypes(ctx, req)
		if err != nil {
			t.Fatalf("ValidateTypes failed: %v", err)
		}
		
		if !resp.Success {
			t.Error("Expected validation to succeed")
		}
		
		if len(resp.Results) == 0 {
			t.Error("Expected validation results")
		}
		
		// Verify we have both value slices and pointer slices
		if resp.ValueSliceCount == 0 {
			t.Error("Expected some value slices to be found")
		}
		
		if resp.PointerSliceCount == 0 {
			t.Error("Expected some pointer slices to be found")
		}
		
		// Log results for verification
		t.Logf("Validation Results: Success=%v, ValueSlices=%d, PointerSlices=%d", 
			resp.Success, resp.ValueSliceCount, resp.PointerSliceCount)
		
		for i, result := range resp.Results {
			t.Logf("Result %d: %s -> %s (Expected: %s, Passed: %v)", 
				i, result.Scenario, result.ActualType, result.ExpectedType, result.Passed)
			
			if !result.Passed {
				t.Errorf("Validation failed for %s: %s", result.Scenario, result.ErrorMessage)
			}
		}
	})
	
	t.Run("RunBenchmarks_Success", func(t *testing.T) {
		req := &v1.BenchmarkRequest{
			Iterations:     1000,
			DataSize:       100,
			BenchmarkNames: []string{"value_slice", "pointer_slice"},
		}
		
		resp, err := client.RunBenchmarks(ctx, req)
		if err != nil {
			t.Fatalf("RunBenchmarks failed: %v", err)
		}
		
		if !resp.Success {
			t.Error("Expected benchmarks to succeed")
		}
		
		if len(resp.Results) == 0 {
			t.Error("Expected benchmark results")
		}
		
		if resp.Summary == nil {
			t.Error("Expected benchmark summary")
		}
		
		// Log benchmark results
		t.Logf("Benchmark Summary: ValueSlice=%.2fns, PointerSlice=%.2fns, Improvement=%.2fx", 
			resp.Summary.ValueSliceAvgDuration, resp.Summary.PointerSliceAvgDuration, 
			resp.Summary.PerformanceImprovementRatio)
		
		for i, result := range resp.Results {
			t.Logf("Benchmark %d: %s -> %.2fns, %.2f ops/sec", 
				i, result.Name, result.DurationNs, result.OperationsPerSecond)
		}
	})
	
	t.Run("RunBenchmarks_InvalidParameters", func(t *testing.T) {
		req := &v1.BenchmarkRequest{
			Iterations: -1, // Invalid
			DataSize:   100,
		}
		
		_, err := client.RunBenchmarks(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid iterations")
		}
	})
}

// TestStreamingValidation tests the streaming validation functionality
func TestStreamingValidation(t *testing.T) {
	cleanup := setupTestServer()
	defer cleanup()
	
	client, closeConn := createTestClient(t)
	defer closeConn()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	stream, err := client.StreamValidation(ctx)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}
	
	// Send test requests
	numRequests := 5
	for i := 0; i < numRequests; i++ {
		req := &v1.StreamRequest{
			RequestId:      fmt.Sprintf("req_%d", i),
			SequenceNumber: int32(i),
			TestData: &v1.ValidationTestMessage{
				ValueSliceData: []v1.DataPoint{
					{Id: fmt.Sprintf("dp_%d", i), Value: float64(i), Timestamp: int64(i)},
				},
				PointerSliceData: []*v1.DataPoint{
					{Id: fmt.Sprintf("ptr_%d", i), Value: float64(i*2), Timestamp: int64(i*2)},
				},
			},
		}
		
		if err := stream.Send(req); err != nil {
			t.Fatalf("Failed to send request %d: %v", i, err)
		}
	}
	
	// Close the sending side
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("Failed to close send: %v", err)
	}
	
	// Receive responses
	receivedCount := 0
	for {
		resp, err := stream.Recv()
		if err != nil {
			// End of stream
			break
		}
		
		if !resp.Success {
			t.Errorf("Stream validation failed for request %s: %s", 
				resp.RequestId, resp.Message)
		}
		
		if resp.Stats == nil {
			t.Error("Expected processing stats in response")
		}
		
		t.Logf("Stream Response: RequestId=%s, Success=%v, ItemsProcessed=%d, Throughput=%.2f", 
			resp.RequestId, resp.Success, resp.Stats.ItemsProcessed, resp.Stats.Throughput)
		
		receivedCount++
	}
	
	if receivedCount != numRequests {
		t.Errorf("Expected %d responses, got %d", numRequests, receivedCount)
	}
}

// TestProtobufCompatibility tests protobuf serialization/deserialization
func TestProtobufCompatibility(t *testing.T) {
	cleanup := setupTestServer()
	defer cleanup()
	
	client, closeConn := createTestClient(t)
	defer closeConn()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	t.Run("MessageSerialization", func(t *testing.T) {
		// Create a complex test message
		original := &v1.ValidateTypesRequest{
			TestScenarios:  []string{"test1", "test2", "test3"},
			DeepValidation: true,
		}
		
		// Send it through the service (which will serialize/deserialize it)
		resp, err := client.ValidateTypes(ctx, original)
		if err != nil {
			t.Fatalf("Failed to call service: %v", err)
		}
		
		// Verify response structure is valid
		if resp == nil {
			t.Fatal("Response is nil")
		}
		
		if len(resp.Results) == 0 {
			t.Error("Expected results in response")
		}
		
		t.Logf("Serialization test passed: got %d results", len(resp.Results))
	})
	
	t.Run("ComplexMessageTypes", func(t *testing.T) {
		// Test streaming with complex nested messages
		stream, err := client.StreamValidation(ctx)
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}
		
		// Create message with all field types
		complexMsg := &v1.ValidationTestMessage{
			ValueSliceData: []v1.DataPoint{
				{
					Id:        "complex_test",
					Value:     123.456,
					Timestamp: time.Now().Unix(),
					Tags:      []string{"integration", "test", "complex"},
				},
			},
			PointerSliceData: []*v1.DataPoint{
				{
					Id:        "pointer_test", 
					Value:     789.012,
					Timestamp: time.Now().Unix(),
					Tags:      []string{"pointer", "test"},
				},
			},
			Metrics: []v1.MetricPoint{
				{
					Name:        "test_metric",
					Measurement: 0.95,
					Labels:      map[string]string{"env": "test", "type": "integration"},
				},
			},
		}
		
		req := &v1.StreamRequest{
			RequestId:      "complex_test",
			SequenceNumber: 1,
			TestData:       complexMsg,
		}
		
		if err := stream.Send(req); err != nil {
			t.Fatalf("Failed to send complex message: %v", err)
		}
		
		if err := stream.CloseSend(); err != nil {
			t.Fatalf("Failed to close send: %v", err)
		}
		
		// Receive response
		resp, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to receive response: %v", err)
		}
		
		if !resp.Success {
			t.Error("Complex message validation failed")
		}
		
		if resp.Stats.ItemsProcessed != 2 { // 1 value slice + 1 pointer slice
			t.Errorf("Expected 2 items processed, got %d", resp.Stats.ItemsProcessed)
		}
		
		t.Logf("Complex message test passed: %s", resp.Message)
	})
}

// TestConcurrentAccess tests concurrent access to the service
func TestConcurrentAccess(t *testing.T) {
	cleanup := setupTestServer()
	defer cleanup()
	
	client, closeConn := createTestClient(t)
	defer closeConn()
	
	// Run multiple concurrent requests
	numWorkers := 10
	results := make(chan error, numWorkers)
	
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			req := &v1.ValidateTypesRequest{
				TestScenarios:  []string{fmt.Sprintf("worker_%d", workerID)},
				DeepValidation: false,
			}
			
			resp, err := client.ValidateTypes(ctx, req)
			if err != nil {
				results <- fmt.Errorf("worker %d failed: %v", workerID, err)
				return
			}
			
			if !resp.Success {
				results <- fmt.Errorf("worker %d validation failed", workerID)
				return
			}
			
			results <- nil
		}(i)
	}
	
	// Collect results
	for i := 0; i < numWorkers; i++ {
		if err := <-results; err != nil {
			t.Error(err)
		}
	}
	
	t.Logf("Concurrent access test completed with %d workers", numWorkers)
}

// BenchmarkServicePerformance benchmarks the service under load
func BenchmarkServicePerformance(b *testing.B) {
	cleanup := setupTestServer()
	defer cleanup()
	
	client, closeConn := createTestClient(b)
	defer closeConn()
	
	req := &v1.ValidateTypesRequest{
		TestScenarios:  []string{"performance"},
		DeepValidation: false,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_, err := client.ValidateTypes(ctx, req)
			cancel()
			
			if err != nil {
				b.Fatalf("Service call failed: %v", err)
			}
		}
	})
}