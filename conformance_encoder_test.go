// Package zon_test provides conformance tests based on FORMAL_SPEC.md §11.1 Encoder Checklist.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"math"
	"regexp"
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestEmitUTF8WithLF tests that encoder emits UTF-8 with LF line endings.
func TestEmitUTF8WithLF(t *testing.T) {
	data := map[string]any{"a": float64(1), "b": float64(2)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Should use LF, not CRLF
	if strings.Contains(encoded, "\r\n") {
		t.Errorf("Should not contain CRLF, got: %s", encoded)
	}
}

// TestEncodeBoolsAsTF tests that booleans are encoded as T/F.
func TestEncodeBoolsAsTF(t *testing.T) {
	data := map[string]any{"active": true, "archived": false}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "active:T") {
		t.Errorf("Expected active:T, got: %s", encoded)
	}
	if !strings.Contains(encoded, "archived:F") {
		t.Errorf("Expected archived:F, got: %s", encoded)
	}
	if strings.Contains(encoded, "true") || strings.Contains(encoded, "false") {
		t.Errorf("Should not contain true/false, got: %s", encoded)
	}
}

// TestEncodeNullAsNull tests that null is encoded as "null".
func TestEncodeNullAsNull(t *testing.T) {
	data := map[string]any{"value": nil}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "value:null") {
		t.Errorf("Expected value:null, got: %s", encoded)
	}
}

// TestEmitCanonicalNumbers tests canonical number formatting.
func TestEmitCanonicalNumbers(t *testing.T) {
	data := map[string]any{"int": float64(42), "float": 3.14, "big": float64(1000000)}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// No scientific notation
	if !strings.Contains(encoded, "1000000") {
		t.Errorf("Expected 1000000, got: %s", encoded)
	}
	if strings.Contains(encoded, "1e6") || strings.Contains(encoded, "1e+6") {
		t.Errorf("Should not use scientific notation, got: %s", encoded)
	}

	// Has decimal for floats
	if !strings.Contains(encoded, "3.14") {
		t.Errorf("Expected 3.14, got: %s", encoded)
	}
}

// TestNormalizeNaNInfinity tests that NaN/Infinity are normalized to null.
func TestNormalizeNaNInfinity(t *testing.T) {
	data := map[string]any{
		"nan":    math.NaN(),
		"inf":    math.Inf(1),
		"negInf": math.Inf(-1),
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "nan:null") {
		t.Errorf("Expected nan:null, got: %s", encoded)
	}
	if !strings.Contains(encoded, "inf:null") {
		t.Errorf("Expected inf:null, got: %s", encoded)
	}
	if !strings.Contains(encoded, "negInf:null") {
		t.Errorf("Expected negInf:null, got: %s", encoded)
	}
}

// TestDetectUniformArraysAsTable tests uniform arrays → table format detection.
func TestDetectUniformArraysAsTable(t *testing.T) {
	data := map[string]any{
		"users": []any{
			map[string]any{"id": float64(1), "name": "Alice"},
			map[string]any{"id": float64(2), "name": "Bob"},
		},
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Should have table marker
	matched, _ := regexp.MatchString(`users:@\(\d+\)`, encoded)
	if !matched {
		t.Errorf("Expected table marker, got: %s", encoded)
	}
	if !strings.Contains(encoded, "id,name") && !strings.Contains(encoded, "name,id") {
		t.Errorf("Expected column headers, got: %s", encoded)
	}
}

// TestEmitTableHeadersWithCountAndColumns tests table headers with count and columns.
func TestEmitTableHeadersWithCountAndColumns(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"x": float64(1), "y": float64(2)},
			map[string]any{"x": float64(3), "y": float64(4)},
			map[string]any{"x": float64(5), "y": float64(6)},
		},
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if !strings.Contains(encoded, "items:@(3):") {
		t.Errorf("Expected items:@(3):, got: %s", encoded)
	}
}

// TestSortColumnsAlphabetically tests that columns are sorted alphabetically.
func TestSortColumnsAlphabetically(t *testing.T) {
	data := map[string]any{
		"records": []any{
			map[string]any{"z": float64(1), "a": float64(2), "m": float64(3)},
		},
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Columns should be sorted: a, m, z
	matched, _ := regexp.MatchString(`records:@\(1\):a,m,z`, encoded)
	if !matched {
		t.Errorf("Expected sorted columns a,m,z, got: %s", encoded)
	}
}

// TestQuoteStringsWithSpecialChars tests quoting strings with special characters.
func TestQuoteStringsWithSpecialChars(t *testing.T) {
	data := map[string]any{
		"comma": "a,b",
		"colon": "x:y",
		"quote": `say "hi"`,
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Comma should be quoted
	if !strings.Contains(encoded, `"a,b"`) {
		t.Errorf("Expected quoted comma, got: %s", encoded)
	}

	// v2.0.5: Colons are allowed unquoted
	if !strings.Contains(encoded, "x:y") && !strings.Contains(encoded, "colon:x:y") {
		t.Errorf("Expected colon in value, got: %s", encoded)
	}
}

// TestEscapeQuotesInStrings tests escaping quotes in strings.
func TestEscapeQuotesInStrings(t *testing.T) {
	data := map[string]any{"text": `he said "hello"`}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Uses escaped quotes in JSON-style
	if !strings.Contains(encoded, `\"hello\"`) && !strings.Contains(encoded, `""hello""`) {
		t.Errorf("Expected escaped quotes, got: %s", encoded)
	}
}

// TestProduceDeterministicOutput tests deterministic output.
func TestProduceDeterministicOutput(t *testing.T) {
	data := map[string]any{"b": float64(2), "a": float64(1), "c": float64(3)}

	encoded1, _ := zon.Encode(data)
	encoded2, _ := zon.Encode(data)

	if encoded1 != encoded2 {
		t.Errorf("Output should be deterministic\n%s\n!=\n%s", encoded1, encoded2)
	}
}

// TestHandleEmptyObjects tests empty object handling.
func TestHandleEmptyObjects(t *testing.T) {
	data := map[string]any{}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Empty object is empty string in ZON
	if encoded != "" {
		t.Errorf("Expected empty string for empty object, got: %s", encoded)
	}
}

// TestHandleEmptyArrays tests empty array handling.
func TestHandleEmptyArrays(t *testing.T) {
	data := map[string]any{"items": []any{}}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if encoded == "" {
		t.Error("Expected non-empty output for object with empty array")
	}
}
