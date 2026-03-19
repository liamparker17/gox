package codegen

import (
	"os"
	"strings"
	"testing"

	"github.com/liamp/gox/parser"
)

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
