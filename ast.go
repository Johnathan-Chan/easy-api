package easy_api

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ScanComments 扫描注解
func ScanComments(filepath string, fileSet *token.FileSet)([]ast.Decl, error){
	fileParse, err := parser.ParseFile(fileSet, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return fileParse.Decls, nil
}
