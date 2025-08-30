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
			Id:        "test",
			Value:     float64(i),
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
			Id:        "test",
			Value:     float64(i),
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