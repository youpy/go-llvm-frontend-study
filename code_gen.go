package main

import (
	"github.com/axw/gollvm/llvm"
	"llvm_study/frontend"
	"os"
)

func main() {
	filename := os.Args[1]

	parser := frontend.NewParser(filename)
	parser.DoParse()
	ast := parser.GetAST()

	llvm.InitializeNativeTarget()

	codeGen := frontend.NewCodeGen(llvm.GlobalContext())
	codeGen.DoCodeGen(ast, filename)
	mod := codeGen.GetModule()

	pass := llvm.NewPassManager()
	defer pass.Dispose()

	pass.AddPromoteMemoryToRegisterPass()
	pass.Run(mod)

	mod.Dump()
}
