// Package zon_test provides tests for ZON encoding and decoding.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"reflect"
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestRoundTripEmptyObject tests empty object round-trip.
func TestRoundTripEmptyObject(t *testing.T) {
	data := map[string]any{}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if !reflect.DeepEqual(decoded, data) {
		t.Errorf("Expected %v, got %v", data, decoded)
	}
}

// TestRoundTripSimpleMetadata tests simple metadata round-trip.
func TestRoundTripSimpleMetadata(t *testing.T) {
	data := map[string]any{
		"name":   "Alice",
		"age":    float64(30),
		"active": true,
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
	if decodedMap["name"] != data["name"] {
		t.Errorf("name: expected %v, got %v", data["name"], decodedMap["name"])
	}
	if decodedMap["age"] != data["age"] {
		t.Errorf("age: expected %v, got %v", data["age"], decodedMap["age"])
	}
	if decodedMap["active"] != data["active"] {
		t.Errorf("active: expected %v, got %v", data["active"], decodedMap["active"])
	}
}

// TestRoundTripNestedObject tests nested object round-trip.
func TestRoundTripNestedObject(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Bob",
			"profile": map[string]any{
				"age":  float64(25),
				"city": "NYC",
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
	user := decodedMap["user"].(map[string]any)
	profile := user["profile"].(map[string]any)
	if user["name"] != "Bob" {
		t.Errorf("user.name: expected Bob, got %v", user["name"])
	}
	if profile["age"] != float64(25) {
		t.Errorf("user.profile.age: expected 25, got %v", profile["age"])
	}
	if profile["city"] != "NYC" {
		t.Errorf("user.profile.city: expected NYC, got %v", profile["city"])
	}
}

// TestRoundTripArrayOfObjects tests array of objects (table) round-trip.
func TestRoundTripArrayOfObjects(t *testing.T) {
	data := []any{
		map[string]any{"id": float64(1), "name": "Alice", "score": float64(95)},
		map[string]any{"id": float64(2), "name": "Bob", "score": float64(87)},
		map[string]any{"id": float64(3), "name": "Charlie", "score": float64(92)},
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	decodedArr := decoded.([]any)
	if len(decodedArr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(decodedArr))
	}
	first := decodedArr[0].(map[string]any)
	if first["id"] != float64(1) {
		t.Errorf("Expected id 1, got %v", first["id"])
	}
	if first["name"] != "Alice" {
		t.Errorf("Expected name Alice, got %v", first["name"])
	}
}

// TestRoundTripMixedMetadataAndTable tests mixed metadata and table round-trip.
func TestRoundTripMixedMetadataAndTable(t *testing.T) {
	data := map[string]any{
		"title": "Sales Report",
		"year":  float64(2024),
		"records": []any{
			map[string]any{"month": "Jan", "sales": float64(1000)},
			map[string]any{"month": "Feb", "sales": float64(1200)},
			map[string]any{"month": "Mar", "sales": float64(1100)},
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
	if decodedMap["title"] != "Sales Report" {
		t.Errorf("title: expected Sales Report, got %v", decodedMap["title"])
	}
	if decodedMap["year"] != float64(2024) {
		t.Errorf("year: expected 2024, got %v", decodedMap["year"])
	}
	records := decodedMap["records"].([]any)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

// TestRoundTripBooleanValues tests boolean values round-trip.
func TestRoundTripBooleanValues(t *testing.T) {
	data := map[string]any{
		"success": true,
		"error":   false,
		"items": []any{
			map[string]any{"id": float64(1), "active": true},
			map[string]any{"id": float64(2), "active": false},
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
	if decodedMap["success"] != true {
		t.Errorf("success: expected true, got %v", decodedMap["success"])
	}
	if decodedMap["error"] != false {
		t.Errorf("error: expected false, got %v", decodedMap["error"])
	}
}

// TestRoundTripNullValues tests null values round-trip.
func TestRoundTripNullValues(t *testing.T) {
	data := map[string]any{
		"name":  "Test",
		"value": nil,
		"items": []any{
			map[string]any{"id": float64(1), "data": nil},
			map[string]any{"id": float64(2), "data": "value"},
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
	if decodedMap["name"] != "Test" {
		t.Errorf("name: expected Test, got %v", decodedMap["name"])
	}
	if decodedMap["value"] != nil {
		t.Errorf("value: expected nil, got %v", decodedMap["value"])
	}
}

// TestRoundTripNumbers tests numbers (integers and floats) round-trip.
func TestRoundTripNumbers(t *testing.T) {
	data := map[string]any{
		"integer":       float64(42),
		"float":         3.14,
		"negative":      float64(-10),
		"negativeFloat": -2.5,
		"items": []any{
			map[string]any{"id": float64(1), "value": float64(100)},
			map[string]any{"id": float64(2), "value": 200.5},
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
	if decodedMap["integer"] != float64(42) {
		t.Errorf("integer: expected 42, got %v", decodedMap["integer"])
	}
	if decodedMap["float"] != 3.14 {
		t.Errorf("float: expected 3.14, got %v", decodedMap["float"])
	}
}

// TestRoundTripStringsWithSpecialChars tests strings with special characters.
func TestRoundTripStringsWithSpecialChars(t *testing.T) {
	data := map[string]any{
		"plain":       "hello",
		"withComma":   "hello, world",
		"withQuotes":  `say "hello"`,
		"withNewline": "line1\nline2",
		"items": []any{
			map[string]any{"id": float64(1), "text": "normal"},
			map[string]any{"id": float64(2), "text": "with, comma"},
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
	if decodedMap["plain"] != "hello" {
		t.Errorf("plain: expected hello, got %v", decodedMap["plain"])
	}
	if decodedMap["withComma"] != "hello, world" {
		t.Errorf("withComma: expected 'hello, world', got %v", decodedMap["withComma"])
	}
	if decodedMap["withNewline"] != "line1\nline2" {
		t.Errorf("withNewline: expected 'line1\\nline2', got %v", decodedMap["withNewline"])
	}
}

// TestRoundTripEmptyArrays tests empty arrays.
func TestRoundTripEmptyArrays(t *testing.T) {
	data := map[string]any{
		"empty": []any{},
		"nested": map[string]any{
			"also_empty": []any{},
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
	emptyArr := decodedMap["empty"].([]any)
	if len(emptyArr) != 0 {
		t.Errorf("empty: expected 0 items, got %d", len(emptyArr))
	}
}

// TestRoundTripNestedArrays tests nested arrays in metadata.
func TestRoundTripNestedArrays(t *testing.T) {
	data := map[string]any{
		"tags": []any{"javascript", "typescript", "node"},
		"matrix": []any{
			[]any{float64(1), float64(2)},
			[]any{float64(3), float64(4)},
		},
		"items": []any{
			map[string]any{"id": float64(1), "values": []any{float64(10), float64(20)}},
			map[string]any{"id": float64(2), "values": []any{float64(30), float64(40)}},
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
	tags := decodedMap["tags"].([]any)
	if len(tags) != 3 {
		t.Errorf("tags: expected 3, got %d", len(tags))
	}
}

// TestHikesExample tests the full hikes example from README.
func TestHikesExample(t *testing.T) {
	data := map[string]any{
		"context": map[string]any{
			"task":     "Our favorite hikes together",
			"location": "Boulder",
			"season":   "spring_2025",
		},
		"friends": []any{"ana", "luis", "sam"},
		"hikes": []any{
			map[string]any{
				"id":            float64(1),
				"name":          "Blue Lake Trail",
				"distanceKm":    7.5,
				"elevationGain": float64(320),
				"companion":     "ana",
				"wasSunny":      true,
			},
			map[string]any{
				"id":            float64(2),
				"name":          "Ridge Overlook",
				"distanceKm":    9.2,
				"elevationGain": float64(540),
				"companion":     "luis",
				"wasSunny":      false,
			},
			map[string]any{
				"id":            float64(3),
				"name":          "Wildflower Loop",
				"distanceKm":    5.1,
				"elevationGain": float64(180),
				"companion":     "sam",
				"wasSunny":      true,
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
	context := decodedMap["context"].(map[string]any)
	if context["task"] != "Our favorite hikes together" {
		t.Errorf("context.task mismatch")
	}

	hikes := decodedMap["hikes"].([]any)
	if len(hikes) != 3 {
		t.Errorf("Expected 3 hikes, got %d", len(hikes))
	}

	// Verify encoded format structure
	if !strings.Contains(encoded, "context.task:") {
		t.Errorf("Expected context.task: in encoded output")
	}
	if !strings.Contains(encoded, "hikes:@(3):") {
		t.Errorf("Expected hikes:@(3): in encoded output")
	}
}

// TestStringThatLooksLikeNumber tests string that looks like a number.
func TestStringThatLooksLikeNumber(t *testing.T) {
	data := map[string]any{
		"stringNumber": "123",
		"actualNumber": float64(123),
		"items": []any{
			map[string]any{"id": float64(1), "code": "001"},
			map[string]any{"id": float64(2), "code": "002"},
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
	if _, ok := decodedMap["stringNumber"].(string); !ok {
		t.Errorf("stringNumber should be string, got %T", decodedMap["stringNumber"])
	}
	if _, ok := decodedMap["actualNumber"].(float64); !ok {
		t.Errorf("actualNumber should be float64, got %T", decodedMap["actualNumber"])
	}
}

// TestStringThatLooksLikeBoolean tests string that looks like boolean.
func TestStringThatLooksLikeBoolean(t *testing.T) {
	data := map[string]any{
		"stringTrue":  "true",
		"actualTrue":  true,
		"stringFalse": "false",
		"actualFalse": false,
		"items": []any{
			map[string]any{"id": float64(1), "status": "T"},
			map[string]any{"id": float64(2), "status": true},
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
	if _, ok := decodedMap["stringTrue"].(string); !ok {
		t.Errorf("stringTrue should be string, got %T", decodedMap["stringTrue"])
	}
	if _, ok := decodedMap["actualTrue"].(bool); !ok {
		t.Errorf("actualTrue should be bool, got %T", decodedMap["actualTrue"])
	}
}

// TestEmptyStrings tests empty strings.
func TestEmptyStrings(t *testing.T) {
	data := map[string]any{
		"empty": "",
		"items": []any{
			map[string]any{"id": float64(1), "name": ""},
			map[string]any{"id": float64(2), "name": "value"},
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
	if decodedMap["empty"] != "" {
		t.Errorf("empty: expected empty string, got %v", decodedMap["empty"])
	}
}

// TestWhitespacePreservation tests whitespace preservation.
func TestWhitespacePreservation(t *testing.T) {
	data := map[string]any{
		"leading":  "  space",
		"trailing": "space  ",
		"both":     "  both  ",
		"items": []any{
			map[string]any{"id": float64(1), "text": "  padded  "},
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
	if decodedMap["leading"] != "  space" {
		t.Errorf("leading: expected '  space', got '%v'", decodedMap["leading"])
	}
	if decodedMap["trailing"] != "space  " {
		t.Errorf("trailing: expected 'space  ', got '%v'", decodedMap["trailing"])
	}
}

// TestVeryLongStrings tests very long strings.
func TestVeryLongStrings(t *testing.T) {
	longString := strings.Repeat("a", 1000)
	data := map[string]any{
		"long": longString,
		"items": []any{
			map[string]any{"id": float64(1), "text": longString},
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
	if decodedMap["long"] != longString {
		t.Errorf("long: expected %d chars, got %d", len(longString), len(decodedMap["long"].(string)))
	}
}

// TestLargeArrays tests large arrays.
func TestLargeArrays(t *testing.T) {
	items := make([]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]any{
			"id":    float64(i + 1),
			"name":  "Item " + strings.Repeat("x", i%10),
			"value": float64(i * 10),
		}
	}
	data := map[string]any{"items": items}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	decodedMap := decoded.(map[string]any)
	decodedItems := decodedMap["items"].([]any)
	if len(decodedItems) != 100 {
		t.Errorf("Expected 100 items, got %d", len(decodedItems))
	}
}

// TestArrayOfPrimitives tests array of primitives.
func TestArrayOfPrimitives(t *testing.T) {
	data := []any{"apple", "banana", "cherry"}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	decodedArr := decoded.([]any)
	if len(decodedArr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(decodedArr))
	}
	// Should be encoded as JSON array
	if !strings.HasPrefix(encoded, "[") {
		t.Errorf("Expected encoded to start with [, got %s", encoded[:10])
	}
}

// TestDeeplyNestedObjects tests deeply nested objects.
func TestDeeplyNestedObjects(t *testing.T) {
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": "deep",
					},
				},
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
	l1 := decodedMap["level1"].(map[string]any)
	l2 := l1["level2"].(map[string]any)
	l3 := l2["level3"].(map[string]any)
	l4 := l3["level4"].(map[string]any)
	if l4["value"] != "deep" {
		t.Errorf("Expected deep, got %v", l4["value"])
	}
}

// TestIntegerVsFloatDistinction tests integer vs float distinction.
func TestIntegerVsFloatDistinction(t *testing.T) {
	data := map[string]any{
		"integer":       float64(42),
		"float":         42.0,
		"explicitFloat": 3.14,
		"items": []any{
			map[string]any{"id": float64(1), "intVal": float64(100), "floatVal": 100.5},
			map[string]any{"id": float64(2), "intVal": float64(200), "floatVal": 200.0},
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
	if decodedMap["integer"] != float64(42) {
		t.Errorf("integer: expected 42, got %v", decodedMap["integer"])
	}
	if decodedMap["float"] != float64(42) {
		t.Errorf("float: expected 42.0, got %v", decodedMap["float"])
	}
	if decodedMap["explicitFloat"] != 3.14 {
		t.Errorf("explicitFloat: expected 3.14, got %v", decodedMap["explicitFloat"])
	}
}

// TestBooleanShorthandTF tests boolean shorthand T/F encoding.
func TestBooleanShorthandTF(t *testing.T) {
	data := []any{
		map[string]any{"id": float64(1), "flag": true},
		map[string]any{"id": float64(2), "flag": false},
		map[string]any{"id": float64(3), "flag": true},
	}
	encoded, err := zon.Encode(data)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Check that booleans are encoded as T/F
	if !strings.Contains(encoded, "T") || !strings.Contains(encoded, "F") {
		t.Errorf("Expected T and F in encoded output, got: %s", encoded)
	}

	decoded, err := zon.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	decodedArr := decoded.([]any)
	first := decodedArr[0].(map[string]any)
	if first["flag"] != true {
		t.Errorf("Expected true, got %v", first["flag"])
	}
}
