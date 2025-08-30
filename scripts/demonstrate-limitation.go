package main

import (
	"fmt"
	"reflect"

	v1 "github.com/benjamin-rood/protogo-values-validation-demo/gen/api/validation/v1"
	"google.golang.org/protobuf/proto"
)

// Demonstration script showing the plugin's correct behavior and fundamental limitation
func main() {
	fmt.Println("🔍 protoc-gen-go-values Plugin Behavior Demonstration")
	fmt.Println("====================================================")
	fmt.Println()

	demonstratePluginCorrectness()
	fmt.Println()
	demonstrateMarshalingFailure()
}

func demonstratePluginCorrectness() {
	fmt.Println("📋 Plugin Behavior: WORKING CORRECTLY")
	fmt.Println("-------------------------------------")

	// Check ValidationTestMessage types
	msg := v1.ValidationTestMessage{}
	
	fmt.Println("ValidationTestMessage field transformations:")
	fmt.Printf("  ValueSliceData:   %s", reflect.TypeOf(msg.ValueSliceData))
	if reflect.TypeOf(msg.ValueSliceData).String() == "[]v1.DataPoint" {
		fmt.Println(" ✅ (has field option → transformed)")
	} else {
		fmt.Println(" ❌ (should be transformed)")
	}
	
	fmt.Printf("  PointerSliceData: %s", reflect.TypeOf(msg.PointerSliceData))
	if reflect.TypeOf(msg.PointerSliceData).String() == "[]*v1.DataPoint" {
		fmt.Println(" ✅ (no field option → unchanged)")
	} else {
		fmt.Println(" ❌ (should remain unchanged)")
	}
	
	fmt.Printf("  Metrics:          %s", reflect.TypeOf(msg.Metrics))
	if reflect.TypeOf(msg.Metrics).String() == "[]v1.MetricPoint" {
		fmt.Println(" ✅ (has field option → transformed)")
	} else {
		fmt.Println(" ❌ (should be transformed)")
	}

	// Check validation service types (should be untransformed)
	response := v1.ValidateTypesResponse{}
	fmt.Printf("  ValidateTypesResponse.Results: %s", reflect.TypeOf(response.Results))
	if reflect.TypeOf(response.Results).String() == "[]*v1.ValidationResult" {
		fmt.Println(" ✅ (no field option → unchanged)")
	} else {
		fmt.Println(" ❌ (should remain unchanged)")
	}
	
	fmt.Println()
	fmt.Println("🎯 Plugin correctly transforms ONLY fields with explicit options")
}

func demonstrateMarshalingFailure() {
	fmt.Println("📋 Runtime Marshaling: CRITICAL FAILURE")
	fmt.Println("---------------------------------------")

	// Create a message with value slices (transformed fields)
	msg := &v1.ValidationTestMessage{
		ValueSliceData: []v1.DataPoint{
			{
				Id:        "test1", 
				Value:     42.0,
				Timestamp: 1234567890,
				Tags:      []string{"demo"},
			},
		},
		Metrics: []v1.MetricPoint{
			{
				Name:        "test_metric",
				Measurement: 0.95,
				Labels:      map[string]string{"type": "demo"},
			},
		},
	}

	fmt.Println("Attempting to marshal ValidationTestMessage with value slices...")
	fmt.Println()

	// This will panic because protobuf marshaler expects pointer slices
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ PANIC (as expected): %v\n", r)
			fmt.Println()
			fmt.Println("🔍 Root Cause Analysis:")
			fmt.Println("   - Protobuf reflection calls .Elem() on slice types")
			fmt.Println("   - Expects pointers that can be dereferenced: []*Type")  
			fmt.Println("   - Value slices don't support .Elem(): []Type")
			fmt.Println("   - Runtime panic: 'reflect: Elem of invalid type'")
			fmt.Println()
			fmt.Println("🚨 ARCHITECTURAL INCOMPATIBILITY:")
			fmt.Println("   Value slices for message types fundamentally incompatible with protobuf")
			fmt.Println()
			fmt.Println("📚 This validates the plugin project discontinuation decision")
		}
	}()

	// This line will cause a panic
	_, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
	} else {
		fmt.Println("⚠️ Unexpected: Marshaling should have panicked")
	}
}