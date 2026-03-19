package parser

import "github.com/liamp/gox/ast"

func (p *Parser) parseFunc() *ast.Func {
	p.next()
	p.errorf("func parsing not yet implemented")
	return &ast.Func{}
}
