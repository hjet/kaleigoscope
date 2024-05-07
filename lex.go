package main

import (
	"unicode"
)

type Token int

const (
	TokEOF        Token = -1
	TokDef        Token = -2
	TokExtern     Token = -3
	TokIdentifier Token = -4
	TokNumber     Token = -5
	TokOperator   Token = -6
)

const EOF = 0

type Scanner struct {
	src      []byte
	pos      int
	lastChar rune
}

func newScanner(src []byte) *Scanner {
	s := &Scanner{
		src: src,
		pos: 0,
	}
	if len(src) > 0 {
		s.lastChar = rune(src[0])
	} else {
		s.lastChar = EOF
	}
	return s
}

func (s *Scanner) error(e string) {
	panic(e)
}

func (s *Scanner) next() {
	if s.lastChar == EOF {
		return
	}

	s.pos += 1
	if s.pos < len(s.src) {
		s.lastChar = rune(s.src[s.pos])
	} else {
		s.lastChar = EOF
	}
}

func (s *Scanner) GetTok() (tok Token, lit string) {
	// skip whitespace
	for s.lastChar == ' ' || s.lastChar == '\t' || s.lastChar == '\n' || s.lastChar == '\r' {
		s.next()
	}

	// identifiers or keywords
	if unicode.IsLetter(s.lastChar) {
		st := s.pos

		for isAlphaNum(s.lastChar) {
			s.next()
		}

		ident := string(s.src[st:s.pos])

		if ident == "def" {
			return TokDef, ""
		}

		if ident == "extern" {
			return TokExtern, ""
		}

		return TokIdentifier, ident
	}

	// numeric values
	if unicode.IsDigit(s.lastChar) || s.lastChar == '.' {
		st := s.pos
		for unicode.IsDigit(s.lastChar) || s.lastChar == '.' {
			s.next()
		}

		numStr := string(s.src[st:s.pos])
		return TokNumber, numStr
	}

	// comments
	if s.lastChar == '#' {
		s.next()
		for s.lastChar != EOF && s.lastChar != '\n' && s.lastChar != '\r' {
			s.next()
		}

		if s.lastChar != EOF {
			return s.GetTok()
		}
	}

	if s.lastChar == EOF {
		return TokEOF, ""
	}

	// return ASCII value of random char (e.g. "+")
	op := string(s.src[s.pos])
	s.next()
	return TokOperator, op

}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
