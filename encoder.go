// Package zon provides ZON encoding and decoding functionality.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Encoder encodes Go data structures to ZON format.
type Encoder struct {
	anchorInterval int
	safeStrRegex   *regexp.Regexp
}

// NewEncoder creates a new ZON encoder with the given anchor interval.
func NewEncoder(anchorInterval int) *Encoder {
	if anchorInterval <= 0 {
		anchorInterval = DefaultAnchorInterval
	}
	return &Encoder{
		anchorInterval: anchorInterval,
		safeStrRegex:   regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`),
	}
}

// Encode encodes data to ZON format.
func (e *Encoder) Encode(data any) (string, error) {
	visited := make(map[uintptr]bool)
	return e.encode(data, visited)
}

func (e *Encoder) encode(data any, visited map[uintptr]bool) (string, error) {
	// Extract primary stream (table data) and metadata
	streamData, metadata, streamKey := e.extractPrimaryStream(data)

	// Fallback for simple/empty data
	if streamData == nil && len(metadata) == 0 {
		switch v := data.(type) {
		case map[string]any:
			if len(v) == 0 {
				return "", nil
			}
			return e.formatZonNode(v, visited)
		case []any:
			return e.formatZonNode(v, visited)
		default:
			b, err := json.Marshal(data)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}

	// Special case: Detect schema uniformity for lists of dicts
	if arr, ok := data.([]any); ok && len(arr) > 0 {
		allObjects := true
		for _, item := range arr {
			if _, ok := item.(map[string]any); !ok {
				allObjects = false
				break
			}
		}
		if allObjects {
			irregularityScore := e.calculateIrregularity(arr)
			if irregularityScore > 0.6 {
				return e.formatZonNode(data, visited)
			}
		}
	}

	// If streamKey is empty (pure list input), use default key
	if streamData != nil && streamKey == "" {
		streamKey = "data"
	}

	var output []string

	// Write metadata
	if len(metadata) > 0 {
		metaLines := e.writeMetadata(metadata, visited)
		output = append(output, metaLines...)
	}

	// Write table
	if streamData != nil && streamKey != "" {
		if len(output) > 0 {
			output = append(output, "")
		}
		tableLines, err := e.writeTable(streamData, streamKey, visited)
		if err != nil {
			return "", err
		}
		output = append(output, tableLines...)
	}

	return strings.Join(output, "\n"), nil
}

// extractPrimaryStream finds the main table in the data.
func (e *Encoder) extractPrimaryStream(data any) ([]any, map[string]any, string) {
	if arr, ok := data.([]any); ok {
		// Only promote to table if it contains objects
		if len(arr) > 0 {
			if _, isObj := arr[0].(map[string]any); isObj {
				return arr, map[string]any{}, ""
			}
		}
		// Root-level array of primitives
		if len(arr) > 0 {
			allPrimitives := true
			for _, item := range arr {
				switch item.(type) {
				case map[string]any, []any:
					allPrimitives = false
				}
			}
			if allPrimitives {
				return nil, map[string]any{}, ""
			}
		}
		return nil, map[string]any{}, ""
	}

	if obj, ok := data.(map[string]any); ok {
		// Find largest list of objects
		type candidate struct {
			key   string
			arr   []any
			score int
		}
		var candidates []candidate

		for k, v := range obj {
			if arr, ok := v.([]any); ok && len(arr) > 0 {
				if firstObj, ok := arr[0].(map[string]any); ok {
					score := len(arr) * len(firstObj)
					candidates = append(candidates, candidate{k, arr, score})
				}
			}
		}

		if len(candidates) > 0 {
			// Sort by score descending, then by key alphabetically
			sort.Slice(candidates, func(i, j int) bool {
				if candidates[i].score != candidates[j].score {
					return candidates[i].score > candidates[j].score
				}
				return candidates[i].key < candidates[j].key
			})

			key := candidates[0].key
			stream := candidates[0].arr
			meta := make(map[string]any)
			for k, v := range obj {
				if k != key {
					meta[k] = v
				}
			}
			return stream, meta, key
		}
	}

	if obj, ok := data.(map[string]any); ok {
		return nil, obj, ""
	}
	return nil, map[string]any{}, ""
}

// writeMetadata writes metadata in YAML-like format.
func (e *Encoder) writeMetadata(metadata map[string]any, visited map[uintptr]bool) []string {
	var lines []string

	// Flatten top-level objects (depth 1)
	flattened := e.flatten(metadata, "", ".", 1, 0, visited)

	// Sort keys alphabetically
	keys := make([]string, 0, len(flattened))
	for k := range flattened {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := flattened[key]
		valStr, _ := e.formatValue(val, visited)

		// Colon-less syntax for root metadata if value starts with { or [
		if strings.HasPrefix(valStr, "{") || strings.HasPrefix(valStr, "[") {
			lines = append(lines, key+valStr)
		} else {
			lines = append(lines, key+string(MetaSeparator)+valStr)
		}
	}

	return lines
}

// writeTable writes table in v2.0.0 compact format.
func (e *Encoder) writeTable(stream []any, key string, visited map[uintptr]bool) ([]string, error) {
	if len(stream) == 0 {
		return nil, nil
	}

	// Flatten each row
	flatStream := make([]map[string]any, len(stream))
	for i, row := range stream {
		if rowMap, ok := row.(map[string]any); ok {
			flatStream[i] = e.flatten(rowMap, "", ".", 0, 0, visited)
		} else {
			flatStream[i] = make(map[string]any)
		}
	}

	// Get all column names
	colSet := make(map[string]bool)
	for _, d := range flatStream {
		for k := range d {
			colSet[k] = true
		}
	}

	cols := make([]string, 0, len(colSet))
	for k := range colSet {
		cols = append(cols, k)
	}
	sort.Strings(cols)

	// Analyze column sparsity
	columnStats := e.analyzeColumnSparsity(flatStream, cols)
	var coreColumns, optionalColumns []string
	for _, stat := range columnStats {
		if stat.presence >= 0.7 {
			coreColumns = append(coreColumns, stat.name)
		} else {
			optionalColumns = append(optionalColumns, stat.name)
		}
	}

	// Decide encoding strategy
	useSparseEncoding := len(optionalColumns) > 0 && len(optionalColumns) <= 5

	if useSparseEncoding {
		return e.writeSparseTable(flatStream, coreColumns, optionalColumns, len(stream), key, visited)
	}
	return e.writeStandardTable(flatStream, cols, len(stream), key, visited)
}

// writeStandardTable writes a standard compact table.
func (e *Encoder) writeStandardTable(flatStream []map[string]any, cols []string, rowCount int, key string, visited map[uintptr]bool) ([]string, error) {
	var lines []string

	// Build header
	var header string
	if key != "" && key != "data" {
		header = fmt.Sprintf("%s%c%c(%d)", key, MetaSeparator, TableMarker, rowCount)
	} else {
		header = fmt.Sprintf("%c%d", TableMarker, rowCount)
	}

	header += string(MetaSeparator) + strings.Join(cols, ",")
	lines = append(lines, header)

	// Write rows
	for _, row := range flatStream {
		var tokens []string
		for _, col := range cols {
			val, ok := row[col]
			if !ok || val == nil {
				tokens = append(tokens, "null")
			} else {
				valStr, err := e.formatValue(val, visited)
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, valStr)
			}
		}
		lines = append(lines, strings.Join(tokens, ","))
	}

	return lines, nil
}

// writeSparseTable writes a sparse table for semi-uniform data.
func (e *Encoder) writeSparseTable(flatStream []map[string]any, coreColumns, optionalColumns []string, rowCount int, key string, visited map[uintptr]bool) ([]string, error) {
	var lines []string

	// Build header
	var header string
	if key != "" && key != "data" {
		header = fmt.Sprintf("%s%c%c(%d)", key, MetaSeparator, TableMarker, rowCount)
	} else {
		header = fmt.Sprintf("%c%d", TableMarker, rowCount)
	}

	header += string(MetaSeparator) + strings.Join(coreColumns, ",")
	lines = append(lines, header)

	// Write rows
	for _, row := range flatStream {
		var tokens []string

		// Core columns
		for _, col := range coreColumns {
			valStr, _ := e.formatValue(row[col], visited)
			tokens = append(tokens, valStr)
		}

		// Optional columns as key:value if present
		for _, col := range optionalColumns {
			if val, ok := row[col]; ok && val != nil {
				valStr, _ := e.formatValue(val, visited)
				tokens = append(tokens, fmt.Sprintf("%s:%s", col, valStr))
			}
		}

		lines = append(lines, strings.Join(tokens, ","))
	}

	return lines, nil
}

// columnStat holds column sparsity information.
type columnStat struct {
	name     string
	presence float64
}

// analyzeColumnSparsity analyzes column sparsity to determine core vs optional.
func (e *Encoder) analyzeColumnSparsity(data []map[string]any, cols []string) []columnStat {
	stats := make([]columnStat, len(cols))
	for i, col := range cols {
		presenceCount := 0
		for _, row := range data {
			if val, ok := row[col]; ok && val != nil {
				presenceCount++
			}
		}
		stats[i] = columnStat{
			name:     col,
			presence: float64(presenceCount) / float64(len(data)),
		}
	}
	return stats
}

// calculateIrregularity calculates the irregularity score for array of objects.
func (e *Encoder) calculateIrregularity(data []any) float64 {
	if len(data) == 0 {
		return 0
	}

	// Get all unique keys across all objects
	allKeys := make(map[string]bool)
	keySets := make([]map[string]bool, len(data))

	for i, item := range data {
		if obj, ok := item.(map[string]any); ok {
			keySet := make(map[string]bool)
			for k := range obj {
				keySet[k] = true
				allKeys[k] = true
			}
			keySets[i] = keySet
		} else {
			keySets[i] = make(map[string]bool)
		}
	}

	if len(allKeys) == 0 {
		return 0
	}

	// Calculate key overlap score
	var totalOverlap float64
	var comparisons int

	for i := 0; i < len(keySets); i++ {
		for j := i + 1; j < len(keySets); j++ {
			keys1 := keySets[i]
			keys2 := keySets[j]

			// Count shared keys
			shared := 0
			for k := range keys1 {
				if keys2[k] {
					shared++
				}
			}

			// Jaccard similarity
			union := len(keys1) + len(keys2) - shared
			var similarity float64
			if union > 0 {
				similarity = float64(shared) / float64(union)
			} else {
				similarity = 1
			}

			totalOverlap += similarity
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0
	}

	avgSimilarity := totalOverlap / float64(comparisons)
	return 1 - avgSimilarity
}

// csvQuote quotes a string for CSV (RFC 4180).
func (e *Encoder) csvQuote(s string) string {
	escaped := strings.ReplaceAll(s, `"`, `""`)
	return `"` + escaped + `"`
}

// formatZonNode formats nested structure using ZON syntax.
func (e *Encoder) formatZonNode(val any, visited map[uintptr]bool) (string, error) {
	switch v := val.(type) {
	case nil:
		return "null", nil
	case bool:
		if v {
			return "T", nil
		}
		return "F", nil
	case float64:
		return e.formatNumber(v), nil
	case float32:
		return e.formatNumber(float64(v)), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case string:
		return e.formatString(v), nil
	case map[string]any:
		return e.formatObject(v, visited)
	case []any:
		return e.formatArray(v, visited)
	default:
		// Try JSON marshaling for other types
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// formatObject formats a map as ZON object.
func (e *Encoder) formatObject(obj map[string]any, visited map[uintptr]bool) (string, error) {
	if len(obj) == 0 {
		return "{}", nil
	}

	// Sort keys
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []string
	for _, k := range keys {
		// Format key
		kStr := k
		if regexp.MustCompile(`[,:\{\}\[\]"]`).MatchString(kStr) {
			b, _ := json.Marshal(kStr)
			kStr = string(b)
		}

		// Format value recursively
		vStr, err := e.formatZonNode(obj[k], visited)
		if err != nil {
			return "", err
		}

		// Colon-less Objects/Arrays
		if strings.HasPrefix(vStr, "{") || strings.HasPrefix(vStr, "[") {
			items = append(items, kStr+vStr)
		} else {
			items = append(items, kStr+":"+vStr)
		}
	}
	return "{" + strings.Join(items, ",") + "}", nil
}

// formatArray formats a slice as ZON array.
func (e *Encoder) formatArray(arr []any, visited map[uintptr]bool) (string, error) {
	if len(arr) == 0 {
		return "[]", nil
	}

	var items []string
	for _, item := range arr {
		itemStr, err := e.formatZonNode(item, visited)
		if err != nil {
			return "", err
		}
		items = append(items, itemStr)
	}
	return "[" + strings.Join(items, ",") + "]", nil
}

// formatNumber formats a number canonically without scientific notation.
func (e *Encoder) formatNumber(n float64) string {
	// Handle special values
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "null"
	}

	// Check if integer
	if n == math.Trunc(n) {
		return strconv.FormatInt(int64(n), 10)
	}

	// Format float without scientific notation
	s := strconv.FormatFloat(n, 'f', -1, 64)

	// Ensure decimal point for floats
	if !strings.Contains(s, ".") {
		s += ".0"
	}

	return s
}

// formatString formats a string with minimal quoting.
func (e *Encoder) formatString(s string) string {
	// Always JSON-stringify strings with newlines
	if strings.Contains(s, "\n") || strings.Contains(s, "\r") {
		b, _ := json.Marshal(s)
		return string(b)
	}

	// ISO Date Detection
	if e.isISODate(s) {
		return s
	}

	// Check if needs type protection
	if e.needsTypeProtection(s) {
		b, _ := json.Marshal(s)
		return string(b)
	}

	// Quote empty strings or whitespace-only strings
	if strings.TrimSpace(s) == "" {
		b, _ := json.Marshal(s)
		return string(b)
	}

	// Quote if contains structural delimiters
	if regexp.MustCompile(`[,\{\}\[\]"]`).MatchString(s) {
		b, _ := json.Marshal(s)
		return string(b)
	}

	return s
}

// formatValue formats a value with minimal quoting.
func (e *Encoder) formatValue(val any, visited map[uintptr]bool) (string, error) {
	if val == nil {
		return "null", nil
	}

	switch v := val.(type) {
	case bool:
		if v {
			return "T", nil
		}
		return "F", nil
	case float64:
		return e.formatNumber(v), nil
	case float32:
		return e.formatNumber(float64(v)), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case string:
		return e.formatTableCellString(v), nil
	case map[string]any, []any:
		return e.formatZonNode(v, visited)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// formatTableCellString formats a string for use in a table cell.
func (e *Encoder) formatTableCellString(s string) string {
	// Always JSON-stringify strings with newlines
	if strings.Contains(s, "\n") || strings.Contains(s, "\r") {
		b, _ := json.Marshal(s)
		return e.csvQuote(string(b))
	}

	// ISO Date Detection
	if e.isISODate(s) {
		return s
	}

	// Check if needs type protection
	if e.needsTypeProtection(s) {
		b, _ := json.Marshal(s)
		return e.csvQuote(string(b))
	}

	// Check if it needs CSV quoting (delimiters)
	if e.needsQuotes(s) {
		return e.csvQuote(s)
	}

	return s
}

// isISODate checks if string is an ISO 8601 date/datetime.
func (e *Encoder) isISODate(s string) bool {
	// ISO 8601 full datetime with timezone
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})$`, s); matched {
		return true
	}
	// ISO 8601 date only
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, s); matched {
		return true
	}
	// Simple time
	if matched, _ := regexp.MatchString(`^\d{2}:\d{2}:\d{2}$`, s); matched {
		return true
	}
	return false
}

// needsTypeProtection determines if string needs quoting to preserve as string.
func (e *Encoder) needsTypeProtection(s string) bool {
	sLower := strings.ToLower(s)

	// Reserved words
	reserved := []string{"t", "f", "true", "false", "null", "none", "nil"}
	for _, r := range reserved {
		if sLower == r {
			return true
		}
	}

	// Gas/Liquid tokens
	if s == string(GasToken) || s == string(LiquidToken) {
		return true
	}

	// Leading/trailing whitespace must be preserved
	if strings.TrimSpace(s) != s {
		return true
	}

	// Control characters need JSON escaping
	for _, c := range s {
		if c < 32 {
			return true
		}
	}

	// Pure integer
	if matched, _ := regexp.MatchString(`^-?\d+$`, s); matched {
		return true
	}

	// Pure decimal
	if matched, _ := regexp.MatchString(`^-?\d+\.\d+$`, s); matched {
		return true
	}

	// Scientific notation
	if matched, _ := regexp.MatchString(`(?i)^-?\d+(\.\d+)?e[+-]?\d+$`, s); matched {
		return true
	}

	// If it starts/ends with digit but has non-numeric chars, check carefully
	if len(s) > 0 && ((s[0] >= '0' && s[0] <= '9') || (s[len(s)-1] >= '0' && s[len(s)-1] <= '9')) {
		// Try parsing - if it parses cleanly and matches, it's a number
		if n, err := strconv.ParseFloat(s, 64); err == nil && strconv.FormatFloat(n, 'f', -1, 64) == s {
			return true
		}
	}

	return false
}

// needsQuotes determines if a string needs quotes.
func (e *Encoder) needsQuotes(s string) bool {
	if s == "" {
		return true
	}

	// Reserved tokens need quoting
	reserved := []string{"T", "F", "null", string(GasToken), string(LiquidToken)}
	for _, r := range reserved {
		if s == r {
			return true
		}
	}

	// Quote if it looks like a number
	if matched, _ := regexp.MatchString(`^-?\d+$`, s); matched {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	// Quote if leading/trailing whitespace
	if strings.TrimSpace(s) != s {
		return true
	}

	// Only quote if contains delimiter or control chars
	if regexp.MustCompile(`[,\n\r\t"\[\]|;]`).MatchString(s) {
		return true
	}

	return false
}

// flatten flattens nested dictionary with depth limit.
func (e *Encoder) flatten(d map[string]any, parent, sep string, maxDepth, currentDepth int, visited map[uintptr]bool) map[string]any {
	result := make(map[string]any)

	for k, v := range d {
		newKey := k
		if parent != "" {
			newKey = parent + sep + k
		}

		if nested, ok := v.(map[string]any); ok && currentDepth < maxDepth {
			// Recursively flatten this level
			flattened := e.flatten(nested, newKey, sep, maxDepth, currentDepth+1, visited)
			for fk, fv := range flattened {
				result[fk] = fv
			}
		} else {
			result[newKey] = v
		}
	}

	return result
}

// Encode encodes data to ZON format using a default encoder.
func Encode(data any) (string, error) {
	return NewEncoder(DefaultAnchorInterval).Encode(data)
}
