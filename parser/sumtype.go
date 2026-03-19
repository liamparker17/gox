package parser

import (
	"github.com/liamp/gox/ast"
	"github.com/liamp/gox/lexer"
)

func (p *Parser) parseSumType() *ast.SumType {
	pos := p.tok.Pos
	p.next()
	name := p.expect(lexer.TokIdent)
	p.expect(lexer.TokLBrace)
	st := &ast.SumType{
		Position: pos,
		Name:     name.Value,
	}
	seen := map[string]bool{}
	for p.tok.Kind != lexer.TokRBrace && p.tok.Kind != lexer.TokEOF {
		v := p.parseVariant()
		if seen[v.Name] {
			p.errorf("duplicate variant %q in sumtype %s", v.Name, st.Name)
		}
		seen[v.Name] = true
		st.Variants = append(st.Variants, v)
	}
	p.expect(lexer.TokRBrace)
	return st
}

func (p *Parser) parseVariant() ast.Variant {
	pos := p.tok.Pos
	name := p.expect(lexer.TokIdent)
	v := ast.Variant{Position: pos, Name: name.Value}
	if p.tok.Kind == lexer.TokLBrace {
		p.next()
		for p.tok.Kind != lexer.TokRBrace && p.tok.Kind != lexer.TokEOF {
			field := p.parseField()
			v.Fields = append(v.Fields, field)
			if p.tok.Kind == lexer.TokComma {
				p.next()
			}
		}
		p.expect(lexer.TokRBrace)
	}
	return v
}

func (p *Parser) parseField() ast.Field {
	name := p.expect(lexer.TokIdent)
	p.expect(lexer.TokColon)
	typ := p.expect(lexer.TokIdent)
	return ast.Field{Name: name.Value, Type: typ.Value}
}
