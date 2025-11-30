# ZON API Reference

Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)

Complete API documentation for `zon-go` v1.0.5.

## Installation

```bash
go get github.com/ZON-Format/zon-go
```

---

## Encoding Functions

### `Encode(data any) (string, error)`

Encodes Go data to ZON format.

**Parameters:**
- `data` (`any`) - Go data to encode (maps, slices, primitives)

**Returns:** `(string, error)` - ZON-formatted string or error

**Example:**
```go
import zon "github.com/ZON-Format/zon-go"

data := map[string]any{
    "users": []any{
        map[string]any{"id": 1, "name": "Alice", "active": true},
        map[string]any{"id": 2, "name": "Bob", "active": false},
    },
}

encoded, err := zon.Encode(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(encoded)
```

**Output:**
```zon
users:@(2):active,id,name
T,1,Alice
F,2,Bob
```

**Supported Types:**
- ✅ Maps (`map[string]any`)
- ✅ Slices (`[]any`)
- ✅ Strings
- ✅ Numbers (int, int64, float64)
- ✅ Booleans (`T`/`F`)
- ✅ Nil (`null`)

**Encoding Behavior:**
- **Uniform arrays** → Table format (`@(N):columns`)
- **Nested objects** → Quoted notation (`"{key:value}"`)
- **Primitive arrays** → Inline format (`"[a,b,c]"`)
- **Booleans** → `T`/`F` (single character)
- **Null** → `null`

---

## Decoding Functions

### `Decode(zonString string) (any, error)`

Decodes a ZON format string back to the original Go data structure.
Uses strict mode by default.

**Parameters:**
- `zonString` (`string`): The ZON-formatted string to decode

**Returns:** `(any, error)` - Decoded data or error

**Example:**
```go
import zon "github.com/ZON-Format/zon-go"

zonData := `users:@(2):id,name
1,Alice
2,Bob`

decoded, err := zon.Decode(zonData)
if err != nil {
    log.Fatal(err)
}
fmt.Println(decoded)
```

**Output:**
```go
map[string]any{
    "users": []any{
        map[string]any{"id": 1, "name": "Alice"},
        map[string]any{"id": 2, "name": "Bob"},
    },
}
```

### `DecodeWithOptions(zonString string, options *DecodeOptions) (any, error)`

Decodes ZON with custom options.

**Parameters:**
- `zonString` (`string`): The ZON-formatted string to decode
- `options` (`*DecodeOptions`): Decoding options

**DecodeOptions:**
```go
type DecodeOptions struct {
    Strict bool // Enable strict validation (default: true)
}
```

**Example:**
```go
// Non-strict mode - allows row/field count mismatches
decoded, err := zon.DecodeWithOptions(zonData, &zon.DecodeOptions{Strict: false})
```

---

## Error Types

### `DecodeError`

Error type for decoding failures and strict mode validation errors.

**Fields:**
- `Code` (`string`): Error code (e.g., "E001", "E002")
- `Message` (`string`): Error description
- `Line` (`int`): Line number (0 if unknown)
- `Column` (`int`): Column position (0 if unknown)
- `Context` (`string`): Context snippet

**Example:**
```go
import zon "github.com/ZON-Format/zon-go"

decoded, err := zon.Decode(invalidZon)
if err != nil {
    if decodeErr, ok := err.(*zon.DecodeError); ok {
        fmt.Println(decodeErr.Code)     // "E001"
        fmt.Println(decodeErr.Message)  // "Row count mismatch..."
        fmt.Println(decodeErr.Context)  // "Table: users"
    }
}
```

### Common Error Codes

| Code | Description | Example |
|------|-------------|---------|
| `E001` | Row count mismatch | Declared `@(3)` but only 2 rows provided |
| `E002` | Field count mismatch | Declared 3 columns but row has 2 values |
| `E301` | Document size exceeds 100MB | Prevents memory exhaustion |
| `E302` | Line length exceeds 1MB | Prevents buffer overflow |
| `E303` | Array length exceeds 1M items | Prevents excessive iteration |
| `E304` | Object key count exceeds 100K | Prevents hash collision |

---

## Constants

### Security Limits

```go
const (
    MaxDocumentSize = 100 * 1024 * 1024  // 100 MB
    MaxLineLength   = 1024 * 1024        // 1 MB
    MaxArrayLength  = 1_000_000          // 1 million items
    MaxObjectKeys   = 100_000            // 100K keys
    MaxNestingDepth = 100                // 100 levels
)
```

### Error Codes

```go
const (
    ErrRowCountMismatch   = "E001"
    ErrFieldCountMismatch = "E002"
    ErrDocumentTooLarge   = "E301"
    ErrLineTooLong        = "E302"
    ErrArrayTooLarge      = "E303"
    ErrTooManyKeys        = "E304"
)
```

---

## Complete Examples

### Example 1: Simple Object

```go
data := map[string]any{
    "name":    "ZON Format",
    "version": "1.0.5",
    "active":  true,
    "score":   98.5,
}

encoded, _ := zon.Encode(data)
// active:T
// name:ZON Format
// score:98.5
// version:"1.0.5"

decoded, _ := zon.Decode(encoded)
// map[string]any{"name": "ZON Format", "version": "1.0.5", "active": true, "score": 98.5}
```

### Example 2: Uniform Table

```go
data := map[string]any{
    "employees": []any{
        map[string]any{"id": 1, "name": "Alice", "dept": "Eng", "salary": 85000},
        map[string]any{"id": 2, "name": "Bob", "dept": "Sales", "salary": 72000},
        map[string]any{"id": 3, "name": "Carol", "dept": "HR", "salary": 65000},
    },
}

encoded, _ := zon.Encode(data)
// employees:@(3):dept,id,name,salary
// Eng,1,Alice,85000
// Sales,2,Bob,72000
// HR,3,Carol,65000

decoded, _ := zon.Decode(encoded)
// Identical to original!
```

### Example 3: Mixed Structure

```go
data := map[string]any{
    "metadata": map[string]any{"version": "1.0", "env": "prod"},
    "users": []any{
        map[string]any{"id": 1, "name": "Alice"},
        map[string]any{"id": 2, "name": "Bob"},
    },
    "tags": []any{"nodejs", "typescript", "llm"},
}

encoded, _ := zon.Encode(data)
decoded, _ := zon.Decode(encoded)
// Identical to original!
```

---

## Round-Trip Compatibility

ZON **guarantees lossless round-trips**:

```go
import zon "github.com/ZON-Format/zon-go"

func testRoundTrip(data any) bool {
    encoded, err := zon.Encode(data)
    if err != nil {
        return false
    }
    decoded, err := zon.Decode(encoded)
    if err != nil {
        return false
    }
    return reflect.DeepEqual(data, decoded)
}

// All these pass:
testRoundTrip(map[string]any{"name": "test", "value": 123})  // ✅
testRoundTrip([]any{1, 2, 3, 4, 5})                          // ✅
testRoundTrip([]any{map[string]any{"id": 1}})                // ✅
testRoundTrip(nil)                                            // ✅
testRoundTrip("hello")                                        // ✅
```

**Verified:**
- ✅ 94/94 unit tests pass
- ✅ All datasets verified
- ✅ Zero data loss

---

## CLI Usage

Build and install:
```bash
go install github.com/ZON-Format/zon-go/cmd/zon@latest
```

Commands:
```bash
# Encode JSON to ZON
zon encode data.json > data.zonf

# Decode ZON to JSON
zon decode data.zonf > output.json
```

---

## See Also

- [Syntax Cheatsheet](./syntax-cheatsheet.md) - Quick reference
- [Format Specification](../SPEC.md) - Formal grammar
- [LLM Best Practices](./llm-best-practices.md) - Usage guide
- [GitHub Repository](https://github.com/ZON-Format/zon-go)
