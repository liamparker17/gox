package analyzer

import (
	"fmt"

	"github.com/liamp/gox/ast"
)

func (a *analyzer) resolveMatchType(m *ast.Match) {
	if _, ok := a.sumTypes[m.TypeName]; !ok {
		a.addError(m.Position, "type", fmt.Sprintf("unknown sum type %q", m.TypeName))
	}
}
