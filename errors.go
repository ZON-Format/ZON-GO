// Package zon provides ZON encoding and decoding functionality.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon

import (
	"fmt"
)

// DecodeError represents an error that occurred during ZON decoding.
type DecodeError struct {
	Code    string // Error code (e.g., "E001", "E002")
	Message string // Detailed error message
	Line    int    // Line number where error occurred (0 if unknown)
	Column  int    // Column position (0 if unknown)
	Context string // Relevant context snippet
}

// Error implements the error interface.
func (e *DecodeError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Line > 0 {
		msg += fmt.Sprintf(" (line %d)", e.Line)
	}
	if e.Context != "" {
		msg += fmt.Sprintf("\n  Context: %s", e.Context)
	}
	return msg
}

// NewDecodeError creates a new DecodeError with the given parameters.
func NewDecodeError(code, message string, line int, context string) *DecodeError {
	return &DecodeError{
		Code:    code,
		Message: message,
		Line:    line,
		Context: context,
	}
}

// Error codes for ZON decoding errors
const (
	// ErrRowCountMismatch indicates the row count doesn't match the declared count
	ErrRowCountMismatch = "E001"
	// ErrFieldCountMismatch indicates the field count doesn't match the column count
	ErrFieldCountMismatch = "E002"
	// ErrDocumentTooLarge indicates the document exceeds the maximum size
	ErrDocumentTooLarge = "E301"
	// ErrLineTooLong indicates a line exceeds the maximum length
	ErrLineTooLong = "E302"
	// ErrArrayTooLarge indicates an array exceeds the maximum length
	ErrArrayTooLarge = "E303"
	// ErrTooManyKeys indicates an object has too many keys
	ErrTooManyKeys = "E304"
)

// EncodeError represents an error that occurred during ZON encoding.
type EncodeError struct {
	Message string
}

// Error implements the error interface.
func (e *EncodeError) Error() string {
	return e.Message
}

// NewEncodeError creates a new EncodeError with the given message.
func NewEncodeError(message string) *EncodeError {
	return &EncodeError{Message: message}
}
