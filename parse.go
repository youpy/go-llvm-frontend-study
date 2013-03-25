package main

import (
	"encoding/json"
	"fmt"
	"llvm_study/frontend"
	"os"
)

func main() {
	filename := os.Args[1]

	parser := frontend.NewParser(filename)
	parser.DoParse()
	ast := parser.GetAST()

	fmt.Println(json.NewEncoder(os.Stdout).Encode(ast))
}
