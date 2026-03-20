package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liamparker17/gox/analyzer"
	"github.com/liamparker17/gox/ast"
	"github.com/liamparker17/gox/codegen"
	"github.com/liamparker17/gox/parser"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: gox <compile|check> <file.gox|dir>")
		os.Exit(1)
	}

	command := os.Args[1]
	target := os.Args[2]

	files, err := resolveFiles(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "compile":
		if !compile(files) {
			os.Exit(1)
		}
	case "check":
		if !check(files) {
			os.Exit(1)
		}
		fmt.Println("ok")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(1)
	}
}

func resolveFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{target}, nil
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".gox") {
			files = append(files, filepath.Join(target, e.Name()))
		}
	}
	return files, nil
}

// parseAll parses all .gox files and returns parsed ASTs keyed by path.
func parseAll(paths []string) (map[string]*ast.File, bool) {
	parsed := make(map[string]*ast.File)
	ok := true
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			ok = false
			continue
		}
		filename := filepath.Base(path)
		file, parseErrs := parser.Parse(filename, string(src))
		if len(parseErrs) > 0 {
			for _, e := range parseErrs {
				fmt.Fprintln(os.Stderr, e)
			}
			ok = false
			continue
		}
		parsed[path] = file
	}
	return parsed, ok
}

func compile(paths []string) bool {
	parsed, ok := parseAll(paths)
	if !ok {
		return false
	}

	// Collect all files for cross-file analysis
	allFiles := make([]*ast.File, 0, len(parsed))
	for _, f := range parsed {
		allFiles = append(allFiles, f)
	}

	analyzeErrs := analyzer.AnalyzeFiles(allFiles)
	if len(analyzeErrs) > 0 {
		for _, e := range analyzeErrs {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Kind, e.Message)
		}
		return false
	}

	// Generate output for each file
	gen := codegen.New()
	for path, file := range parsed {
		filename := filepath.Base(path)
		outputs := gen.Generate(file, filename)
		for _, out := range outputs {
			outputPath := filepath.Join(filepath.Dir(path), out.Filename)
			if err := os.WriteFile(outputPath, []byte(out.Content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputPath, err)
				return false
			}
			fmt.Printf("wrote %s\n", outputPath)
		}
	}
	return true
}

func check(paths []string) bool {
	parsed, ok := parseAll(paths)
	if !ok {
		return false
	}

	allFiles := make([]*ast.File, 0, len(parsed))
	for _, f := range parsed {
		allFiles = append(allFiles, f)
	}

	analyzeErrs := analyzer.AnalyzeFiles(allFiles)
	if len(analyzeErrs) > 0 {
		for _, e := range analyzeErrs {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Kind, e.Message)
		}
		return false
	}
	return true
}
