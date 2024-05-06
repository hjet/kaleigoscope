package main

import (
	"bytes"
	"strings"
	"testing"
)

func lex(src string) (tokens string, err error) {
	sc := &Scanner{}
	sc.Init([]byte(src))

	var (
		buf bytes.Buffer
		lit string
		tok Token
	)

	for {
		tok, lit = sc.GetTok()

		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}

		switch tok {
		case TokEOF:
			buf.WriteString("EOF")
		case TokDef:
			buf.WriteString("def")
		case TokExtern:
			buf.WriteString("extern")
		case TokIdentifier:
			buf.WriteString(lit)
		case TokNumber:
			buf.WriteString(lit)
		case TokOperator:
			buf.WriteString(lit)
		}

		if tok == TokEOF {
			break
		}

	}

	return buf.String(), nil

}

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		input, want string
	}{
		{"", "EOF"},
		{"123", "123 EOF"},
		{"1.4", "1.4 EOF"},
		{"x = 3", "x = 3 EOF"},
		{"def foo(x) = x + 3", "def foo ( x ) = x + 3 EOF"},
		{"extern sin()", "extern sin ( ) EOF"},
		{"def foo(x) # comment", "def foo ( x ) EOF"},
	} {
		got, err := lex(test.input)
		if err != nil {
			t.Errorf(err.Error())
		}

		if !strings.HasPrefix(got, test.want) {
			t.Errorf("scan `%s` = [%s], want [%s]", test.input, got, test.want)
		}

	}
}
