package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func binName() string {
	if runtime.GOOS == "windows" {
		return "gox.exe"
	}
	return "gox"
}

func TestCLICompile(t *testing.T) {
	// Build the binary
	dir := t.TempDir()
	bin := filepath.Join(dir, binName())
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	// Run compile on testdata
	cmd = exec.Command(bin, "compile", "testdata/sumtype_basic.gox")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile failed: %s\n%s", err, out)
	}

	// Verify output file was created
	outputPath := "testdata/sumtype_basic_gen.go"
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("expected output file %s not found", outputPath)
	}
	defer os.Remove(outputPath)

	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "type OrderState interface") {
		t.Fatalf("output missing sum type interface:\n%s", content)
	}
}

func TestCLICheck(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, binName())
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	cmd = exec.Command(bin, "check", "testdata/sumtype_basic.gox")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check failed: %s\n%s", err, out)
	}
	if !strings.Contains(string(out), "ok") {
		t.Fatalf("expected 'ok' in output, got: %s", out)
	}
}
