package main

import (
	"encoding/json"
	"llvm_study/frontend"
	"os"
)

func main() {
	filename := os.Args[1]

	parser := frontend.NewParser(filename)
	parser.DoParse()
	ast := parser.GetAST()

	json.NewEncoder(os.Stdout).Encode(ast)
}
