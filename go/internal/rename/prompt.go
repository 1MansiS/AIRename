package rename

import (
	"strings"
)

func BuildPrompt(ctx *VarContext) string {
	var b strings.Builder

	b.WriteString("You are a senior Go engineer writing production-grade code.\n\n")
	b.WriteString("Your task is to suggest better variable names.\n\n")

	b.WriteString("Variable to rename:\n")
	b.WriteString("- Name: " + ctx.VarName + "\n")
	b.WriteString("- Scope: " + ctx.Scope + "\n")
	b.WriteString("- Kind: " + ctx.Kind + "\n")
	b.WriteString("- Type: " + ctx.VarType + "\n\n")

	b.WriteString("Context:\n-----------\n")
	b.WriteString("Package: " + ctx.PackageName + "\n\n")

	b.WriteString("Function:\n")
	b.WriteString("- Name: " + ctx.FunctionName + "\n")
	if ctx.FunctionSummary != "" {
		b.WriteString("- Summary: " + ctx.FunctionSummary + "\n")
	}
	b.WriteString("\n")

	b.WriteString("Assignments:\n")
	if len(ctx.Assignments) == 0 {
		b.WriteString("- none\n")
	}
	for _, a := range ctx.Assignments {
		b.WriteString("- " + a + "\n")
	}

	b.WriteString("\nUsages:\n")
	if len(ctx.Usages) == 0 {
		b.WriteString("- none\n")
	}
	for _, u := range ctx.Usages {
		b.WriteString("- " + u + "\n")
	}

	b.WriteString("\nRelated Identifiers:\n")
	if len(ctx.RelatedIdentifiers) == 0 {
		b.WriteString("- none\n")
	}
	for _, r := range ctx.RelatedIdentifiers {
		b.WriteString("- " + r + "\n")
	}

	b.WriteString("\nImports in Scope:\n")
	if len(ctx.Imports) == 0 {
		b.WriteString("- none\n")
	}
	for _, i := range ctx.Imports {
		b.WriteString("- " + i + "\n")
	}

	b.WriteString("\nFile Comments:\n")
	if len(ctx.FileComments) == 0 {
		b.WriteString("- none\n")
	}
	for _, c := range ctx.FileComments {
		b.WriteString("- " + c + "\n")
	}

	b.WriteString(CodeStylePolicy)

	return b.String()
}

type TypeContext struct {
	PackageName string
	TypeName    string
	Fields      []string
	StructDoc   string
}

func BuildTypePrompt(ctx *TypeContext) string {
	var b strings.Builder

	b.WriteString("You are a senior Go engineer writing production-grade code.\n\n")
	b.WriteString("Your task is to suggest better struct type names.\n\n")

	b.WriteString("Type to rename:\n")
	b.WriteString("- Name: " + ctx.TypeName + "\n\n")

	b.WriteString("Context:\n-----------\n")
	b.WriteString("Package: " + ctx.PackageName + "\n\n")

	if ctx.StructDoc != "" {
		b.WriteString("Struct doc: " + ctx.StructDoc + "\n\n")
	}

	b.WriteString("Fields:\n")
	if len(ctx.Fields) == 0 {
		b.WriteString("- none\n")
	}
	for _, f := range ctx.Fields {
		b.WriteString("- " + f + "\n")
	}

	b.WriteString(CodeStylePolicy)

	return b.String()
}

func BuildFieldPrompt(ctx *FieldContext) string {
	var b strings.Builder

	b.WriteString("You are a senior Go engineer writing production-grade code.\n\n")
	b.WriteString("Your task is to suggest better struct field names.\n\n")

	b.WriteString("Field to rename:\n")
	b.WriteString("- Name: " + ctx.FieldName + "\n")
	b.WriteString("- Type: " + ctx.FieldType + "\n")
	b.WriteString("- Struct: " + ctx.StructName + "\n\n")

	b.WriteString("Context:\n-----------\n")
	b.WriteString("Package: " + ctx.PackageName + "\n\n")

	if ctx.StructDoc != "" {
		b.WriteString("Struct doc: " + ctx.StructDoc + "\n\n")
	}

	b.WriteString("Usages (file:line:col):\n")
	if len(ctx.Usages) == 0 {
		b.WriteString("- none\n")
	}
	for _, u := range ctx.Usages {
		b.WriteString("- " + u + "\n")
	}

	b.WriteString(CodeStylePolicy)

	return b.String()
}
