package generator

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"github.com/zylisp/gisp"
	"github.com/zylisp/gisp/parser"
	"go/ast"
	"go/printer"
	"go/token"
	"io/ioutil"
)

var log = logging.MustGetLogger(gisp.ApplicationName)
var anyType = makeSelectorExpr(ast.NewIdent("core"), ast.NewIdent("Any"))

func GenerateAST(tree []parser.Node) *ast.File {
	f := &ast.File{Name: ast.NewIdent("main")}
	decls := make([]ast.Decl, 0, len(tree))

	if len(tree) < 1 {
		return f
	}

	// you can only have (ns ...) as the first form
	if isNSDecl(tree[0]) {
		name, imports := getNamespace(tree[0].(*parser.CallNode))

		f.Name = name
		if imports != nil {
			decls = append(decls, imports)
		}

		tree = tree[1:]
	}

	decls = append(decls, generateDecls(tree)...)

	f.Decls = decls
	return f
}

func generateDecls(tree []parser.Node) []ast.Decl {
	decls := make([]ast.Decl, len(tree))

	for i, node := range tree {
		if node.Type() != parser.NodeCall {
			log.Critical(MissingCallNodeError)
			panic(MissingCallNodeError)
		}

		decls[i] = evalDeclNode(node.(*parser.CallNode))
	}

	return decls
}

func evalDeclNode(node *parser.CallNode) ast.Decl {
	// Let's just assume that all top-level functions called will be "def"
	if node.Callee.Type() != parser.NodeIdent {
		log.Critical(CalleeIndentifierMismatchError)
		panic(CalleeIndentifierMismatchError)
	}

	callee := node.Callee.(*parser.IdentNode)
	switch callee.Ident {
	case "def":
		return evalDef(node)
	}

	return nil
}

func evalDef(node *parser.CallNode) ast.Decl {
	if len(node.Args) < 2 {
		msg := fmt.Sprintf(MissingAssgnmentArgsError, node.Args[0])
		log.Critical(msg)
		panic(msg)
	}

	val := EvalExpr(node.Args[1])
	fn, ok := val.(*ast.FuncLit)

	ident := makeIdomaticIdent(node.Args[0].(*parser.IdentNode).Ident)

	if ok {
		if ident.Name == "main" {
			mainable(fn)
		}

		return makeFunDeclFromFuncLit(ident, fn)
	} else {
		return makeGeneralDecl(token.VAR, []ast.Spec{makeValueSpec([]*ast.Ident{ident}, []ast.Expr{val}, nil)})
	}
}

func isNSDecl(node parser.Node) bool {
	if node.Type() != parser.NodeCall {
		return false
	}

	call := node.(*parser.CallNode)
	if call.Callee.(*parser.IdentNode).Ident != "ns" {
		return false
	}

	if len(call.Args) < 1 {
		return false
	}

	return true
}

func getNamespace(node *parser.CallNode) (*ast.Ident, ast.Decl) {
	return getPackageName(node), getImports(node)
}

func getPackageName(node *parser.CallNode) *ast.Ident {
	if node.Args[0].Type() != parser.NodeIdent {
		log.Critical(NSPackageTypeMismatch)
		panic(NSPackageTypeMismatch)
	}

	return ast.NewIdent(node.Args[0].(*parser.IdentNode).Ident)
}

func checkNSArgs(node *parser.CallNode) bool {
	if node.Callee.Type() != parser.NodeIdent {
		return false
	}

	if callee := node.Callee.(*parser.IdentNode); callee.Ident != "ns" {
		return false
	}

	return true
}

func GenerateASTFromLispFile(filename string) (*token.FileSet, *ast.File) {
	b, err := ioutil.ReadFile(filename)

	// XXX let's improve the error handling here ...
	if err != nil {
		log.Critical(err)
		panic(err)
	}

	fset := token.NewFileSet()
	p := parser.ParseFromString(filename, string(b)+"\n")
	a := GenerateAST(p)

	return fset, a
}

func GenerateASTFromLispString(data string) (*token.FileSet, []ast.Expr) {
	fset := token.NewFileSet()
	p := parser.ParseFromString("<REPL>", data+"\n")
	a := EvalExprs(p)

	return fset, a
}

func PrintASTFromFile(filename string) {
	fset, a := GenerateASTFromLispFile(filename)
	ast.Print(fset, a)
	return
}

func WriteASTFromFile(fromFile string, toFile string) {
	var buf bytes.Buffer
	fset, a := GenerateASTFromLispFile(fromFile)
	ast.Fprint(&buf, fset, a, nil)
	err := ioutil.WriteFile(toFile, buf.Bytes(), 0644)

	// XXX let's improve the error handling here ...
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	return
}

func PrintASTFromLispString(data string) {
	fset, a := GenerateASTFromLispString(data)
	ast.Print(fset, a)
	return
}

func WriteGoFromFile(fromFile string, toFile string) {
	var buf bytes.Buffer
	fset, a := GenerateASTFromLispFile(fromFile)
	printer.Fprint(&buf, fset, a)
	err := ioutil.WriteFile(toFile, buf.Bytes(), 0644)

	// XXX let's improve the error handling here ...
	if err != nil {
		log.Critical(err)
		log.Debug("Tried writing to file:", toFile)
		panic(err)
	}
}

func PrintGoFromFile(filename string) {
	var buf bytes.Buffer
	fset, a := GenerateASTFromLispFile(filename)
	printer.Fprint(&buf, fset, a)
	fmt.Printf("%s\n", buf.String())
	return
}

func PrintGoFromLispString(data string) {
	var buf bytes.Buffer
	fset, a := GenerateASTFromLispString(data)
	printer.Fprint(&buf, fset, a)
	fmt.Printf("%s\n", buf.String())
	return
}
