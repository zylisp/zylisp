package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/zylisp/gisp/generator"
	"github.com/zylisp/gisp/parser"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

func AstMain() {
	banner := Banner {
		commonHelp: CommonReplHelp,
		greeting: ReplBannerGreeting,
		modeHelp: AstReplHelp,
		replMode: "AST",
	}

	banner.printBanner()
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(AstPrompt)
		line, _, _ := r.ReadLine()
		// XXX we should explore REPL-based packages ... that would allow for a
		//     more Go-like experience in the REPL, with the ability to declare a
		//     new package in the REPL, and then refer to work done in the same
		//     session, but in a different REPL package ... I guess this applies
		//     more to the Lisp REPL
		p := parser.ParseFromString("<REPL>", string(line)+"\n")
		fmt.Printf("Parsed:\n%s\n", p)

		// a := generator.GenerateAST(p)
		a := generator.EvalExprs(p)
		fset := token.NewFileSet()
		fmt.Println("AST:")
		ast.Print(fset, a)

		var buf bytes.Buffer
		printer.Fprint(&buf, fset, a)
		fmt.Printf("%s\n", buf.String())
	}
}

// func GoGenMain() {
// 	banner := Banner {
// 		commonHelp: CommonReplHelp,
// 		greeting: ReplBannerGreeting,
// 		modeHelp: GoGenReplHelp,
// 		replMode: "Go-gen",
// 	}
// 	banner.printBanner()
// }

// func LispMain() {
// 	banner := Banner {
// 		commonHelp: CommonReplHelp,
// 		greeting: ReplBannerGreeting,
// 		modeHelp: LispReplHelp,
// 		replMode: "Lisp",
// 	}
// 	banner.printBanner()
// }