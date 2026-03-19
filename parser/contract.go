package parser

import "github.com/liamp/gox/ast"

func (p *Parser) parseContract() *ast.Contract {
	p.next()
	p.errorf("contract parsing not yet implemented")
	return &ast.Contract{}
}
