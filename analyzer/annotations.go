package analyzer

import (
	"fmt"

	"github.com/liamp/gox/ast"
)

var validAnnotations = map[string]int{
	"required": 0, "optional": 0, "email": 0,
	"minlen": 1, "maxlen": 1, "min": 1, "max": 1,
}

func (a *analyzer) validateAnnotations(c *ast.Contract) {
	for _, field := range c.Input {
		for _, ann := range field.Annotations {
			expectedArgs, ok := validAnnotations[ann.Name]
			if !ok {
				a.addError(c.Position, "annotation",
					fmt.Sprintf("unknown annotation @%s on field %q", ann.Name, field.Name))
				continue
			}
			if len(ann.Args) != expectedArgs {
				a.addError(c.Position, "annotation",
					fmt.Sprintf("@%s expects %d args, got %d", ann.Name, expectedArgs, len(ann.Args)))
			}
		}
	}
}

func (a *analyzer) validateRoute(c *ast.Contract) {
	if c.Route == nil {
		return
	}
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	}
	if !validMethods[c.Route.Method] {
		a.addError(c.Position, "route", fmt.Sprintf("invalid HTTP method %q", c.Route.Method))
	}
	if len(c.Route.Path) == 0 || c.Route.Path[0] != '/' {
		a.addError(c.Position, "route", fmt.Sprintf("route path must start with /, got %q", c.Route.Path))
	}
}
