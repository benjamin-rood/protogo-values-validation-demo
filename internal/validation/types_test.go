package validation

import (
	"reflect"
	"testing"

	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
)

func TestPluginTypeTransformation(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		expectedType  string
		getActualType func() any
	}{
		{
			name:         "value_slice_data should be value slice",
			fieldName:    "ValueSliceData",
			expectedType: "[]v1.DataPoint",
			getActualType: func() any {
				return v1.ValidationTestMessage{}.ValueSliceData
			},
		},
		{
			name:         "pointer_slice_data should remain pointer slice",
			fieldName:    "PointerSliceData",
			expectedType: "[]*v1.DataPoint",
			getActualType: func() any {
				return v1.ValidationTestMessage{}.PointerSliceData
			},
		},
		{
			name:         "metrics should be value slice (structured option working)",
			fieldName:    "Metrics",
			expectedType: "[]v1.MetricPoint",
			getActualType: func() any {
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

func TestPerformanceTestMessageTypes(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		expectedType  string
		getActualType func() any
	}{
		{
			name:         "PerformanceTestMessage.ValueSliceData should be value slice",
			fieldName:    "ValueSliceData",
			expectedType: "[]v1.DataPoint",
			getActualType: func() any {
				return v1.PerformanceTestMessage{}.ValueSliceData
			},
		},
		{
			name:         "PerformanceTestMessage.PointerSliceData should be pointer slice (Metadata)",
			fieldName:    "PointerSliceData",
			expectedType: "[]*v1.Metadata",
			getActualType: func() any {
				return v1.PerformanceTestMessage{}.PointerSliceData
			},
		},
		{
			name:         "PerformanceTestMessage.Results should be value slice",
			fieldName:    "Results",
			expectedType: "[]v1.ProcessingResult",
			getActualType: func() any {
				return v1.PerformanceTestMessage{}.Results
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

func TestPerformanceTestMessageIntegration(t *testing.T) {
	// Verify PerformanceTestMessage can be created with all field types
	msg := &v1.PerformanceTestMessage{
		ValueSliceData: []v1.DataPoint{
			{Id: "perf1", Value: 10.5, Timestamp: 1000, Tags: []string{"performance"}},
		},
		PointerSliceData: []*v1.Metadata{
			{Key: "environment", Value: "test", Attributes: map[string]string{"version": "1.0"}},
		},
		Results: []v1.ProcessingResult{
			{OperationId: "op1", Success: true, DurationMs: 15.5, ErrorMessages: []string{}},
		},
	}

	if len(msg.ValueSliceData) != 1 {
		t.Errorf("Expected 1 value slice item, got %d", len(msg.ValueSliceData))
	}

	if len(msg.PointerSliceData) != 1 {
		t.Errorf("Expected 1 pointer slice item, got %d", len(msg.PointerSliceData))
	}

	if len(msg.Results) != 1 {
		t.Errorf("Expected 1 result item, got %d", len(msg.Results))
	}

	// Verify we can access nested fields
	if msg.ValueSliceData[0].Tags[0] != "performance" {
		t.Errorf("Expected tag 'performance', got %s", msg.ValueSliceData[0].Tags[0])
	}

	if msg.PointerSliceData[0].Attributes["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got %s", msg.PointerSliceData[0].Attributes["version"])
	}

	if !msg.Results[0].Success {
		t.Error("Expected operation to be successful")
	}
}
