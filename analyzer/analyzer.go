package analyzer

import (
	"fmt"

	"github.com/liamparker17/gox/ast"
)

type CompileError struct {
	Pos     ast.Position
	Message string
	Kind    string
}

func (e CompileError) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Message)
}

type analyzer struct {
	sumTypes map[string]*ast.SumType
	errors   []CompileError
}

func Analyze(file *ast.File) []CompileError {
	return AnalyzeFiles([]*ast.File{file})
}

func AnalyzeFiles(files []*ast.File) []CompileError {
	a := &analyzer{sumTypes: make(map[string]*ast.SumType)}
	for _, f := range files {
		for _, decl := range f.Decls {
			if st, ok := decl.(*ast.SumType); ok {
				a.sumTypes[st.Name] = st
			}
		}
	}
	for _, f := range files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.Func:
				a.analyzeFunc(d)
			case *ast.Contract:
				a.analyzeContract(d)
			}
		}
	}
	return a.errors
}

func (a *analyzer) addError(pos ast.Position, kind, msg string) {
	a.errors = append(a.errors, CompileError{Pos: pos, Kind: kind, Message: msg})
}

func (a *analyzer) analyzeFunc(fn *ast.Func) {
	for _, stmt := range fn.Stmts {
		if m, ok := stmt.(*ast.Match); ok {
			a.resolveMatchType(m)
			a.checkExhaustiveness(m)
			a.checkBindings(m)
		}
	}
}

func (a *analyzer) analyzeContract(c *ast.Contract) {
	a.validateAnnotations(c)
	a.validateRoute(c)
}
