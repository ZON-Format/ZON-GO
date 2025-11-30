// Package zon_test provides security limit tests (DOS Prevention).
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"strconv"
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestDocumentSizeLimit tests document size limit.
func TestDocumentSizeLimit(t *testing.T) {
	// Test normal document under limit
	doc := strings.Repeat("test:value\n", 1000)
	_, err := zon.Decode(doc)
	if err != nil {
		t.Errorf("Should allow documents under 100MB: %v", err)
	}
}

// TestLineLengthLimit tests line length limit.
func TestLineLengthLimit(t *testing.T) {
	// Test long line exceeds 1MB
	longLine := "key:" + strings.Repeat("x", zon.MaxLineLength+1)

	_, err := zon.Decode(longLine)
	if err == nil {
		t.Error("Expected error for long line")
	}
	decodeErr, ok := err.(*zon.DecodeError)
	if !ok {
		t.Errorf("Expected DecodeError, got %T", err)
	}
	if decodeErr.Code != zon.ErrLineTooLong {
		t.Errorf("Expected E302, got %s", decodeErr.Code)
	}
}

// TestLineLengthUnderLimit tests lines under limit.
func TestLineLengthUnderLimit(t *testing.T) {
	line := "key:" + strings.Repeat("x", 1000)
	result, err := zon.Decode(line)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["key"] == nil {
		t.Error("Expected key to be defined")
	}
}

// TestArrayLengthLimitDefined tests that array length limit is defined.
func TestArrayLengthLimitDefined(t *testing.T) {
	// Verify the limit exists in constants
	if zon.MaxArrayLength != 1_000_000 {
		t.Errorf("Expected MaxArrayLength 1000000, got %d", zon.MaxArrayLength)
	}
}

// TestObjectKeyLimitDefined tests that object key limit is defined.
func TestObjectKeyLimitDefined(t *testing.T) {
	// Verify the limit exists in constants
	if zon.MaxObjectKeys != 100_000 {
		t.Errorf("Expected MaxObjectKeys 100000, got %d", zon.MaxObjectKeys)
	}
}

// TestObjectsUnderKeyLimit tests objects under 100K keys.
func TestObjectsUnderKeyLimit(t *testing.T) {
	var keys []string
	for i := 0; i < 100; i++ {
		keys = append(keys, "key"+strconv.Itoa(i)+":"+strconv.Itoa(i))
	}
	zonData := `data:"{` + strings.Join(keys, ",") + `}"`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	dataMap := resultMap["data"].(map[string]any)
	if len(dataMap) != 100 {
		t.Errorf("Expected 100 keys, got %d", len(dataMap))
	}
}

// TestNestingDepthExceeds100 tests that nesting exceeds 100 levels.
func TestNestingDepthExceeds100(t *testing.T) {
	nested := strings.Repeat("[", 150) + strings.Repeat("]", 150)

	_, err := zon.Decode(nested)
	if err == nil {
		t.Error("Expected error for deep nesting")
	}
	if !strings.Contains(err.Error(), "Maximum nesting depth exceeded") {
		t.Errorf("Expected nesting depth error, got: %v", err)
	}
}

// TestNestingDepthUnder100 tests nesting under 100 levels.
func TestNestingDepthUnder100(t *testing.T) {
	nested := strings.Repeat("[", 50) + strings.Repeat("]", 50)

	result, err := zon.Decode(nested)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestCombinedLimitsWithNormalData tests normal data within all limits.
func TestCombinedLimitsWithNormalData(t *testing.T) {
	zonData := `metadata:"{version:1.0.5,env:prod}"
users:@(3):id,name
1,Alice
2,Bob
3,Carol
tags:"[nodejs,typescript,llm]"`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	metadata := resultMap["metadata"].(map[string]any)
	if metadata["version"] != "1.0.5" {
		t.Errorf("Expected version 1.0.5, got %v", metadata["version"])
	}

	tags := resultMap["tags"].([]any)
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}
