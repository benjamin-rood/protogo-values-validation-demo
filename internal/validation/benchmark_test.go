package validation

import (
	"fmt"
	"testing"

	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
	"google.golang.org/protobuf/proto"
)

// Phase 2: Comprehensive Performance Benchmarking Suite
// This suite validates R3: Performance Validation requirements:
// - Reduced memory allocations for value slice access
// - Improved cache locality for iteration operations  
// - Measurable performance improvements in benchmarks
// - Zero allocations for slice length operations

const (
	smallDataSize  = 100
	mediumDataSize = 1000
	largeDataSize  = 10000
)

// BenchmarkValueSliceVsPointerSlice compares performance using actual generated types
func BenchmarkValueSliceVsPointerSlice(b *testing.B) {
	// Test with different data sizes to show scalability
	dataSizes := []struct {
		name string
		size int
	}{
		{"Small", smallDataSize},
		{"Medium", mediumDataSize},
		{"Large", largeDataSize},
	}

	for _, ds := range dataSizes {
		b.Run(fmt.Sprintf("DataSize_%s", ds.name), func(b *testing.B) {
			// Benchmark value slice iteration (plugin-generated []DataPoint)
			b.Run("ValueSlice_Iteration", func(b *testing.B) {
				msg := createPerformanceTestMessage(ds.size)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					sum := benchmarkValueSliceIteration(msg.ValueSliceData)
					_ = sum // Prevent compiler optimization
				}
			})

			// Benchmark pointer slice iteration (control group []*Metadata)
			b.Run("PointerSlice_Iteration", func(b *testing.B) {
				msg := createPerformanceTestMessage(ds.size)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					count := benchmarkPointerSliceIteration(msg.PointerSliceData)
					_ = count // Prevent compiler optimization
				}
			})

			// Benchmark value slice random access
			b.Run("ValueSlice_RandomAccess", func(b *testing.B) {
				msg := createPerformanceTestMessage(ds.size)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					index := i % ds.size
					value := msg.ValueSliceData[index].Value
					_ = value // Prevent compiler optimization
				}
			})

			// Benchmark pointer slice random access
			b.Run("PointerSlice_RandomAccess", func(b *testing.B) {
				msg := createPerformanceTestMessage(ds.size)
				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					index := i % len(msg.PointerSliceData)
					if index < len(msg.PointerSliceData) {
						value := msg.PointerSliceData[index].Value
						_ = value // Prevent compiler optimization
					}
				}
			})
		})
	}
}

// BenchmarkZeroAllocationOperations validates zero allocation requirements
func BenchmarkZeroAllocationOperations(b *testing.B) {
	msg := createPerformanceTestMessage(mediumDataSize)

	// Length access should be zero allocations for both types
	b.Run("ValueSlice_Length", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			length := len(msg.ValueSliceData)
			_ = length
		}
	})

	b.Run("PointerSlice_Length", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			length := len(msg.PointerSliceData)
			_ = length
		}
	})

	// Index operations should also be zero allocation
	b.Run("ValueSlice_Index", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			index := i % len(msg.ValueSliceData)
			_ = &msg.ValueSliceData[index] // Taking address should be zero alloc
		}
	})

	b.Run("PointerSlice_Index", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			index := i % len(msg.PointerSliceData)
			_ = msg.PointerSliceData[index] // Already pointer, should be zero alloc
		}
	})
}

// BenchmarkMemoryAllocation compares memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("ValueSlice_Creation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			msg := createPerformanceTestMessage(smallDataSize)
			_ = msg
		}
	})

	b.Run("PointerSlice_Creation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Create equivalent with pointer slices for comparison
			msg := &v1.ValidationTestMessage{
				PointerSliceData: createDataPointPointers(smallDataSize),
			}
			_ = msg
		}
	})
}

// BenchmarkCacheLocality tests cache performance differences
func BenchmarkCacheLocality(b *testing.B) {
	// Large data size to amplify cache effects
	const cacheTestSize = 50000

	b.Run("ValueSlice_Sequential", func(b *testing.B) {
		msg := createPerformanceTestMessage(cacheTestSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Sequential access pattern (cache-friendly)
			sum := float64(0)
			for j := 0; j < len(msg.ValueSliceData); j++ {
				sum += msg.ValueSliceData[j].Value
			}
			_ = sum
		}
	})

	b.Run("PointerSlice_Sequential", func(b *testing.B) {
		pointers := createDataPointPointers(cacheTestSize)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Sequential access pattern (less cache-friendly due to pointer indirection)
			sum := float64(0)
			for j := 0; j < len(pointers); j++ {
				sum += pointers[j].Value
			}
			_ = sum
		}
	})
}

// BenchmarkSerializationPerformance tests protobuf serialization differences
func BenchmarkSerializationPerformance(b *testing.B) {
	valueMsg := createPerformanceTestMessage(mediumDataSize)
	pointerMsg := &v1.ValidationTestMessage{
		PointerSliceData: createDataPointPointers(mediumDataSize),
	}

	b.Run("ValueSlice_Marshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, err := proto.Marshal(valueMsg)
			if err != nil {
				b.Fatal(err)
			}
			_ = data
		}
	})

	b.Run("PointerSlice_Marshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data, err := proto.Marshal(pointerMsg)
			if err != nil {
				b.Fatal(err)
			}
			_ = data
		}
	})
}

// Helper functions for benchmark data creation

func createPerformanceTestMessage(size int) *v1.PerformanceTestMessage {
	// Create value slices (plugin-generated []DataPoint, []ProcessingResult)
	dataPoints := make([]v1.DataPoint, size)
	results := make([]v1.ProcessingResult, size/10)

	for i := 0; i < size; i++ {
		dataPoints[i] = v1.DataPoint{
			Id:        fmt.Sprintf("dp_%d", i),
			Value:     float64(i) * 1.5,
			Timestamp: int64(1000000 + i),
			Tags:      []string{"performance", "benchmark", fmt.Sprintf("item_%d", i)},
		}
	}

	for i := 0; i < size/10; i++ {
		results[i] = v1.ProcessingResult{
			OperationId:   fmt.Sprintf("op_%d", i),
			Success:       i%2 == 0,
			DurationMs:    float64(i) * 0.1,
			ErrorMessages: []string{},
		}
	}

	return &v1.PerformanceTestMessage{
		ValueSliceData:   dataPoints,         // []DataPoint (value slice)
		Results:          results,            // []ProcessingResult (value slice)
		PointerSliceData: createMetadataPointers(size / 20), // []*Metadata (control group)
	}
}

func createDataPointPointers(size int) []*v1.DataPoint {
	pointers := make([]*v1.DataPoint, size)
	for i := 0; i < size; i++ {
		pointers[i] = &v1.DataPoint{
			Id:        fmt.Sprintf("dp_%d", i),
			Value:     float64(i) * 1.5,
			Timestamp: int64(1000000 + i),
			Tags:      []string{"performance", "benchmark", fmt.Sprintf("item_%d", i)},
		}
	}
	return pointers
}

func createMetadataPointers(size int) []*v1.Metadata {
	pointers := make([]*v1.Metadata, size)
	for i := 0; i < size; i++ {
		pointers[i] = &v1.Metadata{
			Key:   fmt.Sprintf("key_%d", i),
			Value: fmt.Sprintf("value_%d", i),
			Attributes: map[string]string{
				"index":     fmt.Sprintf("%d", i),
				"type":      "benchmark",
				"source":    "performance_test",
				"category":  "metadata",
			},
		}
	}
	return pointers
}

// Benchmark iteration helper functions

func benchmarkValueSliceIteration(data []v1.DataPoint) float64 {
	var sum float64
	for _, point := range data {
		sum += point.Value
	}
	return sum
}

func benchmarkPointerSliceIteration(data []*v1.Metadata) int {
	var count int
	for _, meta := range data {
		if meta != nil && len(meta.Key) > 0 {
			count++
		}
	}
	return count
}