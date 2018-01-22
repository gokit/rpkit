package rp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gokit/rpkit/static"

	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
)

// InterfaceRP implements code generation for creating RPC based API which work with either
// HTTP 1.0 or other underline transport system.
func InterfaceRP(toPackage string, an ast.AnnotationDeclaration, in ast.InterfaceDeclaration, declr ast.PackageDeclaration, pkg ast.Package) ([]gen.WriteDirective, error) {
	packageName := fmt.Sprintf("%srp", strings.ToLower(in.Object.Name.Name))
	packageFileName := fmt.Sprintf("%s.rp.go", strings.ToLower(in.Object.Name.Name))
	packagePath := filepath.Join(toPackage, packageName)

	sections := strings.Split(strings.TrimSuffix(declr.Path, "/"), "/")
	sections = sections[1:]

	var noArgNoReturnMethods,
		onlyErrorMethods,
		outputWithNoErrorMethods,
		outputWithErrorMethods,
		inputWithOnlyErrorMethods,
		inputWithOutputOnlyMethods,
		inputWithOutputWithErrorMethods []ast.FunctionDefinition

	methods := in.Methods(&declr)
	for _, method := range methods {
		if !method.HasReturns() && !method.HasArgs() {
			noArgNoReturnMethods = append(noArgNoReturnMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.HasArgType("context.Context") && !method.HasReturns() {
			noArgNoReturnMethods = append(noArgNoReturnMethods, method)
			continue
		}

		if !method.HasArgs() && method.HasReturnType("error") && method.TotalReturns() == 1 {
			onlyErrorMethods = append(onlyErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.HasArgType("context.Context") &&
			method.HasReturnType("error") && method.TotalReturns() == 1 {
			onlyErrorMethods = append(onlyErrorMethods, method)
			continue
		}

		if !method.HasArgs() && !method.HasReturnType("error") && method.TotalReturns() == 1 {
			outputWithNoErrorMethods = append(outputWithNoErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.HasArgType("context.Context") &&
			!method.HasReturnType("error") && method.TotalReturns() == 1 {
			outputWithNoErrorMethods = append(outputWithNoErrorMethods, method)
			continue
		}

		if !method.HasArgs() && method.HasReturnType("error") &&
			method.TotalReturns() == 2 && method.ReturnTypePos("error") == 1 {
			outputWithErrorMethods = append(outputWithErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.HasArgType("contenxt.Context") &&
			method.HasReturnType("error") &&
			method.TotalReturns() == 2 && method.ReturnTypePos("error") == 1 {
			outputWithErrorMethods = append(outputWithErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.TotalReturns() == 1 &&
			!method.HasArgType("context.Context") && method.HasReturnType("error") {
			inputWithOnlyErrorMethods = append(inputWithOnlyErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 2 && method.TotalReturns() == 1 &&
			method.ArgTypePos("context.Context") == 0 && method.HasReturnType("error") {
			inputWithOnlyErrorMethods = append(inputWithOnlyErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.TotalReturns() == 1 &&
			!method.HasArgType("context.Context") && !method.HasReturnType("error") {
			inputWithOutputOnlyMethods = append(inputWithOutputOnlyMethods, method)
			continue
		}

		if method.TotalArgs() == 2 && method.TotalReturns() == 1 &&
			method.ArgTypePos("context.Context") == 0 && !method.HasReturnType("error") {

			// if both return types are error skip, this is bad signature for now.
			if method.CountOfReturnType("error") > 1 {
				continue
			}

			inputWithOutputOnlyMethods = append(inputWithOutputOnlyMethods, method)
			continue
		}

		if method.TotalArgs() == 1 && method.TotalReturns() == 2 &&
			!method.HasArgType("context.Context") && method.ReturnTypePos("error") == 1 {

			// if both return types are error skip, this is bad signature for now.
			if method.CountOfReturnType("error") > 1 {
				continue
			}

			inputWithOutputWithErrorMethods = append(inputWithOutputWithErrorMethods, method)
			continue
		}

		if method.TotalArgs() == 2 && method.TotalReturns() == 2 &&
			method.ArgTypePos("context.Context") == 0 && method.ReturnTypePos("error") == 1 {

			// if both return types are error skip, this is bad signature for now.
			if method.CountOfArgType("context.Context") > 1 {
				continue
			}

			// if both return types are error skip, this is bad signature for now.
			if method.CountOfReturnType("error") > 1 {
				continue
			}

			inputWithOutputWithErrorMethods = append(inputWithOutputWithErrorMethods, method)
			continue
		}
	}

	iGen := gen.Package(
		gen.Name(packageName),
		gen.Imports(),
		gen.Block(
			gen.SourceText(
				"rpkit:interface_rpc",
				string(static.MustReadFile("interface_rpc.tml", true)),
				struct {
					ServiceName                    string
					TargetPackage                  string
					An                             ast.AnnotationDeclaration
					Itr                            ast.InterfaceDeclaration
					Pkg                            ast.PackageDeclaration
					NoArgAndReturns                []ast.FunctionDefinition
					OnlyErrorMethods               []ast.FunctionDefinition
					OutputWithErrorMethods         []ast.FunctionDefinition
					OutputWithNoErrorMethods       []ast.FunctionDefinition
					InputWithErrorMethods          []ast.FunctionDefinition
					InputAndOutputMethods          []ast.FunctionDefinition
					InputAndOutputWithErrorMethods []ast.FunctionDefinition
				}{
					An:                             an,
					Itr:                            in,
					Pkg:                            declr,
					TargetPackage:                  packagePath,
					OnlyErrorMethods:               onlyErrorMethods,
					NoArgAndReturns:                noArgNoReturnMethods,
					OutputWithNoErrorMethods:       outputWithNoErrorMethods,
					OutputWithErrorMethods:         outputWithErrorMethods,
					InputWithErrorMethods:          inputWithOnlyErrorMethods,
					InputAndOutputMethods:          inputWithOutputOnlyMethods,
					InputAndOutputWithErrorMethods: inputWithOutputWithErrorMethods,
					ServiceName:                    strings.Join(sections, "."),
				},
			),
		),
	)

	return []gen.WriteDirective{
		{
			Writer:   iGen,
			Dir:      packageName,
			FileName: packageFileName,
		},
	}, nil
}
