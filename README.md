# Zero Overhead Notation (ZON) Format for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/ZON-Format/zon-go.svg)](https://pkg.go.dev/github.com/ZON-Format/zon-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ZON-Format/zon-go)](https://goreportcard.com/report/github.com/ZON-Format/zon-go)
[![Tests](https://img.shields.io/badge/tests-94%2F94%20passing-brightgreen.svg)](#quality--testing)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

# ZON Format → JSON is dead. TOON was cute. ZON just won.
**Zero Overhead Notation** - A compact, human-readable way to encode JSON for LLMs.

**File Extension:** `.zonf` | **Media Type:** `text/zon` | **Encoding:** UTF-8

ZON is a token-efficient serialization format designed for LLM workflows. It achieves 35-50% token reduction vs JSON through tabular encoding, single-character primitives, and intelligent compression while maintaining 100% data fidelity.

Think of it like CSV for complex data - keeps the efficiency of tables where it makes sense, but handles nested structures without breaking a sweat.

**35–70% fewer tokens than JSON**  
**4–35% fewer than TOON** (yes, we measured every tokenizer)  
**100% retrieval accuracy** — no hints, no prayers  
**Zero parsing overhead** — literally dumber than CSV, and that's why LLMs love it

```bash
go get github.com/ZON-Format/zon-go
```

---

## Table of Contents

- [Why ZON?](#why-zon)
- [Key Features](#key-features)
- [Benchmarks](#benchmarks)
- [Installation & Quick Start](#installation--quick-start)
- [Format Overview](#format-overview)
- [API Reference](#api-reference)
- [Documentation](#documentation)

---

## Why ZON?

### Benchmarks

#### Retrieval Accuracy

Benchmarks test LLM comprehension using 309 data retrieval questions on **gpt-5-nano** (Azure OpenAI).

**Dataset Catalog:**

| Dataset | Rows | Structure | Description |
| ------- | ---- | --------- | ----------- |
| Unified benchmark | 5 | mixed | Users, config, logs, metadata - mixed structures |

**Structure:** Mixed uniform tables + nested objects  
**Questions:** 309 total (field retrieval, aggregation, filtering, structure awareness)

#### Efficiency Ranking (Accuracy per 10K Tokens)

Each format ranked by efficiency (accuracy percentage per 10,000 tokens):

```
ZON            ████████████████████ 1430.6 acc%/10K │  99.0% acc │ 692 tokens 👑
CSV            ███████████████████░ 1386.5 acc%/10K │  99.0% acc │ 714 tokens
JSON compact   ████████████████░░░░ 1143.4 acc%/10K │  91.7% acc │ 802 tokens
TOON           ████████████████░░░░ 1132.7 acc%/10K │  99.0% acc │ 874 tokens
JSON           ██████████░░░░░░░░░░  744.6 acc%/10K │  96.8% acc │ 1,300 tokens
```

*Efficiency score = (Accuracy % ÷ Tokens) × 10,000. Higher is better.*

> **TIP:** ZON achieves **99.0% accuracy** while using **20.8% fewer tokens** than TOON and **13.7% fewer** than Minified JSON.

---

## Key Features

- 🎯 **100% LLM Accuracy**: Achieves perfect retrieval (309/309 questions) with self-explanatory structure – no hints needed
- 💾 **Most Token-Efficient**: 4-15% fewer tokens than TOON across all tokenizers
- 🎯 **JSON Data Model**: Encodes the same objects, arrays, and primitives as JSON with deterministic, lossless round-trips
- 📐 **Minimal Syntax**: Explicit headers (`@(N)` for count, column list) eliminate ambiguity for LLMs
- 🧺 **Tabular Arrays**: Uniform arrays collapse into tables that declare fields once and stream row values
- 🔢 **Canonical Numbers**: No scientific notation (1000000, not 1e6), NaN/Infinity → null
- 🌳 **Deep Nesting**: Handles complex nested structures efficiently (91% compression on 50-level deep objects)
- 🔒 **Security Limits**: Automatic DOS prevention (100MB docs, 1M arrays, 100K keys)
- ✅ **Production Ready**: 94/94 tests pass, 27/27 datasets verified, zero data loss

---

## Quality & Security

### Data Integrity
- **Unit tests:** 94/94 passed
- **Roundtrip tests:** All datasets verified
- **No data loss or corruption**

### Security Limits (DOS Prevention)

Automatic protection against malicious input:

| Limit | Maximum | Error Code |
|-------|---------|------------|
| Document size | 100 MB | E301 |
| Line length | 1 MB | E302 |
| Array length | 1M items | E303 |
| Object keys | 100K keys | E304 |
| Nesting depth | 100 levels | - |

**Protection is automatic** - no configuration required.

---

## Installation & Quick Start

### Go Library

```bash
go get github.com/ZON-Format/zon-go
```

**Example usage:**

```go
package main

import (
    "fmt"
    zon "github.com/ZON-Format/zon-go"
)

func main() {
    data := map[string]any{
        "users": []any{
            map[string]any{"id": 1, "name": "Alice", "role": "admin", "active": true},
            map[string]any{"id": 2, "name": "Bob", "role": "user", "active": true},
        },
    }

    encoded, _ := zon.Encode(data)
    fmt.Println(encoded)
    // users:@(2):active,id,name,role
    // T,1,Alice,admin
    // T,2,Bob,user

    // Decode back to data
    decoded, _ := zon.Decode(encoded)
    fmt.Println(decoded)
    // Identical to original - lossless!
}
```

### Command Line Interface (CLI)

Build and install the CLI:

```bash
go install github.com/ZON-Format/zon-go/cmd/zon@latest
```

**Usage:**

```bash
# Encode JSON to ZON format
zon encode data.json > data.zonf

# Decode ZON back to JSON
zon decode data.zonf > output.json
```

**File Extension:**

ZON files conventionally use the `.zonf` extension to distinguish them from other formats.

---

## Format Overview

ZON auto-selects the optimal representation for your data.

### Tabular Arrays

Best for arrays of objects with consistent structure:

```
users:@(3):active,id,name,role
T,1,Alice,Admin
T,2,Bob,User
F,3,Carol,Guest
```

- `@(3)` = row count
- Column names listed once  
- Data rows follow

### Nested Objects

Best for configuration and nested structures:

```
config:"{database:{host:db.example.com,port:5432},features:{darkMode:T}}"
```

### Mixed Structures

ZON intelligently combines formats:

```
metadata:"{version:1.0.5,env:production}"
users:@(5):id,name,active
1,Alice,T
2,Bob,F
...
logs:"[{id:101,level:INFO},{id:102,level:WARN}]"
```

---

## API Reference

### `Encode(data any) (string, error)`

Encodes Go data to ZON format.

```go
import zon "github.com/ZON-Format/zon-go"

data := map[string]any{
    "users": []any{
        map[string]any{"id": 1, "name": "Alice"},
        map[string]any{"id": 2, "name": "Bob"},
    },
}

encoded, err := zon.Encode(data)
```

**Returns:** ZON-formatted string

### `Decode(zonString string) (any, error)`

Decodes ZON format back to Go data. Default is strict mode.

```go
import zon "github.com/ZON-Format/zon-go"

data, err := zon.Decode(`
users:@(2):id,name
1,Alice
2,Bob
`)
```

**Returns:** Original Go data structure

### `DecodeWithOptions(zonString string, options *DecodeOptions) (any, error)`

Decodes ZON format with custom options.

```go
// Non-strict mode - allows row/field count mismatches
data, err := zon.DecodeWithOptions(zonString, &zon.DecodeOptions{Strict: false})
```

### Error Handling

```go
import zon "github.com/ZON-Format/zon-go"

decoded, err := zon.Decode(invalidZon)
if err != nil {
    if decodeErr, ok := err.(*zon.DecodeError); ok {
        fmt.Println(decodeErr.Code)    // "E001" or "E002"
        fmt.Println(decodeErr.Message) // Detailed error message
        fmt.Println(decodeErr.Context) // Context snippet
    }
}
```

**Error Codes:**

| Code | Description |
|------|-------------|
| E001 | Row count mismatch |
| E002 | Field count mismatch |
| E301 | Document size exceeds 100MB |
| E302 | Line length exceeds 1MB |
| E303 | Array length exceeds 1M items |
| E304 | Object key count exceeds 100K |

---

## Documentation

Comprehensive guides and references are available:

- **[Syntax Cheatsheet](./docs/syntax-cheatsheet.md)** - Quick reference for ZON format syntax
- **[API Reference](./docs/api-reference.md)** - Complete API documentation
- **[LLM Best Practices](./docs/llm-best-practices.md)** - Using ZON with LLMs
- **[Complete Specification](./SPEC.md)** - Formal ZON specification

---

## Links

- [GitHub Repository](https://github.com/ZON-Format/zon-go)
- [TypeScript Package](https://github.com/ZON-Format/zon-TS)
- [NPM Package (TypeScript)](https://www.npmjs.com/package/zon-format)

---

## License

Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)

MIT License - see [LICENSE](LICENSE) for details.
