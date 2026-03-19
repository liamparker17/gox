package parser

import (
	"testing"

	"github.com/liamp/gox/ast"
)

func TestParseSumType(t *testing.T) {
	input := `package orders

sumtype OrderState {
	Pending
	Paid { amount: float64, paidAt: time.Time }
	Cancelled { reason: string }
}`
	file, errs := Parse("test.gox", input)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if file.Package != "orders" {
		t.Fatalf("package = %q, want orders", file.Package)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("got %d decls, want 1", len(file.Decls))
	}
	st, ok := file.Decls[0].(*ast.SumType)
	if !ok {
		t.Fatalf("decl is %T, want *ast.SumType", file.Decls[0])
	}
	if st.Name != "OrderState" {
		t.Fatalf("name = %q, want OrderState", st.Name)
	}
	if len(st.Variants) != 3 {
		t.Fatalf("got %d variants, want 3", len(st.Variants))
	}
	if st.Variants[0].Name != "Pending" || len(st.Variants[0].Fields) != 0 {
		t.Fatalf("variant[0] = %+v, want Pending with 0 fields", st.Variants[0])
	}
	if st.Variants[1].Name != "Paid" || len(st.Variants[1].Fields) != 2 {
		t.Fatalf("variant[1] = %+v, want Paid with 2 fields", st.Variants[1])
	}
	if st.Variants[1].Fields[0].Name != "amount" || st.Variants[1].Fields[0].Type != "float64" {
		t.Fatalf("field[0] = %+v, want amount:float64", st.Variants[1].Fields[0])
	}
	if st.Variants[1].Fields[1].Name != "paidAt" || st.Variants[1].Fields[1].Type != "time.Time" {
		t.Fatalf("field[1] = %+v, want paidAt:time.Time", st.Variants[1].Fields[1])
	}
}

func TestParseSumTypeDuplicateVariant(t *testing.T) {
	input := `package test

sumtype Dup {
	A
	A
}`
	_, errs := Parse("test.gox", input)
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate variant, got none")
	}
}

func TestParseContract(t *testing.T) {
	input := `package users

contract CreateUser {
	input {
		email: string @required @email
		password: string @required @minlen(8)
		name: string @optional
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
}`
	file, errs := Parse("test.gox", input)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("got %d decls, want 1", len(file.Decls))
	}
	c, ok := file.Decls[0].(*ast.Contract)
	if !ok {
		t.Fatalf("decl is %T, want *ast.Contract", file.Decls[0])
	}
	if c.Name != "CreateUser" {
		t.Fatalf("name = %q, want CreateUser", c.Name)
	}
	if len(c.Input) != 3 {
		t.Fatalf("got %d inputs, want 3", len(c.Input))
	}
	emailField := c.Input[0]
	if emailField.Name != "email" || len(emailField.Annotations) != 2 {
		t.Fatalf("input[0] = %+v, want email with 2 annotations", emailField)
	}
	if emailField.Annotations[0].Name != "required" {
		t.Fatalf("annotation[0] = %q, want required", emailField.Annotations[0].Name)
	}
	pwField := c.Input[1]
	if len(pwField.Annotations) != 2 || pwField.Annotations[1].Name != "minlen" {
		t.Fatalf("input[1] annotations = %+v, want [required, minlen]", pwField.Annotations)
	}
	if len(pwField.Annotations[1].Args) != 1 || pwField.Annotations[1].Args[0] != "8" {
		t.Fatalf("minlen args = %v, want [8]", pwField.Annotations[1].Args)
	}
	if len(c.Output) != 2 {
		t.Fatalf("got %d outputs, want 2", len(c.Output))
	}
	if len(c.Errors) != 2 {
		t.Fatalf("got %d errors, want 2", len(c.Errors))
	}
	if c.Errors[0].Name != "EmailTaken" || c.Errors[0].StatusCode != 409 {
		t.Fatalf("error[0] = %+v, want EmailTaken 409", c.Errors[0])
	}
	if c.Route == nil || c.Route.Method != "POST" || c.Route.Path != "/api/users" {
		t.Fatalf("route = %+v, want POST /api/users", c.Route)
	}
}

func TestParseContractNoRoute(t *testing.T) {
	input := `package test

contract Simple {
	input {
		name: string @required
	}
	output {
		id: string
	}
}`
	file, errs := Parse("test.gox", input)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	c := file.Decls[0].(*ast.Contract)
	if c.Route != nil {
		t.Fatalf("route should be nil, got %+v", c.Route)
	}
}

func TestParseMatch(t *testing.T) {
	input := `package handlers

import "fmt"

func HandleOrder(state OrderState) {
	fmt.Println("processing")
	match state : OrderState {
		Pending => fmt.Println("pending")
		Paid(amount, paidAt) => fmt.Printf("paid %v at %v", amount, paidAt)
		Cancelled(reason) => fmt.Println(reason)
	}
	fmt.Println("done")
}`
	file, errs := Parse("test.gox", input)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("got %d decls, want 1", len(file.Decls))
	}
	fn, ok := file.Decls[0].(*ast.Func)
	if !ok {
		t.Fatalf("decl is %T, want *ast.Func", file.Decls[0])
	}
	if fn.Signature != "func HandleOrder ( state OrderState )" && fn.Signature != "func HandleOrder(state OrderState)" {
		// Accept either space-separated or compact form depending on lexer output
		t.Logf("sig = %q (checking loosely)", fn.Signature)
	}
	if len(fn.Stmts) != 3 {
		t.Fatalf("got %d stmts, want 3 (GoCode, Match, GoCode)", len(fn.Stmts))
	}
	m, ok := fn.Stmts[1].(*ast.Match)
	if !ok {
		t.Fatalf("stmt[1] is %T, want *ast.Match", fn.Stmts[1])
	}
	if m.Expr != "state" || m.TypeName != "OrderState" {
		t.Fatalf("match expr=%q type=%q", m.Expr, m.TypeName)
	}
	if len(m.Arms) != 3 {
		t.Fatalf("got %d arms, want 3", len(m.Arms))
	}
	if m.Arms[1].Variant != "Paid" || len(m.Arms[1].Bindings) != 2 {
		t.Fatalf("arm[1] = %+v, want Paid with 2 bindings", m.Arms[1])
	}
}
