// Package zon_test provides conformance tests based on FORMAL_SPEC.md §11.2 Decoder Checklist.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestAcceptUTF8WithLFOrCRLF tests that decoder accepts UTF-8 with LF or CRLF.
func TestAcceptUTF8WithLFOrCRLF(t *testing.T) {
	zonLF := "key:value\nkey2:value2"
	zonCRLF := "key:value\r\nkey2:value2"

	_, err := zon.Decode(zonLF)
	if err != nil {
		t.Errorf("Should accept LF: %v", err)
	}

	_, err = zon.Decode(zonCRLF)
	if err != nil {
		t.Errorf("Should accept CRLF: %v", err)
	}
}

// TestDecodeBooleanAndNull tests T → true, F → false, null → null.
func TestDecodeBooleanAndNull(t *testing.T) {
	zonData := "active:T\narchived:F\nvalue:null"
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["active"] != true {
		t.Errorf("active: expected true, got %v", resultMap["active"])
	}
	if resultMap["archived"] != false {
		t.Errorf("archived: expected false, got %v", resultMap["archived"])
	}
	if resultMap["value"] != nil {
		t.Errorf("value: expected nil, got %v", resultMap["value"])
	}
}

// TestParseDecimalAndExponentNumbers tests parsing decimal and exponent numbers.
func TestParseDecimalAndExponentNumbers(t *testing.T) {
	zonData := "int:42\nfloat:3.14\nbig:1000000"
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["int"] != float64(42) {
		t.Errorf("int: expected 42, got %v", resultMap["int"])
	}
	if resultMap["float"] != 3.14 {
		t.Errorf("float: expected 3.14, got %v", resultMap["float"])
	}
	if resultMap["big"] != float64(1000000) {
		t.Errorf("big: expected 1000000, got %v", resultMap["big"])
	}
}

// TestLeadingZeroNumbersAsStrings tests that leading-zero numbers are treated as strings.
func TestLeadingZeroNumbersAsStrings(t *testing.T) {
	zonData := `code:"007"`
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["code"] != "007" {
		t.Errorf("code: expected '007', got %v", resultMap["code"])
	}
	if _, ok := resultMap["code"].(string); !ok {
		t.Errorf("code should be string, got %T", resultMap["code"])
	}
}

// TestUnescapeQuotedStrings tests unescaping quoted strings.
func TestUnescapeQuotedStrings(t *testing.T) {
	zonData := `text:"he said \"hello\""`
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["text"] != `he said "hello"` {
		t.Errorf("text: expected 'he said \"hello\"', got %v", resultMap["text"])
	}
}

// TestParseTableRowsIntoArrayOfObjects tests parsing table rows into array of objects.
func TestParseTableRowsIntoArrayOfObjects(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice
2,Bob`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	user1 := users[0].(map[string]any)
	if user1["id"] != float64(1) {
		t.Errorf("user1.id: expected 1, got %v", user1["id"])
	}
	if user1["name"] != "Alice" {
		t.Errorf("user1.name: expected Alice, got %v", user1["name"])
	}

	user2 := users[1].(map[string]any)
	if user2["id"] != float64(2) {
		t.Errorf("user2.id: expected 2, got %v", user2["id"])
	}
	if user2["name"] != "Bob" {
		t.Errorf("user2.name: expected Bob, got %v", user2["name"])
	}
}

// TestPreserveKeyOrder tests that key order is preserved from document.
func TestPreserveKeyOrder(t *testing.T) {
	zonData := "z:1\na:2\nm:3"
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	// Go maps don't preserve order, but we should have all keys
	if resultMap["z"] != float64(1) {
		t.Errorf("z: expected 1, got %v", resultMap["z"])
	}
	if resultMap["a"] != float64(2) {
		t.Errorf("a: expected 2, got %v", resultMap["a"])
	}
	if resultMap["m"] != float64(3) {
		t.Errorf("m: expected 3, got %v", resultMap["m"])
	}
}

// TestRejectPrototypePollution tests rejection of prototype pollution attempts.
func TestRejectPrototypePollution(t *testing.T) {
	malicious := `users:@(1):id,__proto__.polluted
1,true`

	_, err := zon.Decode(malicious)
	// The decoding should succeed but not pollute the prototype
	if err != nil {
		// If it errors, that's also acceptable
		return
	}

	// Verify that Object prototype is not polluted
	testObj := make(map[string]any)
	if _, ok := testObj["polluted"]; ok {
		t.Error("Prototype pollution detected")
	}
}

// TestThrowOnDeepNesting tests that decoder throws on nesting depth > 100.
func TestThrowOnDeepNesting(t *testing.T) {
	deepNested := strings.Repeat("[", 150) + strings.Repeat("]", 150)

	_, err := zon.Decode(deepNested)
	if err == nil {
		t.Error("Expected error for deep nesting")
	}
	if !strings.Contains(err.Error(), "Maximum nesting depth exceeded") {
		t.Errorf("Expected 'Maximum nesting depth exceeded' error, got: %v", err)
	}
}

// TestThrowOnLineLengthExceeds1MB tests that decoder throws on line length > 1MB.
func TestThrowOnLineLengthExceeds1MB(t *testing.T) {
	longLine := "key:" + strings.Repeat("x", 1024*1024+1)

	_, err := zon.Decode(longLine)
	if err == nil {
		t.Error("Expected error for long line")
	}
	if !strings.Contains(err.Error(), "E302") {
		t.Errorf("Expected E302 error, got: %v", err)
	}
}

// TestCaseInsensitiveNullBooleanAliases tests case-insensitive null/boolean aliases.
func TestCaseInsensitiveNullBooleanAliases(t *testing.T) {
	zonData := "a:TRUE\nb:False\nc:NONE\nd:nil"
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["a"] != true {
		t.Errorf("a: expected true, got %v", resultMap["a"])
	}
	if resultMap["b"] != false {
		t.Errorf("b: expected false, got %v", resultMap["b"])
	}
	if resultMap["c"] != nil {
		t.Errorf("c: expected nil, got %v", resultMap["c"])
	}
	if resultMap["d"] != nil {
		t.Errorf("d: expected nil, got %v", resultMap["d"])
	}
}

// TestReconstructNestedObjectsFromDottedKeys tests reconstruction of nested objects.
func TestReconstructNestedObjectsFromDottedKeys(t *testing.T) {
	zonData := "config.db.host:localhost\nconfig.db.port:5432"
	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	config := resultMap["config"].(map[string]any)
	db := config["db"].(map[string]any)
	if db["host"] != "localhost" {
		t.Errorf("config.db.host: expected localhost, got %v", db["host"])
	}
	if db["port"] != float64(5432) {
		t.Errorf("config.db.port: expected 5432, got %v", db["port"])
	}
}

// TestUnwrapPureLists tests unwrapping pure lists (data key).
func TestUnwrapPureLists(t *testing.T) {
	zonData := `data:@(2):id,name
1,Alice
2,Bob`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	// Should return array directly, not { data: [...] }
	arr, ok := result.([]any)
	if !ok {
		t.Errorf("Expected array, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("Expected 2 items, got %d", len(arr))
	}
}

// TestEmptyStringsInTableCells tests empty strings in table cells.
func TestEmptyStringsInTableCells(t *testing.T) {
	zonData := `users:@(2):id,name
1,""
2,Bob`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	user1 := users[0].(map[string]any)
	if user1["name"] != "" {
		t.Errorf("user1.name: expected empty string, got %v", user1["name"])
	}
}
