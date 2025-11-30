// Package zon provides a ZON (Zero Overhead Notation) encoder and decoder.
// ZON is a compact, human-readable format for encoding JSON data,
// optimized for LLM token efficiency.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon

// ZON Protocol Constants v1.0.5

// Format markers
const (
	// TableMarker is the @ symbol indicating table structure
	TableMarker = '@'
	// MetaSeparator is the colon separating keys and values
	MetaSeparator = ':'
)

// Reserved tokens (for future use)
const (
	// GasToken is a placeholder token
	GasToken = '_'
	// LiquidToken is a variable token
	LiquidToken = '^'
)

// DefaultAnchorInterval is the default interval for large datasets
const DefaultAnchorInterval = 100

// Security limits (DOS prevention)
const (
	// MaxDocumentSize is the maximum document size in bytes (100 MB)
	MaxDocumentSize = 100 * 1024 * 1024
	// MaxLineLength is the maximum line length in bytes (1 MB)
	MaxLineLength = 1024 * 1024
	// MaxArrayLength is the maximum array length (1 million items)
	MaxArrayLength = 1_000_000
	// MaxObjectKeys is the maximum number of object keys (100K keys)
	MaxObjectKeys = 100_000
	// MaxNestingDepth is the maximum nesting depth (100 levels)
	MaxNestingDepth = 100
)

// Legacy compatibility (v1.x)
const (
	// LegacyTableMarker for backward compatibility
	LegacyTableMarker = '@'
	// InlineThresholdRows for metadata flattening
	InlineThresholdRows = 0
)
