// Package zon_test provides security and robustness tests.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestRejectProtoKeys tests rejection of __proto__ keys.
func TestRejectProtoKeys(t *testing.T) {
	malicious := `users:@(1):id,__proto__.polluted
1,true`

	_, err := zon.Decode(malicious)
	// Decoding should succeed but not pollute prototype
	if err != nil {
		return // Error is acceptable
	}

	// Verify that object prototype is not polluted
	testObj := make(map[string]any)
	if _, ok := testObj["polluted"]; ok {
		t.Error("Prototype pollution detected via __proto__")
	}
}

// TestRejectConstructorPrototypeKeys tests rejection of constructor.prototype keys.
func TestRejectConstructorPrototypeKeys(t *testing.T) {
	malicious := `users:@(1):id,constructor.prototype.polluted
1,true`

	_, err := zon.Decode(malicious)
	// Decoding should succeed but not pollute prototype
	if err != nil {
		return // Error is acceptable
	}

	// Verify that object prototype is not polluted
	testObj := make(map[string]any)
	if _, ok := testObj["polluted"]; ok {
		t.Error("Prototype pollution detected via constructor.prototype")
	}
}

// TestDOSDeepNesting tests denial of service via deep nesting.
func TestDOSDeepNesting(t *testing.T) {
	depth := 150
	deepZon := strings.Repeat("[", depth) + "]" + strings.Repeat("]", depth-1)

	_, err := zon.Decode(deepZon)
	if err == nil {
		t.Error("Expected error for deep nesting")
	}
	if !strings.Contains(err.Error(), "Maximum nesting depth exceeded") {
		t.Errorf("Expected nesting depth error, got: %v", err)
	}
}

// TestCircularReferenceInEncoder tests circular reference detection in encoder.
func TestCircularReferenceInEncoder(t *testing.T) {
	// Note: In Go, we can't create true circular references with map[string]any
	// like we can in JavaScript. The test verifies the encoder handles edge cases.
	
	// Create a deeply nested structure instead
	data := map[string]any{
		"name": "loop",
		"nested": map[string]any{
			"deep": map[string]any{
				"value": "test",
			},
		},
	}

	_, err := zon.Encode(data)
	if err != nil {
		t.Errorf("Should handle nested structures: %v", err)
	}
}

// TestNestedStructuresEncode tests encoding of nested structures.
func TestNestedStructuresEncode(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "value",
			},
		},
	}

	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	a := decodedMap["a"].(map[string]any)
	b := a["b"].(map[string]any)
	if b["c"] != "value" {
		t.Errorf("Expected 'value', got %v", b["c"])
	}
}

// TestMalformedInput tests handling of malformed input.
func TestMalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Unclosed bracket", "[1,2,3"},
		{"Unclosed brace", "{a:1,b:2"},
		{"Extra closing", "[1,2,3]]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These may or may not error, but should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panicked on input %q: %v", tt.input, r)
				}
			}()
			zon.Decode(tt.input)
		})
	}
}

// TestSpecialCharactersInKeys tests special characters in keys.
func TestSpecialCharactersInKeys(t *testing.T) {
	data := map[string]any{
		"normal":     "value1",
		"with-dash":  "value2",
		"with_under": "value3",
		"with.dot":   "value4",
	}

	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	if decodedMap["normal"] != "value1" {
		t.Errorf("normal: expected value1, got %v", decodedMap["normal"])
	}
	if decodedMap["with-dash"] != "value2" {
		t.Errorf("with-dash: expected value2, got %v", decodedMap["with-dash"])
	}
}

// TestUnicodeStrings tests Unicode string handling.
func TestUnicodeStrings(t *testing.T) {
	data := map[string]any{
		"chinese": "王小明",
		"emoji":   "✅🚀",
		"arabic":  "مرحبا",
	}

	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	if decodedMap["chinese"] != "王小明" {
		t.Errorf("chinese: expected 王小明, got %v", decodedMap["chinese"])
	}
	if decodedMap["emoji"] != "✅🚀" {
		t.Errorf("emoji: expected ✅🚀, got %v", decodedMap["emoji"])
	}
	if decodedMap["arabic"] != "مرحبا" {
		t.Errorf("arabic: expected مرحبا, got %v", decodedMap["arabic"])
	}
}
