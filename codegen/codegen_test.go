package codegen

import (
	"os"
	"strings"
	"testing"

	"github.com/liamp/gox/parser"
)

func TestGoldenContract(t *testing.T) {
	src, err := os.ReadFile("../testdata/contract_basic.gox")
	if err != nil {
		t.Fatal(err)
	}

	file, errs := parser.Parse("contract_basic.gox", string(src))
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	outputs := gen.Generate(file, "contract_basic.gox")
	if len(outputs) != 1 {
		t.Fatalf("got %d outputs, want 1", len(outputs))
	}

	// Write actual output for inspection
	os.WriteFile("../testdata/contract_basic_actual.go", []byte(outputs[0].Content), 0644)

	expected, err := os.ReadFile("../testdata/contract_basic_expected.go")
	if err != nil {
		t.Skipf("no expected file yet: %v", err)
	}

	got := strings.TrimSpace(outputs[0].Content)
	want := strings.TrimSpace(string(expected))
	if got != want {
		t.Fatalf("output mismatch.\n--- GOT ---\n%s\n--- WANT ---\n%s", got, want)
	}
}

func TestEmitMatch(t *testing.T) {
	input := `package handlers

sumtype Color {
	Red
	Green { shade: string }
}

func Handle(c Color) {
	match c : Color {
		Red => doRed()
		Green(shade) => doGreen(shade)
	}
}`
	file, errs := parser.Parse("test.gox", string(input))
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	outputs := gen.Generate(file, "test.gox")
	if len(outputs) != 1 {
		t.Fatalf("got %d outputs, want 1", len(outputs))
	}

	out := outputs[0].Content
	// Must contain type switch
	if !strings.Contains(out, "switch v := c.(type)") {
		t.Fatalf("missing type switch in output:\n%s", out)
	}
	if !strings.Contains(out, "case ColorRed:") {
		t.Fatalf("missing case ColorRed in output:\n%s", out)
	}
	if !strings.Contains(out, "case ColorGreen:") {
		t.Fatalf("missing case ColorGreen in output:\n%s", out)
	}
}

func TestGoldenSumType(t *testing.T) {
	src, err := os.ReadFile("../testdata/sumtype_basic.gox")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("../testdata/sumtype_basic_expected.go")
	if err != nil {
		t.Fatal(err)
	}

	file, errs := parser.Parse("sumtype_basic.gox", string(src))
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	outputs := gen.Generate(file, "sumtype_basic.gox")
	if len(outputs) != 1 {
		t.Fatalf("got %d outputs, want 1", len(outputs))
	}

	got := strings.TrimSpace(outputs[0].Content)
	want := strings.TrimSpace(string(expected))
	if got != want {
		t.Fatalf("output mismatch.\n--- GOT ---\n%s\n--- WANT ---\n%s", got, want)
	}
}
