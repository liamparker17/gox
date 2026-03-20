package analyzer

import (
	"fmt"
	"strings"

	"github.com/liamparker17/gox/ast"
)

func (a *analyzer) checkExhaustiveness(m *ast.Match) {
	if m.Ignore {
		return
	}
	st, ok := a.sumTypes[m.TypeName]
	if !ok {
		return
	}
	handled := make(map[string]bool)
	for _, arm := range m.Arms {
		handled[arm.Variant] = true
	}
	var missing []string
	for _, v := range st.Variants {
		if !handled[v.Name] {
			missing = append(missing, v.Name)
		}
	}
	if len(missing) > 0 {
		a.addError(m.Position, "exhaustiveness",
			fmt.Sprintf("match on %s missing variant(s): %s", m.TypeName, strings.Join(missing, ", ")))
	}
}

func (a *analyzer) checkBindings(m *ast.Match) {
	st, ok := a.sumTypes[m.TypeName]
	if !ok {
		return
	}
	variantFields := make(map[string]int)
	for _, v := range st.Variants {
		variantFields[v.Name] = len(v.Fields)
	}
	for _, arm := range m.Arms {
		expected, ok := variantFields[arm.Variant]
		if !ok {
			a.addError(m.Position, "type",
				fmt.Sprintf("unknown variant %q in match on %s", arm.Variant, m.TypeName))
			continue
		}
		if len(arm.Bindings) > 0 && len(arm.Bindings) != expected {
			a.addError(m.Position, "binding",
				fmt.Sprintf("variant %s has %d fields but %d bindings provided", arm.Variant, expected, len(arm.Bindings)))
		}
	}
}
