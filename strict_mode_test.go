// Package zon_test provides strict mode validation tests.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon_test

import (
	"strings"
	"testing"

	zon "github.com/ZON-Format/zon-go"
)

// TestRowCountMismatchStrictMode tests row count mismatch in strict mode.
func TestRowCountMismatchStrictMode(t *testing.T) {
	zonData := `users:@(3):id,name
1,Alice
2,Bob`

	_, err := zon.Decode(zonData)
	if err == nil {
		t.Error("Expected error for row count mismatch")
	}

	decodeErr, ok := err.(*zon.DecodeError)
	if !ok {
		t.Errorf("Expected DecodeError, got %T", err)
	}
	if decodeErr.Code != zon.ErrRowCountMismatch {
		t.Errorf("Expected E001, got %s", decodeErr.Code)
	}
	if !strings.Contains(err.Error(), "Row count mismatch") {
		t.Errorf("Expected row count mismatch error, got: %v", err)
	}
}

// TestRowCountMismatchNonStrictMode tests row count mismatch in non-strict mode.
func TestRowCountMismatchNonStrictMode(t *testing.T) {
	zonData := `users:@(3):id,name
1,Alice
2,Bob`

	result, err := zon.DecodeWithOptions(zonData, &zon.DecodeOptions{Strict: false})
	if err != nil {
		t.Fatalf("Should allow fewer rows in non-strict mode: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

// TestRowCountMatches tests that matching row count passes.
func TestRowCountMatches(t *testing.T) {
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
}

// TestFieldCountMismatchStrictMode tests field count mismatch in strict mode.
func TestFieldCountMismatchStrictMode(t *testing.T) {
	zonData := `users:@(2):id,name,role
1,Alice
2,Bob,admin`

	_, err := zon.Decode(zonData)
	if err == nil {
		t.Error("Expected error for field count mismatch")
	}

	decodeErr, ok := err.(*zon.DecodeError)
	if !ok {
		t.Errorf("Expected DecodeError, got %T", err)
	}
	if decodeErr.Code != zon.ErrFieldCountMismatch {
		t.Errorf("Expected E002, got %s", decodeErr.Code)
	}
	if !strings.Contains(err.Error(), "Field count mismatch") {
		t.Errorf("Expected field count mismatch error, got: %v", err)
	}
}

// TestFieldCountMismatchNonStrictMode tests field count mismatch in non-strict mode.
func TestFieldCountMismatchNonStrictMode(t *testing.T) {
	zonData := `users:@(2):id,name,role
1,Alice
2,Bob,admin`

	result, err := zon.DecodeWithOptions(zonData, &zon.DecodeOptions{Strict: false})
	if err != nil {
		t.Fatalf("Should allow missing fields in non-strict mode: %v", err)
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
	if user2["role"] != "admin" {
		t.Errorf("user2.role: expected admin, got %v", user2["role"])
	}
}

// TestFieldCountMatchesStrictMode tests that matching field count passes.
func TestFieldCountMatchesStrictMode(t *testing.T) {
	zonData := `users:@(2):id,name,role
1,Alice,user
2,Bob,admin`

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
	if user1["role"] != "user" {
		t.Errorf("user1.role: expected user, got %v", user1["role"])
	}
}

// TestSparseFieldsInStrictMode tests sparse fields in strict mode.
func TestSparseFieldsInStrictMode(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice,role:admin,score:98
2,Bob`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)

	user1 := users[0].(map[string]any)
	if user1["id"] != float64(1) {
		t.Errorf("user1.id: expected 1, got %v", user1["id"])
	}
	if user1["name"] != "Alice" {
		t.Errorf("user1.name: expected Alice, got %v", user1["name"])
	}
	if user1["role"] != "admin" {
		t.Errorf("user1.role: expected admin, got %v", user1["role"])
	}
	if user1["score"] != float64(98) {
		t.Errorf("user1.score: expected 98, got %v", user1["score"])
	}

	user2 := users[1].(map[string]any)
	if user2["id"] != float64(2) {
		t.Errorf("user2.id: expected 2, got %v", user2["id"])
	}
	if user2["name"] != "Bob" {
		t.Errorf("user2.name: expected Bob, got %v", user2["name"])
	}
}

// TestErrorCodeInErrorObject tests error code in error object.
func TestErrorCodeInErrorObject(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice`

	_, err := zon.Decode(zonData)
	if err == nil {
		t.Fatal("Expected error")
	}

	decodeErr, ok := err.(*zon.DecodeError)
	if !ok {
		t.Errorf("Expected DecodeError, got %T", err)
	}
	if decodeErr.Code != zon.ErrRowCountMismatch {
		t.Errorf("Expected E001, got %s", decodeErr.Code)
	}
}

// TestContextInErrorMessage tests context in error message.
func TestContextInErrorMessage(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice`

	_, err := zon.Decode(zonData)
	if err == nil {
		t.Fatal("Expected error")
	}

	decodeErr, ok := err.(*zon.DecodeError)
	if !ok {
		t.Errorf("Expected DecodeError, got %T", err)
	}
	if decodeErr.Context == "" {
		t.Error("Expected context to be set")
	}
	if !strings.Contains(err.Error(), "Table: users") {
		t.Errorf("Expected 'Table: users' in error, got: %v", err)
	}
}

// TestStrictModeEnabledByDefault tests that strict mode is enabled by default.
func TestStrictModeEnabledByDefault(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice`

	// Should throw because default is strict: true
	_, err := zon.Decode(zonData)
	if err == nil {
		t.Error("Expected error in default strict mode")
	}
}

// TestExplicitStrictMode tests explicit strict mode.
func TestExplicitStrictMode(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice`

	_, err := zon.DecodeWithOptions(zonData, &zon.DecodeOptions{Strict: true})
	if err == nil {
		t.Error("Expected error in explicit strict mode")
	}
}

// TestMultipleTablesIndependentValidation tests multiple tables independent validation.
func TestMultipleTablesIndependentValidation(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice
2,Bob
products:@(1):id,title
100,Widget`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	products := resultMap["products"].([]any)
	if len(products) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products))
	}
}

// TestValidDataAcrossMultipleTables tests valid data across multiple tables.
func TestValidDataAcrossMultipleTables(t *testing.T) {
	zonData := `users:@(2):id,name
1,Alice
2,Bob
products:@(2):id,title
100,Widget
200,Gadget`

	result, err := zon.Decode(zonData)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	resultMap := result.(map[string]any)
	users := resultMap["users"].([]any)
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	products := resultMap["products"].([]any)
	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}
}
