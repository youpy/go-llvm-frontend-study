package main

import (
	"fmt"
	"llvm_study/frontend"
	"os"
)

func main() {
	filename := os.Args[1]

	parser := frontend.NewParser(filename)
	parser.DoParse()

	fmt.Println(parser.GetAST().Functions[0].Proto.Name)
}
