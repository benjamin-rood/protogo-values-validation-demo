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
		getActualType func() interface{}
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
			name:         "metrics should be value slice (structured option working)",
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
