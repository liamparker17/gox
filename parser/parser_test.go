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
