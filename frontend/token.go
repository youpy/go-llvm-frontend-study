package frontend

import (
	"strconv"
)

type TokenType int

const (
	TOK_IDENTIFIER TokenType = 0
	TOK_DIGIT      TokenType = 1
	TOK_SYMBOL     TokenType = 2
	TOK_INT        TokenType = 3
	TOK_RETURN     TokenType = 4
	TOK_EOF        TokenType = 5
)

type Token struct {
	Type        TokenType
	TokenString string
	Number      int
	Line        int
}

func NewToken(str string, tokenType TokenType, line int) *Token {
	var number int

	if tokenType == TOK_DIGIT {
		number, _ = strconv.Atoi(str)
	} else {
		number = 0x7fffffff
	}

	return &Token{
		Type:        tokenType,
		TokenString: str,
		Number:      number,
		Line:        line}
}
