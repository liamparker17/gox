package parser

import (
	"fmt"
	"strings"

	"github.com/liamparker17/gox/ast"
	"github.com/liamparker17/gox/lexer"
)

type Parser struct {
	lex           *lexer.Lexer
	tok           lexer.Token
	errors        []string
	lastDirective string
}

func Parse(filename, src string) (*ast.File, []string) {
	p := &Parser{lex: lexer.New(filename, src)}
	p.next()
	file := p.parseFile()
	return file, p.errors
}

func (p *Parser) next() {
	for {
		p.tok = p.lex.NextToken()
		if p.tok.Kind == lexer.TokNewline {
			continue
		}
		if p.tok.Kind == lexer.TokComment {
			if strings.Contains(p.tok.Value, "//gox:ignore exhaustive") {
				p.lastDirective = "ignore-exhaustive"
			}
			continue
		}
		break
	}
}

func (p *Parser) expect(kind lexer.TokenKind) lexer.Token {
	tok := p.tok
	if tok.Kind != kind {
		p.errorf("expected %s, got %s (%q)", kind, tok.Kind, tok.Value)
	}
	p.next()
	return tok
}

func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf("%s:%d:%d: %s", p.tok.Pos.File, p.tok.Pos.Line, p.tok.Pos.Column, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseFile() *ast.File {
	file := &ast.File{}
	if p.tok.Kind == lexer.TokPackage {
		p.next()
		file.Package = p.tok.Value
		p.next()
	}
	for p.tok.Kind == lexer.TokImport {
		file.Imports = append(file.Imports, p.parseImport())
	}
	for p.tok.Kind != lexer.TokEOF {
		switch p.tok.Kind {
		case lexer.TokSumtype:
			file.Decls = append(file.Decls, p.parseSumType())
		case lexer.TokContract:
			file.Decls = append(file.Decls, p.parseContract())
		case lexer.TokFunc:
			file.Decls = append(file.Decls, p.parseFunc())
		default:
			p.errorf("unexpected token %s (%q)", p.tok.Kind, p.tok.Value)
			p.next()
		}
	}
	return file
}

func (p *Parser) parseImport() ast.Import {
	p.next()
	imp := ast.Import{}
	if p.tok.Kind == lexer.TokString {
		imp.Path = p.tok.Value
		p.next()
	}
	return imp
}
