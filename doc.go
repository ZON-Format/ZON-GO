// Package zon provides a ZON (Zero Overhead Notation) encoder and decoder.
//
// ZON is a compact, human-readable format for encoding JSON data,
// optimized for LLM token efficiency. It achieves 35-50% token reduction
// vs JSON through tabular encoding, single-character primitives, and
// intelligent compression while maintaining 100% data fidelity.
//
// Basic usage:
//
//	import "github.com/ZON-Format/zon-go"
//
//	// Encode data to ZON format
//	data := map[string]any{
//	    "users": []any{
//	        map[string]any{"id": 1, "name": "Alice", "active": true},
//	        map[string]any{"id": 2, "name": "Bob", "active": false},
//	    },
//	}
//	zonStr, err := zon.Encode(data)
//
//	// Decode ZON format back to data
//	decoded, err := zon.Decode(zonStr)
//
// The encoder automatically detects uniform arrays of objects and encodes
// them in tabular format for maximum token efficiency:
//
//	users:@(2):active,id,name
//	T,1,Alice
//	F,2,Bob
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon
