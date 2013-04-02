package frontend

import (
	"github.com/coocood/assrt"
	"testing"
)

func TestLexicalAnalysis(t *testing.T) {
	filename := "../test.xxx"
	assert := assrt.NewAssert(t)

	tokens := LexicalAnalysis(filename)

	assert.Equal(52, len(tokens.Tokens))
	assert.Equal("int", tokens.Tokens[0].TokenString)
	assert.Equal(TOK_INT, tokens.Tokens[0].Type)
	assert.Equal(10, tokens.Tokens[17].Number)
	assert.Equal(TOK_DIGIT, tokens.Tokens[17].Type)
	assert.Equal("}", tokens.Tokens[50].TokenString)
	assert.Equal(TOK_SYMBOL, tokens.Tokens[50].Type)
}
