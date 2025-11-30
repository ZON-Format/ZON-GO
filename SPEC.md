# ZON Specification

## Zero Overhead Notation - Formal Specification

**Version:** 1.0.5

**Date:** 2025-11-28

**Status:** Stable Release

**Authors:** ZON Format Contributors

**License:** MIT

---

## Abstract

Zero Overhead Notation (ZON) is a compact, line-oriented text format that encodes the JSON data model with minimal redundancy optimized for large language model token efficiency. ZON achieves up to 23.8% token reduction compared to JSON through single-character primitives (`T`, `F`), null as `null`, explicit table markers (`@`), colon-less nested structures, and intelligent quoting rules. Arrays of uniform objects use tabular encoding with column headers declared once; metadata uses flat key-value pairs. This specification defines ZON's concrete syntax, canonical value formatting, encoding/decoding behavior, conformance requirements, and strict validation rules. ZON provides deterministic, lossless representation achieving 100% LLM retrieval accuracy in benchmarks.

## Status of This Document

This document is a **Stable Release v1.0.5** and defines normative behavior for ZON encoders, decoders, and validators. Implementation feedback should be reported at https://github.com/ZON-Format/zon-go.

Backward compatibility is maintained across v1.0.x releases. Major versions (v2.x) may introduce breaking changes.

## Normative References

**[RFC2119]** Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", BCP 14, RFC 2119, March 1997.  
https://www.rfc-editor.org/rfc/rfc2119

**[RFC8174]** Leiba, B., "Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words", BCP 14, RFC 8174, May 2017.  
https://www.rfc-editor.org/rfc/rfc8174

**[RFC8259]** Bray, T., "The JavaScript Object Notation (JSON) Data Interchange Format", STD 90, RFC 8259, December 2017.  
https://www.rfc-editor.org/rfc/rfc8259

## Informative References

**[RFC4180]** Shafranovich, Y., "Common Format and MIME Type for Comma-Separated Values (CSV) Files", RFC 4180, October 2005.  
https://www.rfc-editor.org/rfc/rfc4180

**[ISO8601]** ISO 8601:2019, "Date and time — Representations for information interchange".

**[UNICODE]** The Unicode Consortium, "The Unicode Standard", Version 15.1, September 2023.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Terminology and Conventions](#1-terminology-and-conventions)
3. [Data Model](#2-data-model)
4. [Encoding Normalization](#3-encoding-normalization)
5. [Decoding Interpretation](#4-decoding-interpretation)
6. [Concrete Syntax](#5-concrete-syntax)
7. [Primitives](#6-primitives)
8. [Strings and Keys](#7-strings-and-keys)
9. [Objects](#8-objects)
10. [Arrays](#9-arrays)
11. [Table Format](#10-table-format)
12. [Quoting and Escaping](#11-quoting-and-escaping)
13. [Whitespace](#12-whitespace-and-line-endings)
14. [Conformance](#13-conformance-and-options)
15. [Strict Mode Errors](#14-strict-mode-errors)
16. [Security](#15-security-considerations)
17. [Internationalization](#16-internationalization)
18. [Interoperability](#17-interoperability)
19. [Media Type](#18-media-type)
20. [Error Handling](#19-error-handling)
21. [Appendices](#appendices)

---

## Introduction (Informative)

### Purpose

ZON addresses token bloat in JSON while maintaining structural fidelity. By declaring column headers once, using single-character tokens, and eliminating redundant punctuation, ZON achieves optimal compression for LLM contexts.

### Design Goals

1. **Minimize tokens** - Every character counts in LLM context windows
2. **Preserve structure** - 100% lossless round-trip conversion
3. **Human readable** - Debuggable, understandable format
4. **LLM friendly** - Explicit markers aid comprehension
5. **Deterministic** - Same input → same output
6. **Deep Nesting** - Efficiently handles complex, recursive structures

### Use Cases

✅ **Use ZON for:**
- LLM prompt contexts (RAG, few-shot examples)
- Log storage and analysis
- Configuration files
- Browser storage (localStorage)
- Tabular data interchange
- **Complex nested data structures** (ZON excels here)

❌ **Don't use ZON for:**
- Public REST APIs (use JSON for compatibility)
- Real-time streaming protocols (not yet supported)
- Files requiring comments (use YAML/JSONC)

### Example

**JSON (118 chars):**
```json
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}
```

**ZON (64 chars, 46% reduction):**
```zon
users:@(2):active,id,name
T,1,Alice
F,2,Bob
```

---

## 1. Terminology and Conventions

### 1.1 RFC2119 Keywords

The keywords **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** are interpreted per [RFC2119] and [RFC8174].

### 1.2 Definitions

**ZON document** - UTF-8 text conforming to this specification

**Line** - Character sequence terminated by LF (`\n`)

**Key-value pair** - Line pattern: `key:value`

**Table** - Array of uniform objects with header + data rows

**Table header** - Pattern: `key:@(N):columns` or `@(N):columns`

**Meta separator** - Colon (`:`) separating keys/values

**Table marker** - At-sign (`@`) indicating table structure

**Primitive** - Boolean, null, number, or string (not object/array)

**Uniform array** - All elements are objects with identical keys

**Strict mode** - Validation enforcing row/column counts

---

## 2. Data Model

### 2.1 JSON Compatibility

ZON encodes the JSON data model:
- **Primitives**: `string | number | boolean | null`
- **Objects**: `{ [string]: JsonValue }`
- **Arrays**: `JsonValue[]`

### 2.2 Ordering

- **Arrays**: Order MUST be preserved exactly
- **Objects**: Key order MUST be preserved
  - Encoders SHOULD sort keys alphabetically
  - Decoders MUST preserve document order

### 2.3 Canonical Numbers

**Requirements for ENCODER:**

1. **No leading zeros:** `007` → invalid
2. **No trailing zeros:** `3.14000` → `3.14`
3. **No unnecessary decimals:** Integer `5` stays `5`, not `5.0`
4. **No scientific notation:** `1e6` → `1000000`, `1e-3` → `0.001`
5. **Special values map to null:**
   - `NaN` → `null`
   - `Infinity` → `null`
   - `-Infinity` → `null`

---

## 6. Primitives

### 6.1 Booleans

**Encoding:**
- `true` → `T`
- `false` → `F`

**Decoding:**
- `T` (case-sensitive) → `true`
- `F` (case-sensitive) → `false`

**Rationale:** 75% character reduction

### 6.2 Null

**Encoding:**
- `null` → `null` (4-character literal)

**Decoding:**
- `null` → `null`
- Also accepts (case-insensitive): `none`, `nil`

### 6.3 Numbers

**Examples:**
```zon
age:30
price:19.99
score:-42
temp:98.6
large:1000000
```

---

## 7. Strings and Keys

### 7.1 Safe Strings (Unquoted)

Pattern: `^[a-zA-Z0-9_\-\.]+$`

**Examples:**
```zon
name:Alice
user_id:u123
version:v1.0.4
api-key:sk_test_key
```

### 7.2 Required Quoting

Quote strings if they:

1. **Contain structural chars:** `,`, `:`, `[`, `]`, `{`, `}`, `"`
2. **Match literal keywords:** `T`, `F`, `true`, `false`, `null`, `none`, `nil`
3. **Look like PURE numbers:** `123`, `3.14`, `1e6`
4. **Have whitespace:** Leading/trailing spaces
5. **Are empty:** `""` (MUST quote)
6. **Contain escapes:** Newlines, tabs, quotes

---

## 10. Table Format

### 10.1 Header Syntax

**With key:**
```
users:@(2):active,id,name
```

**Root array:**
```
@(2):active,id,name
```

**Components:**
- `users` - Array key (optional for root)
- `@` - Table marker (REQUIRED)
- `(2)` - Row count (REQUIRED for strict mode)
- `:` - Separator (REQUIRED)
- `active,id,name` - Columns, comma-separated (REQUIRED)

### 10.2 Column Order

Columns SHOULD be sorted alphabetically.

### 10.3 Data Rows

Each row is comma-separated values, one row per line.

---

## 13. Conformance and Options

### 13.1 Encoder Checklist

✅ **A conforming encoder MUST:**

- [ ] Emit UTF-8 with LF line endings
- [ ] Encode booleans as `T`/`F`
- [ ] Encode null as `null`
- [ ] Emit canonical numbers
- [ ] Normalize NaN/Infinity to `null`
- [ ] Detect uniform arrays → table format
- [ ] Emit table headers: `key:@(N):columns`
- [ ] Sort columns alphabetically
- [ ] Sort object keys alphabetically
- [ ] Quote strings per §7.2-7.3
- [ ] Ensure round-trip: `decode(encode(x)) === x`

### 13.2 Decoder Checklist

✅ **A conforming decoder MUST:**

- [ ] Accept UTF-8 (LF or CRLF)
- [ ] Decode `T` → true, `F` → false, `null` → null
- [ ] Parse decimal and exponent numbers
- [ ] Parse table headers: `key:@(N):columns`
- [ ] Preserve array order
- [ ] Preserve key order
- [ ] Enforce row count (strict mode)
- [ ] Enforce field count (strict mode)

### 13.3 Strict Mode

**Enabled by default** in reference implementation.

Enforces:
- Table row count = declared `(N)`
- Each row field count = column count
- No malformed headers
- No unterminated strings

---

## 14. Strict Mode Errors

### Error Codes

| Code | Description |
|------|-------------|
| E001 | Row count mismatch |
| E002 | Field count mismatch |
| E301 | Document size exceeds 100MB |
| E302 | Line length exceeds 1MB |
| E303 | Array length exceeds 1M items |
| E304 | Object key count exceeds 100K |

---

## 15. Security Considerations

### 15.1 Resource Limits

Implementations SHOULD limit:
- Document size: 100 MB
- Line length: 1 MB
- Nesting depth: 100 levels
- Array length: 1,000,000
- Object keys: 100,000

Prevents denial-of-service attacks.

---

## 18. Media Type & File Extension

### 18.1 File Extension

**Extension:** `.zonf`

### 18.2 Media Type

**Media type:** `text/zon`

**Charset:** UTF-8 (always)

---

## Appendices

### Appendix A: Examples

**A.1 Simple Object**
```zon
active:T
age:30
name:Alice
```

**A.2 Table**
```zon
users:@(2):active,id,name
T,1,Alice
F,2,Bob
```

**A.3 Mixed**
```zon
tags:"[api,auth]"
version:1.0
users:@(1):id,name
1,Alice
```

### Appendix B: Test Suite

**Coverage:**
- ✅ 94/94 unit tests
- ✅ All roundtrip tests pass
- ✅ 100% data integrity

### Appendix C: License

MIT License

Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---

**End of Specification**
