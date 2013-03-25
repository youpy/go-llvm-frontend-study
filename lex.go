package main

import (
	"llvm_study/frontend"
	"os"
)

func main() {
	filename := os.Args[1]

	tokens := frontend.LexicalAnalysis(filename)

	if tokens != nil {
		tokens.PrintTokens()
	}
}
