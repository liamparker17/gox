package parser

import (
	"strings"

	"github.com/liamp/gox/ast"
	"github.com/liamp/gox/lexer"
)

func (p *Parser) parseFunc() *ast.Func {
	pos := p.tok.Pos
	var sigParts []string
	for p.tok.Kind != lexer.TokLBrace && p.tok.Kind != lexer.TokEOF {
		sigParts = append(sigParts, p.tok.Value)
		p.next()
	}
	p.expect(lexer.TokLBrace)
	fn := &ast.Func{Position: pos, Signature: strings.Join(sigParts, " ")}
	fn.Stmts = p.parseFuncBody()
	return fn
}

func (p *Parser) parseFuncBody() []ast.Stmt {
	var stmts []ast.Stmt
	depth := 1
	var goCode strings.Builder
	for p.tok.Kind != lexer.TokEOF {
		if p.tok.Kind == lexer.TokRBrace {
			depth--
			if depth == 0 {
				p.next()
				break
			}
			goCode.WriteString("}")
			p.next()
			continue
		}
		if p.tok.Kind == lexer.TokMatch && depth == 1 {
			if code := strings.TrimSpace(goCode.String()); code != "" {
				stmts = append(stmts, &ast.GoCode{Code: code})
				goCode.Reset()
			}
			m := p.parseMatchExpr()
			stmts = append(stmts, m)
			continue
		}
		if p.tok.Kind == lexer.TokLBrace {
			depth++
			goCode.WriteString("{")
			p.next()
			continue
		}
		goCode.WriteString(p.tok.Value)
		goCode.WriteString(" ")
		p.next()
	}
	if code := strings.TrimSpace(goCode.String()); code != "" {
		stmts = append(stmts, &ast.GoCode{Code: code})
	}
	return stmts
}
