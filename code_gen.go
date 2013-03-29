package main

import (
	"github.com/axw/gollvm/llvm"
	"llvm_study/frontend"
	"os"
)

func main() {
	infile := os.Args[1]
	outfile := os.Args[2]

	parser := frontend.NewParser(infile)
	parser.DoParse()
	ast := parser.GetAST()

	llvm.InitializeNativeTarget()

	codeGen := frontend.NewCodeGen(llvm.GlobalContext())
	codeGen.DoCodeGen(ast, infile)
	mod := codeGen.GetModule()

	pass := llvm.NewPassManager()
	defer pass.Dispose()

	pass.AddPromoteMemoryToRegisterPass()
	pass.Run(mod)

	err := mod.PrintToFile(outfile)

	if err != nil {
		panic(err)
	}
}
