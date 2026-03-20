package codegen

import (
	"fmt"
	"strings"

	"github.com/liamp/gox/ast"
)

func (g *Generator) emitContract(c *ast.Contract) string {
	var buf strings.Builder

	// Input struct
	buf.WriteString(fmt.Sprintf("\ntype %sInput struct {\n", c.Name))
	for _, f := range c.Input {
		tag := f.Name
		if hasAnnotation(f.Annotations, "optional") {
			tag += ",omitempty"
		}
		buf.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", capitalize(f.Name), f.Type, tag))
	}
	buf.WriteString("}\n")

	// Output struct
	buf.WriteString(fmt.Sprintf("\ntype %sOutput struct {\n", c.Name))
	for _, f := range c.Output {
		buf.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", capitalize(f.Name), f.Type, f.Name))
	}
	buf.WriteString("}\n")

	// Error sum type
	if len(c.Errors) > 0 {
		marker := fmt.Sprintf("is%sError", c.Name)
		buf.WriteString(fmt.Sprintf("\ntype %sError interface {\n\t%s()\n}\n", c.Name, marker))
		for _, e := range c.Errors {
			structName := c.Name + "Error" + e.Name
			buf.WriteString(fmt.Sprintf("\ntype %s struct{}\n", structName))
			buf.WriteString(fmt.Sprintf("\nfunc (%s) %s() {}\n", structName, marker))
		}
	}

	// Validation function
	buf.WriteString(g.emitValidation(c))

	// Handler + Client (only if route is set)
	if c.Route != nil {
		buf.WriteString(g.emitHandler(c))
		buf.WriteString(g.emitClient(c))
	}

	return buf.String()
}

func (g *Generator) emitValidation(c *ast.Contract) string {
	g.imports["fmt"] = true
	g.imports["strings"] = true

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\nfunc Validate%sInput(in %sInput) error {\n", c.Name, c.Name))
	buf.WriteString("\tvar errs []string\n")

	for _, f := range c.Input {
		fieldAccess := "in." + capitalize(f.Name)
		for _, ann := range f.Annotations {
			switch ann.Name {
			case "required":
				buf.WriteString(g.emitRequiredCheck(f.Field, fieldAccess))
			case "email":
				buf.WriteString(fmt.Sprintf("\tif %s != \"\" && !strings.Contains(%s, \"@\") {\n\t\terrs = append(errs, \"%s: invalid email\")\n\t}\n", fieldAccess, fieldAccess, f.Name))
			case "minlen":
				if len(ann.Args) > 0 {
					buf.WriteString(fmt.Sprintf("\tif %s != \"\" && len(%s) < %s {\n\t\terrs = append(errs, \"%s: min length %s\")\n\t}\n", fieldAccess, fieldAccess, ann.Args[0], f.Name, ann.Args[0]))
				}
			case "maxlen":
				if len(ann.Args) > 0 {
					buf.WriteString(fmt.Sprintf("\tif len(%s) > %s {\n\t\terrs = append(errs, \"%s: max length %s\")\n\t}\n", fieldAccess, ann.Args[0], f.Name, ann.Args[0]))
				}
			case "min":
				if len(ann.Args) > 0 {
					buf.WriteString(fmt.Sprintf("\tif %s < %s {\n\t\terrs = append(errs, \"%s: must be at least %s\")\n\t}\n", fieldAccess, ann.Args[0], f.Name, ann.Args[0]))
				}
			case "max":
				if len(ann.Args) > 0 {
					buf.WriteString(fmt.Sprintf("\tif %s > %s {\n\t\terrs = append(errs, \"%s: must be at most %s\")\n\t}\n", fieldAccess, ann.Args[0], f.Name, ann.Args[0]))
				}
			}
		}
	}

	buf.WriteString("\tif len(errs) > 0 {\n\t\treturn fmt.Errorf(\"%s\", strings.Join(errs, \"; \"))\n\t}\n")
	buf.WriteString("\treturn nil\n}\n")
	return buf.String()
}

func (g *Generator) emitRequiredCheck(f ast.Field, access string) string {
	switch {
	case f.Type == "string":
		return fmt.Sprintf("\tif %s == \"\" {\n\t\terrs = append(errs, \"%s: required\")\n\t}\n", access, f.Name)
	case f.Type == "int" || f.Type == "int64" || f.Type == "float64":
		return fmt.Sprintf("\tif %s == 0 {\n\t\terrs = append(errs, \"%s: required\")\n\t}\n", access, f.Name)
	case f.Type == "time.Time":
		return fmt.Sprintf("\tif %s.IsZero() {\n\t\terrs = append(errs, \"%s: required\")\n\t}\n", access, f.Name)
	default:
		return fmt.Sprintf("\tif %s == nil {\n\t\terrs = append(errs, \"%s: required\")\n\t}\n", access, f.Name)
	}
}

func (g *Generator) emitHandler(c *ast.Contract) string {
	g.imports["encoding/json"] = true
	g.imports["net/http"] = true

	var buf strings.Builder
	errType := c.Name + "Error"

	buf.WriteString(fmt.Sprintf("\nfunc %sHandler(fn func(%sInput) (%sOutput, %s)) http.HandlerFunc {\n", c.Name, c.Name, c.Name, errType))
	buf.WriteString("\treturn func(w http.ResponseWriter, r *http.Request) {\n")
	buf.WriteString("\t\tw.Header().Set(\"Content-Type\", \"application/json\")\n")
	buf.WriteString(fmt.Sprintf("\t\tvar in %sInput\n", c.Name))
	buf.WriteString("\t\tif err := json.NewDecoder(r.Body).Decode(&in); err != nil {\n")
	buf.WriteString("\t\t\tw.WriteHeader(400)\n")
	buf.WriteString("\t\t\tjson.NewEncoder(w).Encode(map[string]string{\"error\": err.Error()})\n")
	buf.WriteString("\t\t\treturn\n\t\t}\n")
	buf.WriteString(fmt.Sprintf("\t\tif err := Validate%sInput(in); err != nil {\n", c.Name))
	buf.WriteString("\t\t\tw.WriteHeader(422)\n")
	buf.WriteString("\t\t\tjson.NewEncoder(w).Encode(map[string]string{\"error\": err.Error()})\n")
	buf.WriteString("\t\t\treturn\n\t\t}\n")
	buf.WriteString("\t\tout, cerr := fn(in)\n")
	buf.WriteString("\t\tif cerr != nil {\n")
	buf.WriteString("\t\t\tswitch cerr.(type) {\n")
	for _, e := range c.Errors {
		structName := c.Name + "Error" + e.Name
		msg := camelToMessage(e.Name)
		buf.WriteString(fmt.Sprintf("\t\t\tcase %s:\n", structName))
		buf.WriteString(fmt.Sprintf("\t\t\t\tw.WriteHeader(%d)\n", e.StatusCode))
		buf.WriteString(fmt.Sprintf("\t\t\t\tjson.NewEncoder(w).Encode(map[string]string{\"error\": %q})\n", msg))
	}
	buf.WriteString("\t\t\t}\n\t\t\treturn\n\t\t}\n")
	buf.WriteString("\t\tjson.NewEncoder(w).Encode(out)\n")
	buf.WriteString("\t}\n}\n")

	return buf.String()
}

func (g *Generator) emitClient(c *ast.Contract) string {
	g.imports["encoding/json"] = true
	g.imports["net/http"] = true
	g.imports["bytes"] = true
	g.imports["fmt"] = true

	errType := c.Name + "Error"

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\nfunc %sClient(baseURL string, in %sInput) (%sOutput, %s, error) {\n", c.Name, c.Name, c.Name, errType))
	buf.WriteString("\tbody, err := json.Marshal(in)\n")
	buf.WriteString(fmt.Sprintf("\tif err != nil {\n\t\treturn %sOutput{}, nil, fmt.Errorf(\"marshal: %%w\", err)\n\t}\n", c.Name))
	buf.WriteString(fmt.Sprintf("\tresp, err := http.Post(baseURL+\"%s\", \"application/json\", bytes.NewReader(body))\n", c.Route.Path))
	buf.WriteString(fmt.Sprintf("\tif err != nil {\n\t\treturn %sOutput{}, nil, fmt.Errorf(\"request: %%w\", err)\n\t}\n", c.Name))
	buf.WriteString("\tdefer resp.Body.Close()\n")
	buf.WriteString("\tswitch resp.StatusCode {\n")
	buf.WriteString("\tcase 200:\n")
	buf.WriteString(fmt.Sprintf("\t\tvar out %sOutput\n", c.Name))
	buf.WriteString("\t\tif err := json.NewDecoder(resp.Body).Decode(&out); err != nil {\n")
	buf.WriteString(fmt.Sprintf("\t\t\treturn %sOutput{}, nil, fmt.Errorf(\"decode: %%w\", err)\n\t\t}\n", c.Name))
	buf.WriteString("\t\treturn out, nil, nil\n")
	for _, e := range c.Errors {
		structName := c.Name + "Error" + e.Name
		buf.WriteString(fmt.Sprintf("\tcase %d:\n", e.StatusCode))
		buf.WriteString(fmt.Sprintf("\t\treturn %sOutput{}, %s{}, nil\n", c.Name, structName))
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn %sOutput{}, nil, fmt.Errorf(\"unexpected status: %%d\", resp.StatusCode)\n", c.Name))
	buf.WriteString("\t}\n}\n")

	return buf.String()
}

func hasAnnotation(anns []ast.Annotation, name string) bool {
	for _, a := range anns {
		if a.Name == name {
			return true
		}
	}
	return false
}

func camelToMessage(s string) string {
	var result strings.Builder
	for i, ch := range s {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			result.WriteByte(' ')
		}
		result.WriteRune(ch)
	}
	return strings.ToLower(result.String())
}
