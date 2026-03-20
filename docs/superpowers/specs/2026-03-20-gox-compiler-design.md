# gox Compiler Design Spec

**Date:** 2026-03-20
**Status:** Draft
**Module path:** `github.com/liamp/gox`
**Target Go version:** 1.24+
**Project location:** `C:\Users\liamp\OneDrive\Desktop\Newlang`

---

## 1. Overview

gox is a compile-time augmentation layer for Go. It accepts `.gox` files containing extended syntax (sum types, contracts, exhaustive matches) and compiles them to valid, idiomatic Go code.

**Core philosophy:** Do not modify the Go runtime or language. Build a transpiler that adds safety, expressiveness, and developer experience while preserving Go's simplicity and performance.

**Architecture:** Sequential pipeline (each stage runs once) — Lexer → Parser → Typed AST → Analyzer → Codegen → `.go` files.

---

## 2. `.gox` Syntax

### 2.1 File Structure

```
package orders

import "time"

sumtype OrderState { ... }
contract CreateOrder { ... }
```

- `.gox` files declare a `package` like Go files
- `import` uses standard Go syntax
- Multiple constructs per file
- `.gox` files are **purely declarative** — they contain only `sumtype`, `contract`, and `func` blocks
- Regular Go code lives in `.go` files alongside `.gox` files in the same package
- `match` expressions appear inside `func` blocks in `.gox` files (not in raw `.go` files)

### 2.2 Sum Types

```
sumtype OrderState {
    Pending
    Paid { amount: float64, paidAt: time.Time }
    Shipped { trackingNumber: string }
    Cancelled { reason: string }
}
```

- `sumtype` keyword (distinct from Go's `type`)
- Variants can be bare (no data) or carry named fields
- Field types use Go types directly

### 2.3 Match Expressions

```
match state : OrderState {
    Pending => handlePending()
    Paid(amount, paidAt) => processPaid(amount, paidAt)
    Shipped(trackingNumber) => ship(trackingNumber)
    Cancelled(reason) => cancel(reason)
}
```

- `match` keyword signals exhaustiveness enforcement
- `match` is a **statement**, not an expression — it does not return a value
- `match` appears inside `func` blocks in `.gox` files
- `Variant(field1, field2)` destructures variant fields — binding count must match variant field count
- Missing a case is a compile error
- Opt-out with `//gox:ignore exhaustive` directive
- No `default` allowed unless using the ignore directive
- `Match.Expr` must be a simple identifier (variable name) — not a function call or field access

### 2.4 Contracts

```
contract CreateUser {
    input {
        email: string     @required @email
        password: string  @required @minlen(8)
        name: string      @optional
    }
    output {
        id: string
        createdAt: time.Time
    }
    errors {
        EmailTaken 409
        WeakPassword 400
    }
    route POST /api/users
}
```

- `input`/`output` blocks define request/response shapes
- `@` annotations drive validation codegen
- `errors` block defines typed error variants (generated as a sum type)
- Error variants can optionally specify an HTTP status code: `EmailTaken 409`, `WeakPassword 400`. Default is 400 if no code is given.
- `route` declares HTTP method and path — **optional**. If omitted, only structs + validation + error type are generated (no handler/client).

**Supported annotations:**

| Annotation | Generated check |
|---|---|
| `@required` | zero-value check: `== ""` strings, `== 0` numbers, `.IsZero()` for `time.Time`, `== nil` for pointers/slices/maps |
| `@email` | `strings.Contains(field, "@")` |
| `@minlen(N)` | `len(field) < N` |
| `@maxlen(N)` | `len(field) > N` |
| `@min(N)` | `field < N` |
| `@max(N)` | `field > N` |
| `@optional` | `omitempty` json tag, no required check |

---

## 3. AST Node Structure

```go
type Node interface{ Pos() Position }

type Position struct {
    File   string
    Line   int
    Column int
}

func (p Position) Pos() Position { return p }

// Decl is the interface for top-level declarations
type Decl interface {
    Node
    declNode()
}

// Top-level file
type File struct {
    Package string
    Imports []Import
    Decls   []Decl // SumType | Contract | Func
}

type Import struct {
    Alias string
    Path  string
}

// Sum type
type SumType struct {
    Position
    Name     string
    Variants []Variant
}

type Variant struct {
    Position
    Name   string
    Fields []Field
}

type Field struct {
    Name string
    Type string // raw Go type string
}

// Contract
type Contract struct {
    Position
    Name   string
    Input  []AnnotatedField
    Output []Field
    Errors []ContractError
    Route  *Route
}

type Route struct {
    Method string
    Path   string
}

type ContractError struct {
    Name       string
    StatusCode int // default 400 if unspecified
}

type AnnotatedField struct {
    Field
    Annotations []Annotation
}

type Annotation struct {
    Name string
    Args []string
}

// Func block — contains Go code with embedded match expressions
type Func struct {
    Position
    Signature string   // raw Go function signature, e.g. "func ProcessOrder(state OrderState)"
    Stmts     []Stmt   // mix of GoCode and Match
}

// Stmt is either raw Go code or a match expression
type Stmt interface{ stmtNode() }

type GoCode struct {
    Code string // raw Go code passed through verbatim
}
func (GoCode) stmtNode() {}

// Ensure top-level types implement Decl
func (*SumType) declNode()  {}
func (*Contract) declNode() {}
func (*Func) declNode()     {}

// Match expression (a Stmt inside Func)
type Match struct {
    Position
    Expr     string
    TypeName string
    Arms     []MatchArm
    Ignore   bool
}

func (*Match) stmtNode() {}

type MatchArm struct {
    Variant  string
    Bindings []string // must match variant field count (enforced by analyzer)
    Body     string   // raw Go code, terminated by next variant name or closing }
}
```

---

## 4. Compiler Pipeline

### 4.1 Lexer

**Input:** `.gox` file (raw text)
**Output:** Stream of tokens

Token kinds:
- Keywords: `sumtype`, `contract`, `match`, `func`, `input`, `output`, `errors`, `route`, `package`, `import`
- Structural: `{`, `}`, `(`, `)`, `:`, `=>`, `@`
- Literals: identifiers, strings, numbers
- Special: `TokGoBlock` (raw Go code — see termination rules below), `TokComment` (including `//gox:` directives)

**GoBlock termination rules:** A `TokGoBlock` captures raw Go code. The lexer tracks brace depth and is aware of string literals and comments to avoid false termination on `}` inside Go code. Termination depends on context:

1. **Inside `func` body:** GoCode terminates when the lexer encounters the `match` keyword at the current func brace depth, or when the closing `}` of the func body is reached (brace depth returns to pre-func level).
2. **Inside `match` arm:** Arm body terminates when the lexer encounters an identifier at the match brace depth that matches a known variant name (lookahead: next non-whitespace is `=>` or `(`), or when the closing `}` of the match block is reached.
3. **Func signature:** Everything from the `func` keyword up to (but not including) the opening `{` of the function body is captured as the raw signature string. Multi-line signatures with return types are supported: `func Foo(x int) (Bar, error)`.
- Meta: `TokEOF`

Responsibilities:
- Track `Position` for error reporting
- Recognize gox-specific keywords
- Capture annotations with arguments

### 4.2 Parser

**Input:** Token stream
**Output:** Typed AST (`ast.File`)

Recursive descent parser with entry points:
- `parseFile()` → dispatches to construct parsers based on keyword
- `parseSumType()` → `SumType` + `Variant` + `Field`
- `parseContract()` → `Contract` + `AnnotatedField` + `Route`
- `parseFunc()` → `Func` containing `[]Stmt` (mix of `GoCode` and `Match`)
- `parseMatch()` → `Match` + `MatchArm` (called from within `parseFunc`)

Validates basic syntax at parse time (balanced braces, field names, duplicate variant names).

**Func body parsing:** Inside a `func` block, the parser scans for the `match` keyword. Everything between match expressions is captured as `GoCode` nodes. This avoids needing a full Go parser — only `match` is structurally parsed.

### 4.3 Analyzer

**Input:** Typed AST
**Output:** Validated AST + compile errors

Passes:
1. **Registration** — first pass collects all sum type names into `sumTypes` map (enables forward references and cross-file resolution when compiling a directory)
2. **Type resolution** — verify sum type names referenced in `match` and contract `errors` are declared
3. **Exhaustiveness check** — for each `Match`, diff arm variants against `SumType.Variants`, error on missing (unless `Ignore` is set)
4. **Binding validation** — verify each `MatchArm`'s binding count matches its variant's field count
5. **Annotation validation** — reject unknown annotations, check argument counts (`@minlen` needs 1 arg, `@required` needs 0, etc.)
6. **Route validation** — valid HTTP method (`GET|POST|PUT|PATCH|DELETE`), path starts with `/`

```go
type Analyzer struct {
    sumTypes map[string]*ast.SumType
    errors   []CompileError
}

type CompileError struct {
    Pos     Position
    Message string
    Kind    string // "exhaustiveness", "type", "annotation", "route"
}
```

### 4.4 Codegen

**Input:** Validated AST
**Output:** `.go` files (formatted via `go/format`)

Generation rules per node type — see Section 5.

**Output file naming:** `{source}_gen.go` — e.g., `orders.gox` produces `orders_gen.go`. One output file per source file. The `_gen.go` suffix follows Go convention and signals the file is auto-generated.

**Import resolution:** Codegen tracks required imports per output file. User imports from `.gox` are forwarded. Codegen adds stdlib imports as needed: `fmt` (validation errors), `strings` (email check), `encoding/json` + `net/http` + `bytes` (contract handlers/clients). Duplicates are deduplicated. All validation logic is inlined — no runtime package dependency in MVP.

### 4.5 Future Passes (not in MVP)

- AST-to-AST transformations
- Desugar contracts into sum types
- Pattern journaling and macro expansion

---

## 5. Codegen Rules

### 5.1 Sum Type → Interface + Structs

```go
// Interface with sealed marker
type OrderState interface { isOrderState() }

// Per-variant struct
type OrderStatePending struct{}
func (OrderStatePending) isOrderState() {}

type OrderStatePaid struct {
    Amount float64
    PaidAt time.Time
}
func (OrderStatePaid) isOrderState() {}
```

**Naming:** variant struct = `{TypeName}{VariantName}`. Fields are exported (capitalized).

### 5.2 Match → Type Switch + Destructuring

```go
switch v := state.(type) {
case OrderStatePending:
    handlePending()
case OrderStatePaid:
    processPaid(v.Amount, v.PaidAt)
}
```

Bindings map positionally to struct fields: `Paid(amount, paidAt)` → `v.Amount`, `v.PaidAt`.

### 5.3 Contract → Structs + Validate + Handler + Client

**Input/Output structs:**
```go
type CreateUserInput struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name,omitempty"`
}

type CreateUserOutput struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"createdAt"`
}
```

**Error sum type:**
```go
type CreateUserError interface { isCreateUserError() }
type CreateUserErrorEmailTaken struct{}
func (CreateUserErrorEmailTaken) isCreateUserError() {}
```

**Validation function (collects all errors):**
```go
func ValidateCreateUserInput(in CreateUserInput) error {
    var errs []string
    if in.Email == "" { errs = append(errs, "email: required") }
    if in.Email != "" && !strings.Contains(in.Email, "@") { errs = append(errs, "email: invalid email") }
    if in.Password == "" { errs = append(errs, "password: required") }
    if in.Password != "" && len(in.Password) < 8 { errs = append(errs, "password: min length 8") }
    if len(errs) > 0 { return fmt.Errorf("%s", strings.Join(errs, "; ")) }
    return nil
}
```

**HTTP handler:**
```go
// Handler signature uses the contract's error sum type for typed error handling.
// The handler maps contract errors to HTTP status codes via a type switch.
// If route is omitted from the contract, handler and client are NOT generated —
// only structs, validation, and error sum type are emitted.
// All responses are JSON for consistency. Errors use {"error": "message"} format.
func CreateUserHandler(fn func(CreateUserInput) (CreateUserOutput, CreateUserError)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        var in CreateUserInput
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            w.WriteHeader(400)
            json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
            return
        }
        if err := ValidateCreateUserInput(in); err != nil {
            w.WriteHeader(422)
            json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
            return
        }
        out, cerr := fn(in)
        if cerr != nil {
            switch cerr.(type) {
            case CreateUserErrorEmailTaken:
                w.WriteHeader(409)
                json.NewEncoder(w).Encode(map[string]string{"error": "email already taken"})
            case CreateUserErrorWeakPassword:
                w.WriteHeader(400)
                json.NewEncoder(w).Encode(map[string]string{"error": "password too weak"})
            }
            return
        }
        json.NewEncoder(w).Encode(out)
    }
}
```

**Typed client (returns contract errors separately from transport errors):**
```go
// Returns (output, contractError, transportError).
// contractError is non-nil when the server returns a known error status.
// transportError is non-nil for network/decode failures.
func CreateUserClient(baseURL string, in CreateUserInput) (CreateUserOutput, CreateUserError, error) {
    body, err := json.Marshal(in)
    if err != nil { return CreateUserOutput{}, nil, fmt.Errorf("marshal: %w", err) }
    resp, err := http.Post(baseURL+"/api/users", "application/json", bytes.NewReader(body))
    if err != nil { return CreateUserOutput{}, nil, fmt.Errorf("request: %w", err) }
    defer resp.Body.Close()
    switch resp.StatusCode {
    case 200:
        var out CreateUserOutput
        if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
            return CreateUserOutput{}, nil, fmt.Errorf("decode: %w", err)
        }
        return out, nil, nil
    case 409:
        return CreateUserOutput{}, CreateUserErrorEmailTaken{}, nil
    case 400:
        return CreateUserOutput{}, CreateUserErrorWeakPassword{}, nil
    default:
        return CreateUserOutput{}, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
}
```

---

## 6. Project Structure

```
Newlang/
├── go.mod                    # module github.com/liamp/gox
├── go.sum
├── main.go                   # CLI entry point
├── cmd/
│   └── gox/
│       └── root.go           # CLI commands (compile, check, init)
├── lexer/
│   ├── lexer.go              # Lexer struct, NextToken()
│   ├── token.go              # TokenKind, Token struct
│   └── lexer_test.go
├── parser/
│   ├── parser.go             # Parser struct, ParseFile()
│   ├── sumtype.go            # parseSumType()
│   ├── contract.go           # parseContract()
│   ├── func.go               # parseFunc() + func body scanning
│   ├── match.go              # parseMatch()
│   └── parser_test.go
├── ast/
│   └── ast.go                # All AST node types
├── analyzer/
│   ├── analyzer.go           # Analyze(File) []CompileError
│   ├── exhaustiveness.go     # checkExhaustiveness()
│   ├── annotations.go        # validateAnnotations()
│   ├── types.go              # resolveTypes()
│   └── analyzer_test.go
├── codegen/
│   ├── codegen.go            # Generator struct, Generate(File) []OutputFile
│   ├── sumtype.go            # emitSumType()
│   ├── contract.go           # emitContract()
│   ├── match.go              # emitMatch()
│   ├── validate.go           # annotation → validation mapping
│   └── codegen_test.go
├── testdata/
│   ├── orders.gox            # Example sum type + match
│   ├── users.gox             # Example contract
│   ├── orders_expected.go    # Golden file test output
│   └── users_expected.go
└── docs/
    └── superpowers/
        └── specs/
            └── 2026-03-20-gox-compiler-design.md
```

Each compiler stage is its own Go package with no circular dependencies. `ast/` depends on nothing. All other packages import `ast/`.

---

## 7. CLI Interface

```
gox compile <file.gox>          # Compile .gox to .go
gox compile <dir>               # Compile all .gox files in directory
gox check <file.gox>            # Analyze only (no codegen), report errors
gox init                        # Create a sample .gox project
```

Flags:
- `-o <dir>` — output directory (default: same directory as source)
- `-v` — verbose output (show pipeline stages)
- `--dry-run` — show what would be generated without writing files

---

## 8. Exhaustiveness Enforcement

**Default behavior:** Missing a variant in a `match` is a compile error.

```
orders.gox:14:1: exhaustiveness error: match on OrderState missing variant: Cancelled
```

**Escape hatch:** `//gox:ignore exhaustive` on the line before the match.

```
//gox:ignore exhaustive
match state : OrderState {
    Pending => handlePending()
    Paid(amount, paidAt) => processPaid(amount, paidAt)
}
```

When ignored, a `default: panic("unhandled variant")` is inserted in the generated Go switch.

---

## 9. Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Parser | Custom recursive descent | `.gox` syntax isn't valid Go; need full control |
| Go target | 1.24+ | Bleeding edge, generics mature |
| Exhaustiveness | Strict + opt-out | Safety by default, pragmatic exceptions |
| Contract scope | Full stack (MVP) | Structs, validation, handlers, clients from day one |
| Module path | `github.com/liamp/gox` | Standard GitHub convention |
| Sum type codegen | Interface + structs | Idiomatic Go pattern for tagged unions |
| Match codegen | Type switch | Direct mapping, zero overhead |
| Output format | `go/format` | Ensures idiomatic formatting |
| Match semantics | Statement only | Simpler codegen, matches Go's switch semantics |
| Match expr | Identifier only | Avoids expression parsing complexity in MVP |
| File scope | Declarative (sumtype, contract, func) | No raw Go passthrough — keeps parser simple |
| Output naming | `{name}_gen.go` | Follows Go convention, signals auto-generated |
| Runtime dependency | None (all inlined) | Pure compile-time tool, no import tax |
| Contract route | Optional | Omit route → no handler/client generated |
| Cross-file resolution | Yes (directory compile) | Sum types visible across `.gox` files in same package |

---

## 10. Phasing

### Phase 1 (MVP)
- CLI tool (`gox compile`, `gox check`)
- Lexer + parser
- Sum types + exhaustiveness checking
- Contracts (full stack: structs, validation, handlers, clients)
- Golden file tests

### Phase 2
- Structured concurrency layer (`task.Group`)
- Code generation plugin system

### Phase 3
- Pattern journaling
- AI-assisted suggestions

---

## 11. Technical Constraints

- Output must pass `go build`
- Generated code must be readable and idiomatic
- Avoid runtime reflection where possible
- Prefer compile-time guarantees over runtime checks
- Generated files include `// Code generated by gox from {source}.gox. DO NOT EDIT.` header (includes source file for traceability)
