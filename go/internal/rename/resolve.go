package rename

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type Selector struct {
	Kind string // "funcvar" | "position"
	Func string
	Var  string
	Row  int
	Col  int
}

func ResolveSelector(
	filename string,
	selector Selector,
) (*ast.File, *token.FileSet, *token.File, *ast.Ident, error) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	tokFile := fset.File(file.Pos())
	if tokFile == nil {
		return nil, nil, nil, nil, fmt.Errorf("token file not found")
	}

	switch selector.Kind {
	case "funcvar":
		id, err := resolveFuncVar(file, selector.Func, selector.Var)
		return file, fset, tokFile, id, err

	case "position":
		id, err := resolvePosition(tokFile, file, selector.Row, selector.Col)
		return file, fset, tokFile, id, err

	default:
		return nil, nil, nil, nil, fmt.Errorf("unknown selector kind")
	}
}

func resolveFuncVar(
	file *ast.File,
	funcName, varName string,
) (*ast.Ident, error) {

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != funcName {
			continue
		}

		var found *ast.Ident
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if ok && id.Name == varName && id.Obj != nil {
				found = id
				return false
			}
			return true
		})

		if found != nil {
			return found, nil
		}
	}

	return nil, fmt.Errorf("variable not found in function")
}

func resolvePosition(
	tokFile *token.File,
	file *ast.File,
	row, col int,
) (*ast.Ident, error) {

	lineStart := tokFile.LineStart(row)
	target := lineStart + token.Pos(col)
	targetOffset := tokFile.Offset(target)

	var best *ast.Ident

	// First pass: prefer idents with Obj set (local vars, declarations)
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Obj == nil {
			return true
		}
		start := tokFile.Offset(id.Pos())
		end := tokFile.Offset(id.End())
		if start <= targetOffset && targetOffset <= end {
			best = id
			return false
		}
		return true
	})

	if best != nil {
		return best, nil
	}

	// Second pass: check selector .Sel idents (struct field usages have no Obj)
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id := sel.Sel
		start := tokFile.Offset(id.Pos())
		end := tokFile.Offset(id.End())
		if start <= targetOffset && targetOffset <= end {
			best = id
			return false
		}
		return true
	})

	if best == nil {
		return nil, fmt.Errorf("no identifier at position")
	}

	return best, nil
}
