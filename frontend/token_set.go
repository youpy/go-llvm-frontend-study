package frontend

import (
	"fmt"
)

type TokenSet struct {
	Tokens   []*Token
	CurIndex int
}

func NewTokenSet() *TokenSet {
	var tokens []*Token

	return &TokenSet{Tokens: tokens, CurIndex: 0}
}

func (t *TokenSet) pushToken(token *Token) bool {
	t.Tokens = append(t.Tokens, token)

	return true
}

func (t *TokenSet) getCurIndex() int {
	return t.CurIndex
}

func (t *TokenSet) getCurType() TokenType {
	return t.Tokens[t.CurIndex].Type
}

func (t *TokenSet) getCurString() string {
	return t.Tokens[t.CurIndex].TokenString
}

func (t *TokenSet) getCurNumVal() int {
	return t.Tokens[t.CurIndex].Number
}

func (t *TokenSet) getToken() Token {
	return *t.Tokens[t.CurIndex]
}

func (t *TokenSet) getNextToken() (result bool) {
	size := len(t.Tokens)

	if size-1 <= t.CurIndex {
		result = false
	} else {
		t.CurIndex++

		result = true
	}

	return
}

func (t *TokenSet) applyTokenIndex(index int) bool {
	t.CurIndex = index

	return true
}

func (t *TokenSet) ungetToken(times int) bool {
	for i := 0; i < times; i++ {
		if t.CurIndex == 0 {
			return false
		} else {
			t.CurIndex--
		}
	}

	return true
}

func (t *TokenSet) PrintTokens() bool {
	for _, token := range t.Tokens {
		tokenType := token.Type
		fmt.Printf("%d:", tokenType)

		if tokenType != TOK_EOF {
			fmt.Printf("%s (%d)\n", token.TokenString, token.Line)
		}
	}

	return true
}
