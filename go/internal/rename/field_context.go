package rename

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// FieldContext holds context for a struct field to be renamed
type FieldContext struct {
	PackageName string
	Filename    string
	StructName  string
	FieldName   string
	FieldType   string
	StructDoc   string
	Usages      []string // "filepath:line:col", 1-based — declaration first, then selector sites
}

// BuildFieldContext parses the file and builds context for a struct field
func BuildFieldContext(filename, structName, fieldName string) (*FieldContext, *ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx := &FieldContext{
		Filename:    filename,
		PackageName: file.Name.Name,
		FieldName:   fieldName,
		StructName:  structName,
	}

	// Find the struct declaration and record the field's declaration position
	found := false
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != structName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if genDecl.Doc != nil {
				ctx.StructDoc = strings.TrimSpace(genDecl.Doc.Text())
			}
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if name.Name == fieldName {
						ctx.FieldType = fieldTypeStr(field.Type)
						pos := fset.Position(name.Pos())
						ctx.Usages = append(ctx.Usages, fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column))
						found = true
					}
				}
			}
		}
	}

	if !found {
		return nil, nil, nil, fmt.Errorf("field %q not found in struct %q", fieldName, structName)
	}

	// Collect all selector expression usages (x.FieldName) throughout the file
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == fieldName {
			pos := fset.Position(sel.Sel.Pos())
			ctx.Usages = append(ctx.Usages, fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column))
		}
		return true
	})

	return ctx, file, fset, nil
}

func fieldTypeStr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + fieldTypeStr(t.X)
	case *ast.ArrayType:
		return "[]" + fieldTypeStr(t.Elt)
	default:
		return "unknown"
	}
}
