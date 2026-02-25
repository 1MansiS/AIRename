package rename

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// VarContext holds full context for a variable to be renamed
type VarContext struct {
	PackageName string
	Filename    string

	FunctionName    string
	FunctionSummary string

	VarName string
	VarType string
	Scope   string // function | file
	Kind    string // local variable | parameter

	Assignments        []string
	Usages             []string
	RelatedIdentifiers []string
	Imports            []string
	FileComments       []string
}

// BuildVarContext parses the file and builds a rich context for the variable
func BuildVarContext(filename, funcName, varName string) (*VarContext, *ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx := &VarContext{
		Filename:    filename,
		VarName:     varName,
		Scope:       "function",
		Kind:        "local variable",
		PackageName: file.Name.Name,
	}

	// file-level comments
	if file.Doc != nil {
		for _, c := range file.Doc.List {
			ctx.FileComments = append(ctx.FileComments, c.Text)
		}
	}

	// imports
	for _, imp := range file.Imports {
		ctx.Imports = append(ctx.Imports, trimQuotes(imp.Path.Value))
	}

	// find the target function
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || (funcName != "" && fn.Name.Name != funcName) {
			continue
		}

		ctx.FunctionName = fn.Name.Name
		ctx.FunctionSummary = extractFuncSummary(fn)

		// parameters
		if fn.Type.Params != nil {
			for _, field := range fn.Type.Params.List {
				for _, name := range field.Names {
					ctx.RelatedIdentifiers = append(ctx.RelatedIdentifiers, name.Name)
				}
			}
		}

		// assignments & usages
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range x.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && ident.Name == varName {
						ctx.Assignments = append(ctx.Assignments, fmt.Sprintf("%s %s", x.Tok.String(), ident.Name))
					}
				}
			case *ast.Ident:
				if x.Name == varName {
					pos := fset.Position(x.Pos())
					ctx.Usages = append(ctx.Usages, fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column))
				}
			}
			return true
		})

		// type inference
		ctx.VarType = inferIdentTypeByName(fn.Body, varName)

		return ctx, file, fset, nil
	}

	return nil, nil, nil, fmt.Errorf("function %q not found", funcName)
}

func extractFuncSummary(fn *ast.FuncDecl) string {
	if fn.Doc == nil {
		return ""
	}
	return strings.TrimSpace(fn.Doc.Text())
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

// simple type inference stub
func inferIdentTypeByName(block *ast.BlockStmt, name string) string {
	var typ string
	ast.Inspect(block, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name == name {
				if bl, ok := assign.Rhs[i].(*ast.BasicLit); ok {
					typ = bl.Kind.String()
					return false
				}
			}
		}
		return true
	})
	if typ == "" {
		typ = "unknown"
	}
	return typ
}
