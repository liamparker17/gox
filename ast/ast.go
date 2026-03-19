package ast

// Node is the interface all AST nodes implement.
type Node interface {
	Pos() Position
}

// Position tracks source location for error reporting.
type Position struct {
	File   string
	Line   int
	Column int
}

func (p Position) Pos() Position { return p }

// Decl is the interface for top-level declarations.
type Decl interface {
	Node
	declNode()
}

// File represents a parsed .gox file.
type File struct {
	Package string
	Imports []Import
	Decls   []Decl
}

// Import represents a Go import statement.
type Import struct {
	Alias string
	Path  string
}

// SumType represents a tagged union declaration.
type SumType struct {
	Position
	Name     string
	Variants []Variant
}

func (*SumType) declNode() {}

// Variant is one arm of a sum type.
type Variant struct {
	Position
	Name   string
	Fields []Field
}

// Field is a named, typed field in a variant or contract.
type Field struct {
	Name string
	Type string
}

// Contract represents a full-stack operation declaration.
type Contract struct {
	Position
	Name   string
	Input  []AnnotatedField
	Output []Field
	Errors []ContractError
	Route  *Route
}

func (*Contract) declNode() {}

// Route is the HTTP binding for a contract.
type Route struct {
	Method string
	Path   string
}

// ContractError is a named error variant with HTTP status code.
type ContractError struct {
	Name       string
	StatusCode int
}

// AnnotatedField is a field with validation annotations.
type AnnotatedField struct {
	Field
	Annotations []Annotation
}

// Annotation is a validation directive like @required or @minlen(8).
type Annotation struct {
	Name string
	Args []string
}

// Func is a function block containing Go code and match expressions.
type Func struct {
	Position
	Signature string
	Stmts     []Stmt
}

func (*Func) declNode() {}

// Stmt is the interface for statements inside a Func body.
type Stmt interface {
	stmtNode()
}

// GoCode is raw Go code passed through verbatim.
type GoCode struct {
	Code string
}

func (*GoCode) stmtNode() {}

// Match is an exhaustive pattern match on a sum type.
type Match struct {
	Position
	Expr     string
	TypeName string
	Arms     []MatchArm
	Ignore   bool
}

func (*Match) stmtNode() {}

// MatchArm is one case in a match expression.
type MatchArm struct {
	Variant  string
	Bindings []string
	Body     string
}
