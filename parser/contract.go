package parser

import (
	"strconv"

	"github.com/liamp/gox/ast"
	"github.com/liamp/gox/lexer"
)

func (p *Parser) parseContract() *ast.Contract {
	pos := p.tok.Pos
	p.next()
	name := p.expect(lexer.TokIdent)
	p.expect(lexer.TokLBrace)
	c := &ast.Contract{Position: pos, Name: name.Value}
	for p.tok.Kind != lexer.TokRBrace && p.tok.Kind != lexer.TokEOF {
		switch p.tok.Kind {
		case lexer.TokInput:
			p.next()
			p.expect(lexer.TokLBrace)
			c.Input = p.parseAnnotatedFields()
			p.expect(lexer.TokRBrace)
		case lexer.TokOutput:
			p.next()
			p.expect(lexer.TokLBrace)
			c.Output = p.parsePlainFields()
			p.expect(lexer.TokRBrace)
		case lexer.TokErrors:
			p.next()
			p.expect(lexer.TokLBrace)
			c.Errors = p.parseContractErrors()
			p.expect(lexer.TokRBrace)
		case lexer.TokRoute:
			p.next()
			method := p.expect(lexer.TokIdent)
			path := p.expect(lexer.TokString)
			c.Route = &ast.Route{Method: method.Value, Path: path.Value}
		default:
			p.errorf("unexpected token in contract: %s (%q)", p.tok.Kind, p.tok.Value)
			p.next()
		}
	}
	p.expect(lexer.TokRBrace)
	return c
}

func (p *Parser) parseAnnotatedFields() []ast.AnnotatedField {
	var fields []ast.AnnotatedField
	for p.tok.Kind == lexer.TokIdent {
		name := p.expect(lexer.TokIdent)
		p.expect(lexer.TokColon)
		typ := p.expect(lexer.TokIdent)
		af := ast.AnnotatedField{Field: ast.Field{Name: name.Value, Type: typ.Value}}
		for p.tok.Kind == lexer.TokAt {
			p.next()
			aName := p.expect(lexer.TokIdent)
			ann := ast.Annotation{Name: aName.Value}
			if p.tok.Kind == lexer.TokLParen {
				p.next()
				for p.tok.Kind != lexer.TokRParen && p.tok.Kind != lexer.TokEOF {
					ann.Args = append(ann.Args, p.tok.Value)
					p.next()
				}
				p.expect(lexer.TokRParen)
			}
			af.Annotations = append(af.Annotations, ann)
		}
		fields = append(fields, af)
	}
	return fields
}

func (p *Parser) parsePlainFields() []ast.Field {
	var fields []ast.Field
	for p.tok.Kind == lexer.TokIdent {
		name := p.expect(lexer.TokIdent)
		p.expect(lexer.TokColon)
		typ := p.expect(lexer.TokIdent)
		fields = append(fields, ast.Field{Name: name.Value, Type: typ.Value})
	}
	return fields
}

func (p *Parser) parseContractErrors() []ast.ContractError {
	var errs []ast.ContractError
	for p.tok.Kind == lexer.TokIdent {
		name := p.expect(lexer.TokIdent)
		ce := ast.ContractError{Name: name.Value, StatusCode: 400}
		if p.tok.Kind == lexer.TokNumber {
			if code, err := strconv.Atoi(p.tok.Value); err == nil {
				ce.StatusCode = code
			}
			p.next()
		}
		errs = append(errs, ce)
	}
	return errs
}
