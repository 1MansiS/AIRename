package rename

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func findEnclosingFuncName(file *ast.File, pos token.Pos) string {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Body.Pos() <= pos && pos <= fn.Body.End() {
			return fn.Name.Name
		}
	}
	return ""
}

// findStructForField checks whether ident is (or refers to) a struct field.
// It first tries pointer equality (cursor on the declaration), then falls back
// to name matching for selector .Sel idents (cursor on a usage, Obj == nil).
func findStructForField(file *ast.File, ident *ast.Ident) (string, bool) {
	var structName string
	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if name == ident || (ident.Obj == nil && name.Name == ident.Name) {
					structName = typeSpec.Name.Name
					return false
				}
			}
		}
		return true
	})
	return structName, structName != ""
}

func Run(filename string, selector Selector, provider string) (*Result, error) {
	var prompt string
	var name string

	if selector.Kind == "position" {
		file, _, _, ident, err := ResolveSelector(filename, selector)
		if err != nil {
			return nil, err
		}
		name = ident.Name

		if structName, ok := findStructForField(file, ident); ok {
			ctx, _, _, err := BuildFieldContext(filename, structName, name)
			if err != nil {
				return nil, err
			}
			prompt = BuildFieldPrompt(ctx)
		} else if ident.Obj != nil && ident.Obj.Kind == ast.Typ {
			typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
			if !ok {
				return nil, fmt.Errorf("type declaration not found for %q", name)
			}
			typeCtx := buildTypeContext(file, typeSpec)
			prompt = BuildTypePrompt(typeCtx)
		} else {
			funcName := findEnclosingFuncName(file, ident.Pos())
			ctx, _, _, err := BuildVarContext(filename, funcName, name)
			if err != nil {
				return nil, err
			}
			prompt = BuildPrompt(ctx)
		}
	} else {
		// funcvar path
		name = selector.Var
		ctx, _, _, err := BuildVarContext(filename, selector.Func, name)
		if err != nil {
			return nil, err
		}
		prompt = BuildPrompt(ctx)
	}

	lines, err := CallLLM(prompt, provider)
	if err != nil {
		return nil, err
	}

	var suggestions []Suggestion
	for _, l := range lines {
		parts := splitOnce(l, " - ")
		if len(parts) == 2 {
			suggestions = append(suggestions, Suggestion{
				Name:   strings.TrimSpace(parts[0]),
				Reason: strings.TrimSpace(parts[1]),
			})
		}
	}

	if len(suggestions) == 0 {
		return nil, fmt.Errorf("no valid suggestions from LLM")
	}

	return &Result{
		Suggestions: suggestions,
		Debug: Debug{
			Prompt: prompt,
		},
	}, nil
}

func splitOnce(s, sep string) []string {
	if idx := strings.Index(s, sep); idx >= 0 {
		return []string{s[:idx], s[idx+len(sep):]}
	}
	return nil
}

// buildTypeContext builds a TypeContext from a resolved TypeSpec.
func buildTypeContext(file *ast.File, typeSpec *ast.TypeSpec) *TypeContext {
	ctx := &TypeContext{
		PackageName: file.Name.Name,
		TypeName:    typeSpec.Name.Name,
	}
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return ctx
	}
	for _, field := range structType.Fields.List {
		for _, fname := range field.Names {
			ctx.Fields = append(ctx.Fields, fname.Name)
		}
	}
	// Find the enclosing GenDecl to pick up the doc comment.
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		for _, spec := range genDecl.Specs {
			if spec == typeSpec && genDecl.Doc != nil {
				ctx.StructDoc = strings.TrimSpace(genDecl.Doc.Text())
				return false
			}
		}
		return true
	})
	return ctx
}
