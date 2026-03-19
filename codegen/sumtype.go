package codegen

import (
	"fmt"
	"strings"

	"github.com/liamp/gox/ast"
)

func (g *Generator) emitSumType(st *ast.SumType) string {
	var buf strings.Builder
	marker := fmt.Sprintf("is%s", st.Name)
	buf.WriteString(fmt.Sprintf("\ntype %s interface {\n\t%s()\n}\n", st.Name, marker))
	for _, v := range st.Variants {
		structName := st.Name + v.Name
		if len(v.Fields) == 0 {
			buf.WriteString(fmt.Sprintf("\ntype %s struct{}\n", structName))
		} else {
			buf.WriteString(fmt.Sprintf("\ntype %s struct {\n", structName))
			for _, f := range v.Fields {
				buf.WriteString(fmt.Sprintf("\t%s %s\n", capitalize(f.Name), f.Type))
			}
			buf.WriteString("}\n")
		}
		buf.WriteString(fmt.Sprintf("\nfunc (%s) %s() {}\n", structName, marker))
	}
	return buf.String()
}
