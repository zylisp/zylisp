/*
The ZYLISP command line and multi-REPL tool.

Overview

The ZYLISP project's zylisp command is a tool for both performing command line
actions (i.e., batch jobs) as well as interactive programming (REPLs). In both
cases, there are three modes:

   * AST
   * Go code generation
   * Byte-code compilation (no REPL; just CLI support)
   * Lisp (the actual, classic REPL)

Each of these is covered in more detail below.

Setup

For setting the project GOPATH and building the zylisp command, see the project
Github page at https://github.com/zylisp/zylisp#development.

If you would prefer to use ZYLISP without cloning the repo and setting up a
development environment, you may install it with the following:

	$ go get github.com/zylisp/gisp/cmd/zylisp

Since the zylisp command makes use of Go flags, it has a generated help output.
You may view this with the usual -h option:

	$ zylisp -h

Logging

The `zylisp` executable supports passing a `-loglevel` option with one of the
following a legal associated value:

  * debug
  * info
  * notice
  * warning
  * error
  * failure


REPL

The ZYLISP REPL is designed for interactive programming.

In AST mode, the REPL will take any expression given and attempt to create
a Go abstract syntax tree from that:

	$ zylisp -ast

	AST> (+ 1 2)
	Parsed:
	[(+ 1 2)]
	AST:
	     0  []ast.Expr (len = 1) {
	     1  .  0: *ast.CallExpr {
	     2  .  .  Fun: *ast.SelectorExpr {
	     3  .  .  .  X: *ast.Ident {
	     4  .  .  .  .  NamePos: -
	     5  .  .  .  .  Name: "core"
	     6  .  .  .  }
	     7  .  .  .  Sel: *ast.Ident {
	     8  .  .  .  .  NamePos: -
	     9  .  .  .  .  Name: "ADD"
	    10  .  .  .  }
	    11  .  .  }
	    12  .  .  Lparen: -
	    13  .  .  Args: []ast.Expr (len = 2) {
	    14  .  .  .  0: *ast.Ident {
	    15  .  .  .  .  NamePos: -
	    16  .  .  .  .  Name: "1"
	    17  .  .  .  }
	    18  .  .  .  1: *ast.Ident {
	    19  .  .  .  .  NamePos: -
	    20  .  .  .  .  Name: "2"
	    21  .  .  .  }
	    22  .  .  }
	    23  .  .  Ellipsis: -
	    24  .  .  Rparen: -
	    25  .  }
	    26  }


CLI

The command line interface is for performing batched or scripted operations
and is enabled with the -cli option. However, in CLI mode, you must pass one
or more file names (file globbing allowed).

In AST mode, the CLI will parse the given file and produce as output the
corresponding Go AST for the Lisp code in the file provided:

	$ zylisp -ast -cli examples/factorial.gsp

The output is a bit long (302 lines), but here are the first 20 lines:

     0  *ast.File {
     1  .  Package: -
     2  .  Name: *ast.Ident {
     3  .  .  NamePos: -
     4  .  .  Name: "main"
     5  .  }
     6  .  Decls: []ast.Decl (len = 3) {
     7  .  .  0: *ast.GenDecl {
     8  .  .  .  TokPos: -
     9  .  .  .  Tok: import
    10  .  .  .  Lparen: -
    11  .  .  .  Specs: []ast.Spec (len = 2) {
    12  .  .  .  .  0: *ast.ImportSpec {
    13  .  .  .  .  .  Path: *ast.BasicLit {
    14  .  .  .  .  .  .  ValuePos: -
    15  .  .  .  .  .  .  Kind: STRING
    16  .  .  .  .  .  .  Value: "\"fmt\""
    17  .  .  .  .  .  }
    18  .  .  .  .  .  EndPos: -
    19  .  .  .  .  }
    20  .  .  .  .  1: *ast.ImportSpec {

In Go-generation mode, the output is instead compilable Go code:

	$ zylisp -cli -go examples/factorial.gsp

	package main

	import (
		"fmt"
		"github.com/zylisp/gisp/core"
	)

	func main() {
		fmt.Printf("10! = %d\n", int(factorial(10).(float64)))
	}
	func factorial(n core.Any) core.Any {
		return func() core.Any {
			if core.LT(n, 2) {
				return 1
			} else {
				return core.MUL(n, factorial(core.ADD(n, -1)))
			}
		}()
	}

In byte-code compilation mode, Go is generated under the hood, and then it is
compiled to byte code using `go build`:

  $ zylisp -cli -bytecode -dir bin/examples examples/*.gsp

This also demonstrate support for file globbing, allowing you to generate
output for multiple files at once.

*/
package main

import (
  "flag"
  "fmt"
  "github.com/op/go-logging"
  "github.com/zylisp/gisp"
  "github.com/zylisp/gisp/generator"
  "github.com/zylisp/gisp/repl"
  "os"
  "os/exec"
  "path/filepath"
)

type Modes struct {
	cli bool
	ast bool
	gogen bool
  bytecode bool
	lisp bool
}

type Inputs struct {
	files []string
	hasFiles bool
}

type Outputs struct {
	dir string
	files []string
	isDir bool
	isFile bool
	useDir bool
	useFile bool
}

func RemoveExtension(filename string) string {
	extension := filepath.Ext(filename)
	return filename[0:len(filename)-len(extension)]
}

func PrepareOutputDir(dir string) {
	log := logging.MustGetLogger(gisp.ApplicationName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Info("Directory '%s' does not exist; creating ...", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
}

func PrepareOutputFile(filename string) {
  log := logging.MustGetLogger(gisp.ApplicationName)
  log.Debug("Preparing output file:", filename)
  basename := filepath.Dir(filename)
  log.Debug("Got basename:", basename)
  PrepareOutputDir(basename)
}

func MakeOutputFilename(prefix string, inputFile string, extension string) string {
	var template string
  if extension == "" {
    template = "%s%s%s%s"
  } else {
    template = "%s%s%s.%s"
  }
  return fmt.Sprintf(
		template,
		prefix,
		string(os.PathSeparator),
		filepath.Base(RemoveExtension(inputFile)),
		extension)
}

func isDir(filename string) bool {
  log := logging.MustGetLogger(gisp.ApplicationName)
  file, err := os.Stat(filename)
  if err != nil {
      log.Debugf(gisp.DirectoryError, filename, err.Error())
      return false
  }
  if file.Mode().IsDir() {
    return true
  } else {
    return false
  }
}

func compileGo(infile string, outfile string) {
  log := logging.MustGetLogger(gisp.ApplicationName)
  log.Noticef("Compiling %s ...", outfile)
  cmd := exec.Command("go", "build", "-o", outfile, infile)
  _, err := cmd.Output()
  if err != nil {
    log.Errorf(repl.CompileError, err.Error())
  }
}

func dispatchLisp(modes Modes) {
	if modes.cli {
		// LISP CLI
		fmt.Println(repl.LispCLIUnsupportedError)
	} else {
		// LISP REPL
		fmt.Println(repl.LispREPLUnsupportedError)
		// repl.LispMain()
	}
}

func dispatchAST(modes Modes, inputs Inputs, outputs Outputs) {
	log := logging.MustGetLogger(gisp.ApplicationName)
	if modes.cli {
		// AST CLI
		for i, inputFile := range inputs.files {
			log.Infof("Processing file '%s' for AST output '%s' ...",
				inputFile, outputs.files[i])
      log.Debug("Use file for output?", outputs.useFile)
      log.Debug("Use directory for output?", outputs.useDir)
			if outputs.useFile {
				generator.WriteASTFromFile(inputFile, outputs.files[i])
			} else {
				generator.PrintASTFromFile(inputFile)
			}
		}
	} else {
		// AST REPL
		repl.ASTMain()
	}
}

func dispatchGoGen(modes Modes, inputs Inputs, outputs Outputs) {
	log := logging.MustGetLogger(gisp.ApplicationName)
	if modes.cli {
		// Go-generator CLI
		for i, inputFile := range inputs.files {
			log.Infof("Processing file '%s' for Go output '%s' ...",
				inputFile, outputs.files[i])
			if outputs.useFile {
				generator.WriteGoFromFile(inputFile, outputs.files[i])
			} else {
				generator.PrintGoFromFile(inputFile)
			}
		}
	} else {
		// GOGEN REPL
		repl.GoGenMain()
	}
}

func dispatchByteCode(modes Modes, inputs Inputs, outputs Outputs) {
  log := logging.MustGetLogger(gisp.ApplicationName)
  if modes.cli {
    // Go-compiler CLI
    for i, inputFile := range inputs.files {
      outputFile := outputs.files[i]
      goOutputFile := outputFile + ".go"
      log.Infof("Processing file '%s' for Go output '%s' ...",
        inputFile, goOutputFile)
      if outputs.useFile {
        generator.WriteGoFromFile(inputFile, goOutputFile)
      } else {
        log.Error(repl.CompileWithoutFileError)
      }
      log.Infof("Processing file '%s' for byte-code output '%s' ...",
        goOutputFile, outputFile)
      compileGo(goOutputFile, outputFile)
    }
  } else {
    log.Error(repl.CompoileWithoutCLIError)
  }
}

func getUseDir (dir bool) bool {
	if dir {
		return true
	} else {
		return false
	}
}

func getHasFiles (files []string) bool {
	if len(files) > 0 {
		return true
	} else {
		return false
	}

}

func extensionFromMode(modes Modes) string {
	var extension string
	if modes.lisp {
		extension = "zsp"
	} else if modes.ast {
		extension = "ast"
	} else if modes.gogen {
		extension = "go"
  } else if modes.bytecode {
    extension = ""
	} else {
		fmt.Println(repl.ModeNeededError)
		os.Exit(1)
	}
	return extension
}

func dispatch(modes Modes, inputs Inputs, outputs Outputs) {
	log := logging.MustGetLogger(gisp.ApplicationName)
	log.Debug("Dispatched")
	log.Debug("Got modes:", modes)
	log.Debug("Got inputs:", inputs)
	log.Debug("Got outputs:", outputs)
	if modes.lisp {
		dispatchLisp(modes)
	} else if modes.ast {
		dispatchAST(modes, inputs, outputs)
	} else if modes.gogen {
		dispatchGoGen(modes, inputs, outputs)
  } else if modes.bytecode {
    dispatchByteCode(modes, inputs, outputs)
	} else {
		fmt.Println(repl.ModeNeededError)
		os.Exit(1)
	}
}

func setupLogging (stringLogLevel string) {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile}, %{shortfunc} ▶ {level:.4s} %{id:03x}%{color:reset} %{message}`,)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevel, err := logging.LogLevel(stringLogLevel)
	if err != nil {
		panic(gisp.LogLevelUnsupportedError)
	}
	backendLeveled.SetLevel(logLevel, "")
	logging.SetBackend(backendLeveled)
	log := logging.MustGetLogger(gisp.ApplicationName)
	log.Info("Set up logging")
}

func main() {
	astPtr := flag.Bool("ast", false, "Enable AST mode")
	cliPtr := flag.Bool("cli", false, "Run as a CLI tool")
	dirPtr := flag.String("dir", "", "Default directory for writing operations")
	goPtr := flag.Bool("go", false, "Enable Go code-generation mode")
  byteCodePtr := flag.Bool("bytecode", false, "Enable byte-code compilation from generated Go")
	lispPtr := flag.Bool("lisp", false, "Enable LISP mode")
	logLevelPtr := flag.String("loglevel", "warning", "Set the logging level")
	outPtr := flag.String("o", "", "Default filename for writing operations")

	flag.Parse()


	log := logging.MustGetLogger(gisp.ApplicationName)
	setupLogging(*logLevelPtr)

	inputFiles := flag.Args()
	hasFiles := getHasFiles(inputFiles)
	isDir := len(*dirPtr) > 0

	modes := Modes {
		cli: *cliPtr,
		ast: *astPtr,
		gogen: *goPtr,
    bytecode: *byteCodePtr,
		lisp: *lispPtr,
	}

	inputs := Inputs {
		files: inputFiles,
		hasFiles: hasFiles,
	}

	outputs := Outputs {
		dir: *dirPtr,
		isDir: isDir,
		isFile: len(*outPtr) > 0,
		useDir: isDir,
		useFile: false,
	}

	if modes.cli {
		// Check for at least one file to operate upon, when in CLI mode
		if inputs.hasFiles {
			if len(inputs.files) > 1 {
				log.Debug("Got multiple input files")
				// If more than one file is given, ignore output file and only use dir
				if outputs.isDir {
					log.Debug("Output dir is defined: using it ...")
					outputs.files = []string{}
					// Since we're going to be using the dir, make sure it exists/create
					// if necessary
					PrepareOutputDir(outputs.dir)
					log.Debug("Input files:", inputs.files)
					for _, file := range inputs.files {
						outputs.files = append(outputs.files, MakeOutputFilename(
							outputs.dir, file, extensionFromMode(modes)))
					}
					outputs.useFile = true
				} else {
					log.Error(repl.DirNeededError)
					os.Exit(1)
				}
			} else {
				log.Debug("Got a single input file")
				log.Debug("Original output files:", outputs.files)
				log.Debug("Original input files:", inputs.files)
				// if only one file is given and dir is given, then set the output file to
				// be the dir/infile.updated-extension
				if outputs.isDir {
					log.Debug("Outputs is a directory, using it ...")
					PrepareOutputDir(outputs.dir)
					outputs.files = append(outputs.files, MakeOutputFilename(
						outputs.dir, inputs.files[0], extensionFromMode(modes)))
					outputs.useFile = true
					log.Debug("Modified output files:", outputs.files)
					log.Debug("Modified input files:", inputs.files)
				// If only one file is given, and the output file is set
				} else {
					log.Debug("Outputs is not a directory.")
          PrepareOutputFile(*outPtr)
					outputs.files = append(outputs.files, *outPtr)
          outputs.useFile = true
					log.Debug("Modified output files:", outputs.files)
					log.Debug("Modified input files:", inputs.files)
				}
			}
		} else {
			log.Error(repl.FilesNeededError)
			os.Exit(1)
		}
	}
	log.Debug("Preparing to dispatch ...")
	dispatch(modes, inputs, outputs)
}
