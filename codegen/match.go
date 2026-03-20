package codegen

import (
	"fmt"
	"strings"

	"github.com/liamparker17/gox/ast"
)

func (g *Generator) emitFunc(f *ast.Func) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("\n%s {\n", f.Signature))

	for _, stmt := range f.Stmts {
		switch s := stmt.(type) {
		case *ast.GoCode:
			buf.WriteString("\t" + s.Code + "\n")
		case *ast.Match:
			buf.WriteString(g.emitMatch(s))
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

func (g *Generator) emitMatch(m *ast.Match) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("\tswitch v := %s.(type) {\n", m.Expr))

	for _, arm := range m.Arms {
		typeName := m.TypeName
		structName := typeName + arm.Variant

		buf.WriteString(fmt.Sprintf("\tcase %s:\n", structName))

		// Replace bindings with v.FieldName
		body := arm.Body
		if len(arm.Bindings) > 0 {
			// We need the sum type to know field names — store it during Generate
			if st, ok := g.sumTypes[typeName]; ok {
				for _, v := range st.Variants {
					if v.Name == arm.Variant {
						for i, binding := range arm.Bindings {
							if i < len(v.Fields) {
								body = strings.ReplaceAll(body, binding, "v."+capitalize(v.Fields[i].Name))
							}
						}
						break
					}
				}
			}
		}

		buf.WriteString(fmt.Sprintf("\t\t%s\n", body))
	}

	if m.Ignore {
		buf.WriteString("\tdefault:\n\t\tpanic(\"unhandled variant\")\n")
	}

	buf.WriteString("\t}\n")
	return buf.String()
}
