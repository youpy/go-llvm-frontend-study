// based on http://cuddle.googlecode.com/hg/talk/lex.html
package frontend

import (
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

type Lexer struct {
	input   string
	start   int
	pos     int
	width   int
	lineNum int
	tokens  *TokenSet
}

type StateFn func(*Lexer) StateFn

const (
	leftComment  string = "/*"
	rightComment string = "*/"
	idInt        string = "int"
	idReturn     string = "return"
	eof          rune   = rune(0)
)

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, tokens: NewTokenSet()}
}

func (l *Lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width

	return r
}

func (l *Lexer) emit(t TokenType) {
	l.tokens.Tokens = append(l.tokens.Tokens, NewToken(l.input[l.start:l.pos], t, l.lineNum))
	l.start = l.pos
}

func (l *Lexer) accept(str string) bool {
	if strings.IndexRune(str, l.next()) >= 0 {
		return true
	}

	l.backup()
	return false
}

func (l *Lexer) acceptRun(str string) {
	for strings.IndexRune(str, l.next()) >= 0 {
	}
	l.backup()
}

func (l *Lexer) acceptPrefix(prefix string) bool {
	if strings.HasPrefix(l.input[l.pos:], prefix) {
		l.pos += len(prefix)
		return true
	}

	return false
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) run() {
	for state := lexCode; state != nil; {
		state = state(l)
	}
}

func lexCode(l *Lexer) StateFn {
	for {
		if l.acceptPrefix(leftComment) {
			return lexComment
		}

		if l.acceptPrefix(idInt) {
			l.emit(TOK_INT)
		} else if l.acceptPrefix(idReturn) {
			l.emit(TOK_RETURN)
		} else if l.accept("abcdefghijklnmopqrstuvwxyz") {
			l.acceptRun("abcdefghijklnmopqrstuvwxyz0123456789")
			l.emit(TOK_IDENTIFIER)
		} else if l.accept("\n") {
			l.lineNum += 1
			l.ignore()
		} else if l.accept("0123456789") {
			l.acceptRun("0123456789")
			l.emit(TOK_DIGIT)
		} else if l.accept("*+-=;,(){}") {
			l.emit(TOK_SYMBOL)
		} else {
			l.next()
			l.ignore()
		}

		if l.next() == eof {
			l.emit(TOK_EOF)
			break
		} else {
			l.backup()
		}
	}

	return nil
}

func lexComment(l *Lexer) StateFn {
	for {
		if l.acceptPrefix(rightComment) {
			l.ignore()
			return lexCode
		}

		if l.next() == eof {
			break
		}
	}

	return nil
}

func LexicalAnalysis(filename string) *TokenSet {
	input, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	lexer := NewLexer(string(input))
	lexer.run()

	return lexer.tokens
}
