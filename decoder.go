// Package zon provides ZON encoding and decoding functionality.
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
package zon

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// DecodeOptions contains options for decoding.
type DecodeOptions struct {
	Strict bool // Enable strict validation (default: true)
}

// Decoder decodes ZON format to Go data structures.
type Decoder struct {
	strict      bool
	currentLine int
}

// NewDecoder creates a new ZON decoder with the given options.
func NewDecoder(options *DecodeOptions) *Decoder {
	strict := true
	if options != nil {
		strict = options.Strict
	}
	return &Decoder{
		strict:      strict,
		currentLine: 0,
	}
}

// tableInfo holds information about a table being parsed.
type tableInfo struct {
	cols         []string
	omittedCols  []string
	rows         []map[string]any
	prevVals     map[string]any
	rowIndex     int
	expectedRows int
}

// Decode decodes ZON format string to original data structure.
func (d *Decoder) Decode(zonStr string) (any, error) {
	if zonStr == "" {
		return map[string]any{}, nil
	}

	// Security: Check document size
	if len(zonStr) > MaxDocumentSize {
		return nil, NewDecodeError(
			ErrDocumentTooLarge,
			"Document size exceeds maximum (100MB)",
			0, "",
		)
	}

	lines := strings.Split(strings.TrimSpace(zonStr), "\n")
	if len(lines) == 0 {
		return map[string]any{}, nil
	}

	// Special case: Root-level ZON list
	if len(lines) == 1 {
		line := strings.TrimSpace(lines[0])
		if strings.HasPrefix(line, "[") {
			return d.parseZonNode(line, 0)
		}

		// Check for colon-less object/array pattern
		hasBlock := regexp.MustCompile(`^[a-zA-Z0-9_]+\s*[\{\[]`).MatchString(line)
		if !strings.Contains(line, string(MetaSeparator)) && !strings.HasPrefix(line, string(TableMarker)) && !hasBlock {
			return d.parsePrimitive(line), nil
		}
	}

	// Main decode loop
	metadata := make(map[string]any)
	tables := make(map[string]*tableInfo)
	var currentTable *tableInfo
	var currentTableName string

	for _, line := range lines {
		trimmedLine := strings.TrimRight(line, " \t\r")

		// Security: Check line length
		if len(trimmedLine) > MaxLineLength {
			return nil, NewDecodeError(
				ErrLineTooLong,
				"Line length exceeds maximum (1MB)",
				d.currentLine, "",
			)
		}

		// Skip blank lines
		if trimmedLine == "" {
			continue
		}

		// Table header (Anonymous or Legacy): @...
		if strings.HasPrefix(trimmedLine, string(TableMarker)) {
			tableName, tInfo, err := d.parseTableHeader(trimmedLine)
			if err != nil {
				return nil, err
			}
			currentTableName = tableName
			currentTable = tInfo
			tables[currentTableName] = currentTable
		} else if currentTable != nil && currentTable.rowIndex < currentTable.expectedRows {
			// Table row
			row, err := d.parseTableRow(trimmedLine, currentTable)
			if err != nil {
				return nil, err
			}
			currentTable.rows = append(currentTable.rows, row)

			// If we've read all rows, exit table mode
			if currentTable.rowIndex >= currentTable.expectedRows {
				currentTable = nil
			}
		} else {
			// Metadata line OR Named Table
			splitIdx, splitChar := d.findSplitPoint(trimmedLine)

			if splitIdx != -1 {
				var key, val string
				if splitChar == ':' {
					key = strings.TrimSpace(trimmedLine[:splitIdx])
					val = strings.TrimSpace(trimmedLine[splitIdx+1:])
				} else {
					// Split at { or [ (include it in value)
					key = strings.TrimSpace(trimmedLine[:splitIdx])
					val = strings.TrimSpace(trimmedLine[splitIdx:])
				}

				// Check if it's a named table start
				if strings.HasPrefix(val, string(TableMarker)) {
					_, tInfo, err := d.parseTableHeader(val)
					if err != nil {
						return nil, err
					}
					currentTableName = key
					currentTable = tInfo
					tables[currentTableName] = currentTable
				} else {
					currentTable = nil
					parsedVal, err := d.parseValue(val)
					if err != nil {
						return nil, err
					}
					metadata[key] = parsedVal
				}
			}
		}
	}

	// Recombine tables into metadata
	for tableName, table := range tables {
		// Strict mode: validate row count
		if d.strict && len(table.rows) != table.expectedRows {
			return nil, NewDecodeError(
				ErrRowCountMismatch,
				"Row count mismatch in table '"+tableName+"': expected "+strconv.Itoa(table.expectedRows)+", got "+strconv.Itoa(len(table.rows)),
				0, "Table: "+tableName,
			)
		}
		metadata[tableName] = d.reconstructTable(table)
	}

	// Unflatten dotted keys
	result := d.unflatten(metadata)

	// Unwrap pure lists: if only key is 'data', return the list directly
	if resultMap, ok := result.(map[string]any); ok {
		if len(resultMap) == 1 {
			if data, ok := resultMap["data"]; ok {
				if arr, ok := data.([]any); ok {
					return arr, nil
				}
			}
		}
	}

	return result, nil
}

// findSplitPoint finds the split point in a metadata line.
func (d *Decoder) findSplitPoint(line string) (int, byte) {
	splitIdx := -1
	var splitChar byte
	depth := 0
	inQuote := false

	for i := 0; i < len(line); i++ {
		char := line[i]
		if char == '"' {
			inQuote = !inQuote
		}
		if !inQuote {
			if char == '{' || char == '[' {
				depth++
				if depth == 1 && splitIdx == -1 {
					// We just entered a block
					splitIdx = i
					splitChar = char
					break
				}
			}
			if char == '}' || char == ']' {
				depth--
			}
			if char == ':' && depth == 0 {
				splitIdx = i
				splitChar = ':'
				break
			}
		}
	}

	return splitIdx, splitChar
}

// parseTableHeader parses a table header line.
func (d *Decoder) parseTableHeader(line string) (string, *tableInfo, error) {
	// Try v2.0 format with name: @name(count)[col][col]:columns
	v2NamedPattern := regexp.MustCompile(`^@(\w+)\((\d+)\)(\[\w+\])*:(.+)$`)
	if matches := v2NamedPattern.FindStringSubmatch(line); matches != nil {
		tableName := matches[1]
		count, _ := strconv.Atoi(matches[2])
		omittedStr := matches[3]
		colsStr := matches[4]

		var omittedCols []string
		if omittedStr != "" {
			omittedPattern := regexp.MustCompile(`\[(\w+)\]`)
			omittedMatches := omittedPattern.FindAllStringSubmatch(omittedStr, -1)
			for _, m := range omittedMatches {
				omittedCols = append(omittedCols, m[1])
			}
		}

		cols := strings.Split(colsStr, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}

		return tableName, &tableInfo{
			cols:         cols,
			omittedCols:  omittedCols,
			rows:         []map[string]any{},
			prevVals:     make(map[string]any),
			rowIndex:     0,
			expectedRows: count,
		}, nil
	}

	// Try v2.1 format (anonymous/value): @(count)[col]:columns
	v2ValuePattern := regexp.MustCompile(`^@\((\d+)\)(\[\w+\])*:(.+)$`)
	if matches := v2ValuePattern.FindStringSubmatch(line); matches != nil {
		count, _ := strconv.Atoi(matches[1])
		omittedStr := matches[2]
		colsStr := matches[3]

		var omittedCols []string
		if omittedStr != "" {
			omittedPattern := regexp.MustCompile(`\[(\w+)\]`)
			omittedMatches := omittedPattern.FindAllStringSubmatch(omittedStr, -1)
			for _, m := range omittedMatches {
				omittedCols = append(omittedCols, m[1])
			}
		}

		cols := strings.Split(colsStr, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}

		return "data", &tableInfo{
			cols:         cols,
			omittedCols:  omittedCols,
			rows:         []map[string]any{},
			prevVals:     make(map[string]any),
			rowIndex:     0,
			expectedRows: count,
		}, nil
	}

	// Try v2.0 format (anonymous): @count[col][col]:columns
	v2Pattern := regexp.MustCompile(`^@(\d+)(\[\w+\])*:(.+)$`)
	if matches := v2Pattern.FindStringSubmatch(line); matches != nil {
		count, _ := strconv.Atoi(matches[1])
		omittedStr := matches[2]
		colsStr := matches[3]

		var omittedCols []string
		if omittedStr != "" {
			omittedPattern := regexp.MustCompile(`\[(\w+)\]`)
			omittedMatches := omittedPattern.FindAllStringSubmatch(omittedStr, -1)
			for _, m := range omittedMatches {
				omittedCols = append(omittedCols, m[1])
			}
		}

		cols := strings.Split(colsStr, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}

		return "data", &tableInfo{
			cols:         cols,
			omittedCols:  omittedCols,
			rows:         []map[string]any{},
			prevVals:     make(map[string]any),
			rowIndex:     0,
			expectedRows: count,
		}, nil
	}

	// Fallback to v1.x format: @tablename(count):cols
	v1Pattern := regexp.MustCompile(`^@(\w+)\((\d+)\):(.+)$`)
	if matches := v1Pattern.FindStringSubmatch(line); matches != nil {
		tableName := matches[1]
		count, _ := strconv.Atoi(matches[2])
		colsStr := matches[3]

		cols := strings.Split(colsStr, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}

		return tableName, &tableInfo{
			cols:         cols,
			rows:         []map[string]any{},
			prevVals:     make(map[string]any),
			rowIndex:     0,
			expectedRows: count,
		}, nil
	}

	return "", nil, NewDecodeError("E003", "Invalid table header: "+line, 0, "")
}

// parseTableRow parses a table row.
func (d *Decoder) parseTableRow(line string, table *tableInfo) (map[string]any, error) {
	tokens := d.splitByDelimiter(line, ',')

	// Strict mode: validate field count
	coreFieldCount := len(tokens)
	sparseFieldCount := 0

	// Count sparse fields
	for i := len(table.cols); i < len(tokens); i++ {
		tok := tokens[i]
		if strings.Contains(tok, ":") && !d.isURL(tok) && !d.isTimestamp(tok) {
			sparseFieldCount++
		}
	}

	// In strict mode, core fields must match column count
	if d.strict && coreFieldCount < len(table.cols) && sparseFieldCount == 0 {
		return nil, NewDecodeError(
			ErrFieldCountMismatch,
			"Field count mismatch on row "+strconv.Itoa(table.rowIndex+1)+": expected "+strconv.Itoa(len(table.cols))+" fields, got "+strconv.Itoa(coreFieldCount),
			d.currentLine,
			truncateString(line, 50),
		)
	}

	// Pad if needed
	for len(tokens) < len(table.cols) {
		tokens = append(tokens, "")
	}

	row := make(map[string]any)
	tokenIdx := 0

	// Parse core columns
	for _, col := range table.cols {
		if tokenIdx < len(tokens) {
			tok := tokens[tokenIdx]
			val, err := d.parseValue(tok)
			if err != nil {
				return nil, err
			}
			row[col] = val
			tokenIdx++
		}
	}

	// Parse optional fields (sparse encoding)
	for tokenIdx < len(tokens) {
		tok := tokens[tokenIdx]
		if strings.Contains(tok, ":") && !d.isURL(tok) && !d.isTimestamp(tok) {
			colonIdx := strings.Index(tok, ":")
			key := strings.TrimSpace(tok[:colonIdx])
			val := strings.TrimSpace(tok[colonIdx+1:])

			// Validate key is a simple identifier
			if regexp.MustCompile(`^[a-zA-Z_]\w*$`).MatchString(key) {
				parsedVal, err := d.parseValue(val)
				if err != nil {
					return nil, err
				}
				row[key] = parsedVal
			}
		}
		tokenIdx++
	}

	// Reconstruct omitted sequential columns
	for _, col := range table.omittedCols {
		row[col] = table.rowIndex + 1
	}

	table.rowIndex++
	return row, nil
}

// isURL checks if string is a URL.
func (d *Decoder) isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "/")
}

// isTimestamp checks if string is a timestamp with colons.
func (d *Decoder) isTimestamp(s string) bool {
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, s); matched {
		return true
	}
	if matched, _ := regexp.MatchString(`^\d{2}:\d{2}:\d{2}`, s); matched {
		return true
	}
	return false
}

// reconstructTable reconstructs table from parsed rows.
func (d *Decoder) reconstructTable(table *tableInfo) []any {
	result := make([]any, len(table.rows))
	for i, row := range table.rows {
		result[i] = d.unflatten(row)
	}
	return result
}

// parseZonNode parses a ZON nested format.
func (d *Decoder) parseZonNode(text string, depth int) (any, error) {
	if depth > MaxNestingDepth {
		return nil, NewDecodeError("", "Maximum nesting depth exceeded (100)", 0, "")
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, nil
	}

	// Dict: {k:v,k:v}
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		content := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if content == "" {
			return map[string]any{}, nil
		}

		obj := make(map[string]any)
		pairs := d.splitByDelimiter(content, ',')

		// Security: Check object key count
		if len(pairs) > MaxObjectKeys {
			return nil, NewDecodeError(
				ErrTooManyKeys,
				"Object key count exceeds maximum (100K keys)",
				0, "",
			)
		}

		for _, pair := range pairs {
			keyStr, valStr := d.findKeyValueSplit(pair)
			if keyStr == "" && valStr == "" {
				continue
			}

			key := d.parsePrimitive(keyStr)
			val, err := d.parseZonNode(valStr, depth+1)
			if err != nil {
				return nil, err
			}
			if keyStr, ok := key.(string); ok {
				obj[keyStr] = val
			} else {
				obj[strings.TrimSpace(pair)] = nil
			}
		}

		return obj, nil
	}

	// List: [v,v]
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		content := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if content == "" {
			return []any{}, nil
		}

		items := d.splitByDelimiter(content, ',')

		// Security: Check array length
		if len(items) > MaxArrayLength {
			return nil, NewDecodeError(
				ErrArrayTooLarge,
				"Array length exceeds maximum (1M items)",
				0, "",
			)
		}

		result := make([]any, len(items))
		for i, item := range items {
			val, err := d.parseZonNode(item, depth+1)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	}

	// Leaf node (primitive)
	return d.parsePrimitive(trimmed), nil
}

// findKeyValueSplit finds the key-value split point in a pair.
func (d *Decoder) findKeyValueSplit(pair string) (string, string) {
	splitIdx := -1
	var splitChar byte
	inQuote := false
	var quoteChar byte
	depth := 0

	for i := 0; i < len(pair); i++ {
		char := pair[i]

		if char == '\\' && i+1 < len(pair) {
			i++
			continue
		}

		if char == '"' || char == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if char == quoteChar {
				inQuote = false
			}
		} else if !inQuote {
			if char == ':' {
				if depth == 0 {
					splitIdx = i
					splitChar = ':'
					break
				}
			} else if char == '{' || char == '[' {
				if depth == 0 && splitIdx == -1 {
					splitIdx = i
					splitChar = char
					break
				}
				depth++
			} else if char == '}' || char == ']' {
				depth--
			}
		}
	}

	if splitIdx != -1 {
		if splitChar == ':' {
			return strings.TrimSpace(pair[:splitIdx]), strings.TrimSpace(pair[splitIdx+1:])
		}
		// Split at { or [ (include it in value)
		return strings.TrimSpace(pair[:splitIdx]), strings.TrimSpace(pair[splitIdx:])
	}

	return "", ""
}

// splitByDelimiter splits text by delimiter, respecting quotes and nesting.
func (d *Decoder) splitByDelimiter(text string, delim rune) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	var quoteChar rune
	depth := 0

	for i, char := range text {
		// Handle escaped characters
		if char == '\\' && i+1 < len(text) {
			current.WriteRune(char)
			// We'll get the next character in the next iteration
			continue
		}

		if i > 0 && text[i-1] == '\\' {
			current.WriteRune(char)
			continue
		}

		if char == '"' || char == '\'' {
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if char == quoteChar {
				inQuote = false
				quoteChar = 0
			}
			current.WriteRune(char)
		} else if !inQuote {
			if char == '{' || char == '[' {
				depth++
				current.WriteRune(char)
			} else if char == '}' || char == ']' {
				depth--
				current.WriteRune(char)
			} else if char == delim && depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		} else {
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parsePrimitive parses a primitive value.
func (d *Decoder) parsePrimitive(val string) any {
	trimmed := strings.TrimSpace(val)
	valLower := strings.ToLower(trimmed)

	// Booleans
	if valLower == "t" || valLower == "true" {
		return true
	}
	if valLower == "f" || valLower == "false" {
		return false
	}

	// Null
	if valLower == "null" || valLower == "none" || valLower == "nil" {
		return nil
	}

	// Quoted string (JSON style)
	if strings.HasPrefix(trimmed, `"`) {
		var s string
		if err := json.Unmarshal([]byte(trimmed), &s); err == nil {
			return s
		}
	}

	// Try number
	if trimmed != "" {
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return float64(i)
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f
		}
	}

	// String
	return trimmed
}

// parseValue parses a cell value.
func (d *Decoder) parseValue(val string) (any, error) {
	trimmed := strings.TrimSpace(val)

	// Quoted string (JSON style)
	if strings.HasPrefix(trimmed, `"`) {
		var decoded any
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			// If decoded value is a string that looks like a ZON structure, parse it recursively
			if s, ok := decoded.(string); ok {
				stripped := strings.TrimSpace(s)
				if strings.HasPrefix(stripped, "{") || strings.HasPrefix(stripped, "[") {
					return d.parseZonNode(stripped, 0)
				}
			}
			return decoded, nil
		}

		// Fallback: CSV unquoting
		if strings.HasSuffix(trimmed, `"`) {
			unquoted := trimmed[1 : len(trimmed)-1]
			unquoted = strings.ReplaceAll(unquoted, `""`, `"`)

			// Try to parse unquoted value as JSON
			var decoded any
			if err := json.Unmarshal([]byte(unquoted), &decoded); err == nil {
				if s, ok := decoded.(string); ok {
					stripped := strings.TrimSpace(s)
					if strings.HasPrefix(stripped, "{") || strings.HasPrefix(stripped, "[") {
						return d.parseZonNode(stripped, 0)
					}
				}
				return decoded, nil
			}

			// Check for ZON structure in unquoted string
			stripped := strings.TrimSpace(unquoted)
			if strings.HasPrefix(stripped, "{") || strings.HasPrefix(stripped, "[") {
				return d.parseZonNode(stripped, 0)
			}

			return unquoted, nil
		}
	}

	// Booleans (case-insensitive)
	valLower := strings.ToLower(trimmed)
	if valLower == "t" || valLower == "true" {
		return true, nil
	}
	if valLower == "f" || valLower == "false" {
		return false, nil
	}

	// Null
	if valLower == "null" || valLower == "none" || valLower == "nil" {
		return nil, nil
	}

	// Check for ZON-style nested structures
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return d.parseZonNode(trimmed, 0)
	}

	// Try number
	if trimmed != "" {
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return float64(i), nil
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f, nil
		}
	}

	// Double-encoded JSON string fallback
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		var s string
		if err := json.Unmarshal([]byte(trimmed), &s); err == nil {
			return s, nil
		}
	}

	return trimmed, nil
}

// unflatten unflattens dictionary with dotted keys.
func (d *Decoder) unflatten(data map[string]any) any {
	result := make(map[string]any)

	for key, value := range data {
		if !strings.Contains(key, ".") {
			result[key] = value
			continue
		}

		parts := strings.Split(key, ".")

		// SECURITY: Prevent prototype pollution
		skip := false
		for _, p := range parts {
			if p == "__proto__" || p == "constructor" || p == "prototype" {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		target := result
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]

			if _, ok := target[part]; !ok {
				target[part] = make(map[string]any)
			}

			if nested, ok := target[part].(map[string]any); ok {
				target = nested
			} else {
				break
			}
		}

		finalKey := parts[len(parts)-1]
		target[finalKey] = value
	}

	return result
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Decode decodes ZON format string to original data.
// Default is strict mode.
func Decode(data string) (any, error) {
	return NewDecoder(&DecodeOptions{Strict: true}).Decode(data)
}

// DecodeWithOptions decodes ZON format string with custom options.
func DecodeWithOptions(data string, options *DecodeOptions) (any, error) {
	return NewDecoder(options).Decode(data)
}
