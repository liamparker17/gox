package lexer

import (
	"strings"
	"unicode"

	"github.com/liamparker17/gox/ast"
)

type Lexer struct {
	src        string
	filename   string
	pos        int
	line       int
	col        int
	braceDepth int // tracks {} nesting; newlines inside braces are suppressed
}

func New(filename, src string) *Lexer {
	return &Lexer{src: src, filename: filename, pos: 0, line: 1, col: 1}
}

func (l *Lexer) position() ast.Position {
	return ast.Position{File: l.filename, Line: l.line, Column: l.col}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) advance() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// skipWhitespaceExceptNewline skips spaces, tabs, and carriage returns on the
// current line only — newlines are significant at the top level.
func (l *Lexer) skipWhitespaceExceptNewline() {
	for l.pos < len(l.src) && l.src[l.pos] != '\n' && (l.src[l.pos] == ' ' || l.src[l.pos] == '\t' || l.src[l.pos] == '\r') {
		l.advance()
	}
}

// consumeNewlines consumes one or more newline characters (and surrounding
// horizontal whitespace) and returns true if at least one was consumed.
func (l *Lexer) consumeNewlines() bool {
	consumed := false
	for l.pos < len(l.src) {
		// Skip horizontal whitespace.
		for l.pos < len(l.src) && l.src[l.pos] != '\n' && (l.src[l.pos] == ' ' || l.src[l.pos] == '\t' || l.src[l.pos] == '\r') {
			l.advance()
		}
		if l.pos < len(l.src) && l.src[l.pos] == '\n' {
			l.advance()
			consumed = true
		} else {
			break
		}
	}
	return consumed
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespaceExceptNewline()

	if l.pos >= len(l.src) {
		return Token{Kind: TokEOF, Pos: l.position()}
	}

	pos := l.position()
	ch := l.peek()

	// Handle newlines: only emit TokNewline at the top level (braceDepth == 0).
	// Inside braces, newlines are insignificant whitespace.
	if ch == '\n' {
		if l.braceDepth > 0 {
			// Inside braces: consume all newlines and try again.
			l.consumeNewlines()
			return l.NextToken()
		}
		// Top level: collapse all consecutive newlines into one token.
		l.consumeNewlines()
		return Token{Kind: TokNewline, Value: "\n", Pos: pos}
	}

	if ch == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
		start := l.pos
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.advance()
		}
		return Token{Kind: TokComment, Value: l.src[start:l.pos], Pos: pos}
	}

	if ch == '"' {
		return l.lexString(pos)
	}

	if ch >= '0' && ch <= '9' {
		return l.lexNumber(pos)
	}

	switch ch {
	case '{':
		l.advance()
		l.braceDepth++
		return Token{Kind: TokLBrace, Value: "{", Pos: pos}
	case '}':
		l.advance()
		if l.braceDepth > 0 {
			l.braceDepth--
		}
		return Token{Kind: TokRBrace, Value: "}", Pos: pos}
	case '(':
		l.advance()
		return Token{Kind: TokLParen, Value: "(", Pos: pos}
	case ')':
		l.advance()
		return Token{Kind: TokRParen, Value: ")", Pos: pos}
	case ':':
		l.advance()
		return Token{Kind: TokColon, Value: ":", Pos: pos}
	case '@':
		l.advance()
		return Token{Kind: TokAt, Value: "@", Pos: pos}
	case ',':
		l.advance()
		return Token{Kind: TokComma, Value: ",", Pos: pos}
	case '=':
		if l.pos+1 < len(l.src) && l.src[l.pos+1] == '>' {
			l.advance()
			l.advance()
			return Token{Kind: TokArrow, Value: "=>", Pos: pos}
		}
	}

	if ch == '_' || unicode.IsLetter(rune(ch)) {
		return l.lexIdent(pos)
	}

	if ch == '/' {
		return l.lexPath(pos)
	}

	// Unknown character — skip and try again.
	l.advance()
	return l.NextToken()
}

func (l *Lexer) lexString(pos ast.Position) Token {
	l.advance() // consume opening quote
	var b strings.Builder
	for l.pos < len(l.src) && l.src[l.pos] != '"' {
		if l.src[l.pos] == '\\' && l.pos+1 < len(l.src) {
			l.advance() // consume backslash
		}
		b.WriteByte(l.advance())
	}
	if l.pos < len(l.src) {
		l.advance() // consume closing quote
	}
	return Token{Kind: TokString, Value: b.String(), Pos: pos}
}

func (l *Lexer) lexNumber(pos ast.Position) Token {
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9' || l.src[l.pos] == '.') {
		l.advance()
	}
	return Token{Kind: TokNumber, Value: l.src[start:l.pos], Pos: pos}
}

func (l *Lexer) lexIdent(pos ast.Position) Token {
	start := l.pos
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if ch == '_' || ch == '.' || unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) {
			l.advance()
		} else {
			break
		}
	}
	value := l.src[start:l.pos]
	kind := TokIdent
	if kw, ok := keywords[value]; ok {
		kind = kw
	}
	return Token{Kind: kind, Value: value, Pos: pos}
}

func (l *Lexer) lexPath(pos ast.Position) Token {
	start := l.pos
	for l.pos < len(l.src) && !unicode.IsSpace(rune(l.src[l.pos])) {
		l.advance()
	}
	return Token{Kind: TokString, Value: l.src[start:l.pos], Pos: pos}
}
