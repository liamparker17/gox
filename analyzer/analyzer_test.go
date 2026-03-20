package analyzer

import (
	"testing"

	"github.com/liamparker17/gox/parser"
)

func TestExhaustivenessPass(t *testing.T) {
	input := `package test

sumtype Color {
	Red
	Green
	Blue
}

func Handle(c Color) {
	match c : Color {
		Red => doRed()
		Green => doGreen()
		Blue => doBlue()
	}
}`
	file, perrs := parser.Parse("test.gox", input)
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	errs := Analyze(file)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestExhaustivenessFail(t *testing.T) {
	input := `package test

sumtype Color {
	Red
	Green
	Blue
}

func Handle(c Color) {
	match c : Color {
		Red => doRed()
		Green => doGreen()
	}
}`
	file, perrs := parser.Parse("test.gox", input)
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	errs := Analyze(file)
	if len(errs) == 0 {
		t.Fatal("expected exhaustiveness error, got none")
	}
	found := false
	for _, e := range errs {
		if e.Kind == "exhaustiveness" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected exhaustiveness error, got: %v", errs)
	}
}

func TestUnknownSumType(t *testing.T) {
	input := `package test

func Handle(c Unknown) {
	match c : Unknown {
		A => doA()
	}
}`
	file, perrs := parser.Parse("test.gox", input)
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	errs := Analyze(file)
	if len(errs) == 0 {
		t.Fatal("expected type error, got none")
	}
}

func TestBindingCountMismatch(t *testing.T) {
	input := `package test

sumtype Msg {
	Hello { name: string, age: int }
}

func Handle(m Msg) {
	match m : Msg {
		Hello(name) => use(name)
	}
}`
	file, perrs := parser.Parse("test.gox", input)
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	errs := Analyze(file)
	if len(errs) == 0 {
		t.Fatal("expected binding count error, got none")
	}
}

func TestAnnotationValidation(t *testing.T) {
	input := `package test

contract Bad {
	input {
		email: string @unknown
	}
	output {
		id: string
	}
}`
	file, perrs := parser.Parse("test.gox", input)
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	errs := Analyze(file)
	if len(errs) == 0 {
		t.Fatal("expected annotation error, got none")
	}
}
