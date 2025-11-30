// Package zon_test provides tests for canonical number formatting.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"math"
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestIntegerWithoutDecimal tests that integers are encoded without decimal point.
func TestIntegerWithoutDecimal(t *testing.T) {
	data := map[string]any{"value": float64(42)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "42") {
		t.Errorf("Expected 42 in output, got: %s", encoded)
	}
	if strings.Contains(encoded, "42.0") {
		t.Errorf("Should not contain 42.0, got: %s", encoded)
	}
}

// TestZero tests zero handling.
func TestZero(t *testing.T) {
	data := map[string]any{"value": float64(0)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "value:0") {
		t.Errorf("Expected value:0, got: %s", encoded)
	}
}

// TestNegativeIntegers tests negative integer handling.
func TestNegativeIntegers(t *testing.T) {
	data := map[string]any{"value": float64(-123)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "-123") {
		t.Errorf("Expected -123, got: %s", encoded)
	}
}

// TestFloatsWithoutTrailingZeros tests floats without trailing zeros.
func TestFloatsWithoutTrailingZeros(t *testing.T) {
	data := map[string]any{"value": 3.14}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "3.14") {
		t.Errorf("Expected 3.14, got: %s", encoded)
	}
	if strings.Contains(encoded, "3.140000") {
		t.Errorf("Should not have trailing zeros, got: %s", encoded)
	}
}

// TestVerySmallDecimals tests very small decimal handling.
func TestVerySmallDecimals(t *testing.T) {
	data := map[string]any{"value": 0.001}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "0.001") {
		t.Errorf("Expected 0.001, got: %s", encoded)
	}
	if strings.Contains(encoded, "1e-3") {
		t.Errorf("Should not use scientific notation, got: %s", encoded)
	}
}

// TestNoScientificNotationForLargeNumbers tests no scientific notation for large numbers.
func TestNoScientificNotationForLargeNumbers(t *testing.T) {
	data := map[string]any{"value": float64(1000000)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "1000000") {
		t.Errorf("Expected 1000000, got: %s", encoded)
	}
	if strings.Contains(encoded, "1e6") || strings.Contains(encoded, "1e+6") {
		t.Errorf("Should not use scientific notation, got: %s", encoded)
	}
}

// TestManyDecimalPlaces tests numbers with many decimal places.
func TestManyDecimalPlaces(t *testing.T) {
	data := map[string]any{"value": 3.141592653589793}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Should preserve precision
	if !strings.Contains(encoded, "3.14159265358979") {
		t.Errorf("Expected precision preserved, got: %s", encoded)
	}
	// Should not contain scientific notation
	if strings.Contains(encoded, "e+") || strings.Contains(encoded, "e-") || strings.Contains(encoded, "E") {
		t.Errorf("Should not use scientific notation, got: %s", encoded)
	}
}

// TestNaNAsNull tests NaN encoded as null.
func TestNaNAsNull(t *testing.T) {
	data := map[string]any{"value": math.NaN()}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "value:null") {
		t.Errorf("Expected value:null, got: %s", encoded)
	}
}

// TestInfinityAsNull tests Infinity encoded as null.
func TestInfinityAsNull(t *testing.T) {
	data := map[string]any{"value": math.Inf(1)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "value:null") {
		t.Errorf("Expected value:null, got: %s", encoded)
	}
}

// TestNegativeInfinityAsNull tests -Infinity encoded as null.
func TestNegativeInfinityAsNull(t *testing.T) {
	data := map[string]any{"value": math.Inf(-1)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "value:null") {
		t.Errorf("Expected value:null, got: %s", encoded)
	}
}

// TestIntegerRoundTrip tests integer values through round-trip.
func TestIntegerRoundTrip(t *testing.T) {
	data := map[string]any{"value": float64(42)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	if decodedMap["value"] != float64(42) {
		t.Errorf("Expected 42, got %v", decodedMap["value"])
	}
	// Check it's an integer
	val := decodedMap["value"].(float64)
	if val != float64(int(val)) {
		t.Errorf("Expected integer value")
	}
}

// TestFloatRoundTrip tests float values through round-trip.
func TestFloatRoundTrip(t *testing.T) {
	data := map[string]any{"value": 3.14}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	val := decodedMap["value"].(float64)
	if math.Abs(val-3.14) > 0.0000001 {
		t.Errorf("Expected 3.14, got %v", val)
	}
}

// TestLargeNumberRoundTrip tests large numbers through round-trip.
func TestLargeNumberRoundTrip(t *testing.T) {
	data := map[string]any{"value": float64(1000000)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	if decodedMap["value"] != float64(1000000) {
		t.Errorf("Expected 1000000, got %v", decodedMap["value"])
	}
}

// TestVerySmallNumberRoundTrip tests very small numbers through round-trip.
func TestVerySmallNumberRoundTrip(t *testing.T) {
	data := map[string]any{"value": 0.000001}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decodedMap := decoded.(map[string]any)
	val := decodedMap["value"].(float64)
	if math.Abs(val-0.000001) > 0.0000000001 {
		t.Errorf("Expected 0.000001, got %v", val)
	}
}

// TestArrayOfNumbers tests canonical number formatting in arrays.
func TestArrayOfNumbers(t *testing.T) {
	data := map[string]any{
		"values": []any{
			map[string]any{"num": float64(1000000)},
			map[string]any{"num": 0.001},
			map[string]any{"num": float64(42)},
			map[string]any{"num": 3.14},
		},
	}

	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Should not contain scientific notation
	if strings.Contains(encoded, "e+") || strings.Contains(encoded, "e-") || strings.Contains(encoded, "E") {
		t.Errorf("Should not use scientific notation, got: %s", encoded)
	}

	// Should contain actual values
	if !strings.Contains(encoded, "1000000") {
		t.Errorf("Expected 1000000, got: %s", encoded)
	}
	if !strings.Contains(encoded, "0.001") {
		t.Errorf("Expected 0.001, got: %s", encoded)
	}
	if !strings.Contains(encoded, "42") {
		t.Errorf("Expected 42, got: %s", encoded)
	}
	if !strings.Contains(encoded, "3.14") {
		t.Errorf("Expected 3.14, got: %s", encoded)
	}
}
