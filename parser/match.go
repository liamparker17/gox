package parser

import (
	"strings"

	"github.com/liamparker17/gox/ast"
	"github.com/liamparker17/gox/lexer"
)

func (p *Parser) parseMatchExpr() *ast.Match {
	pos := p.tok.Pos
	p.next() // skip "match"
	expr := p.expect(lexer.TokIdent)
	p.expect(lexer.TokColon)
	typeName := p.expect(lexer.TokIdent)
	p.expect(lexer.TokLBrace)
	m := &ast.Match{
		Position: pos,
		Expr:     expr.Value,
		TypeName: typeName.Value,
		Ignore:   p.lastDirective == "ignore-exhaustive",
	}
	p.lastDirective = ""
	for p.tok.Kind != lexer.TokRBrace && p.tok.Kind != lexer.TokEOF {
		arm := p.parseMatchArm()
		m.Arms = append(m.Arms, arm)
	}
	p.expect(lexer.TokRBrace)
	return m
}

func (p *Parser) parseMatchArm() ast.MatchArm {
	variant := p.expect(lexer.TokIdent)
	arm := ast.MatchArm{Variant: variant.Value}
	if p.tok.Kind == lexer.TokLParen {
		p.next()
		for p.tok.Kind != lexer.TokRParen && p.tok.Kind != lexer.TokEOF {
			if p.tok.Kind == lexer.TokIdent {
				arm.Bindings = append(arm.Bindings, p.tok.Value)
			}
			if p.tok.Kind == lexer.TokComma {
				p.next()
				continue
			}
			p.next()
		}
		p.expect(lexer.TokRParen)
	}
	p.expect(lexer.TokArrow)
	var body strings.Builder
	depth := 0
	for p.tok.Kind != lexer.TokEOF {
		if depth == 0 && p.tok.Kind == lexer.TokRBrace {
			break
		}
		if depth == 0 && p.tok.Kind == lexer.TokIdent {
			if p.peekIsArmStart() {
				break
			}
		}
		if p.tok.Kind == lexer.TokLBrace {
			depth++
		}
		if p.tok.Kind == lexer.TokRBrace {
			depth--
		}
		body.WriteString(p.tok.Value)
		body.WriteString(" ")
		p.next()
	}
	arm.Body = strings.TrimSpace(body.String())
	return arm
}

func (p *Parser) peekIsArmStart() bool {
	savedLex := *p.lex
	savedTok := p.tok
	p.next()
	isArm := false
	if p.tok.Kind == lexer.TokArrow {
		isArm = true
	} else if p.tok.Kind == lexer.TokLParen {
		parenDepth := 1
		p.next()
		for parenDepth > 0 && p.tok.Kind != lexer.TokEOF {
			if p.tok.Kind == lexer.TokLParen {
				parenDepth++
			}
			if p.tok.Kind == lexer.TokRParen {
				parenDepth--
			}
			if parenDepth > 0 {
				p.next()
			}
		}
		if parenDepth == 0 {
			p.next()
			isArm = p.tok.Kind == lexer.TokArrow
		}
	}
	*p.lex = savedLex
	p.tok = savedTok
	return isArm
}
