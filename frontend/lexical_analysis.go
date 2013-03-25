package frontend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func compare(char byte, str string) bool {
	return bytes.IndexAny([]byte{char}, str) == 0
}

func byte2string(char byte) string {
	return string([]byte{char})
}

func LexicalAnalysis(filename string) *TokenSet {
	var next_token *Token

	tokens := NewTokenSet()
	line_num := 0
	isComment := false

	file, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	defer func() {
		if file.Close() != nil {
			panic(err)
		}
	}()

	r := bufio.NewReader(file)

	for {
		buf, _, err := r.ReadLine()

		if err != nil && err != io.EOF {
			panic(err)
		}
		if buf == nil {
			break
		}

		index := 0
		length := len(buf)
		token_str := ""

		for index < length {
			next_char := byte2string(buf[index])
			index++

			if isComment {
				if (length-index) < 2 || !compare(buf[index], "*") || !compare(buf[index+1], "/") {
					continue
				} else {
					index += 2
					isComment = false
				}
			}

			if err == io.EOF {
				token_str := "EOF"
				next_token = NewToken(token_str, TOK_EOF, line_num)
			} else if strings.Contains(" \t\r\n", next_char) {
				continue
			} else if strings.Contains("abcdefghijklnmopqrstuvwxyz", strings.ToLower(next_char)) {
				token_str += next_char
				next_char = byte2string(buf[index])
				index++

				for strings.Contains("abcdefghijklnmopqrstuvwxyz0987654321", strings.ToLower(next_char)) {
					token_str += next_char
					next_char = byte2string(buf[index])
					index++

					if index == length {
						break
					}
				}

				index--

				if token_str == "int" {
					next_token = NewToken(token_str, TOK_INT, line_num)
				} else if token_str == "return" {
					next_token = NewToken(token_str, TOK_RETURN, line_num)
				} else {
					next_token = NewToken(token_str, TOK_IDENTIFIER, line_num)
				}
			} else if strings.Contains("0987654321", next_char) {
				if next_char == "0" {
					token_str += next_char
					next_token = NewToken(token_str, TOK_DIGIT, line_num)
				} else {
					token_str += next_char
					next_char = byte2string(buf[index])
					index++

					for strings.Contains("0987654321", next_char) {
						token_str += next_char
						next_char = byte2string(buf[index])
						index++
					}

					next_token = NewToken(token_str, TOK_DIGIT, line_num)
					index--
				}
			} else if next_char == "/" {
				token_str += next_char
				next_char = byte2string(buf[index])
				index++

				if next_char == "/" {
					break
				} else if next_char == "*" {
					isComment = true
					continue
				} else {
					index--
					next_token = NewToken(token_str, TOK_SYMBOL, line_num)
				}
			} else if strings.Contains("*+-=;,(){}", next_char) {
				token_str += next_char
				next_token = NewToken(token_str, TOK_SYMBOL, line_num)
			} else {
				fmt.Fprintf(os.Stderr, "unclear token : %s\n", next_char)
				return nil
			}

			tokens.Tokens = append(tokens.Tokens, next_token)
			token_str = ""
		}

		line_num++
	}

	return tokens
}
