package lexer

import (
	"testing"
)

func TestLexSumtype(t *testing.T) {
	input := `package orders

sumtype OrderState {
	Pending
	Paid { amount: float64 }
}`
	l := New("test.gox", input)
	expected := []TokenKind{
		TokPackage, TokIdent, TokNewline,
		TokSumtype, TokIdent, TokLBrace,
		TokIdent,
		TokIdent, TokLBrace, TokIdent, TokColon, TokIdent, TokRBrace,
		TokRBrace,
		TokEOF,
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Kind != want {
			t.Fatalf("token[%d]: got %s (%q), want %s", i, tok.Kind, tok.Value, want)
		}
	}
}

func TestLexContract(t *testing.T) {
	input := `contract CreateUser {
	input {
		email: string @required @email
	}
	output {
		id: string
	}
	errors {
		EmailTaken 409
	}
	route POST /api/users
}`
	l := New("test.gox", input)
	for {
		tok := l.NextToken()
		if tok.Kind == TokEOF {
			break
		}
	}
}

func TestLexArrow(t *testing.T) {
	input := `=>`
	l := New("test.gox", input)
	tok := l.NextToken()
	if tok.Kind != TokArrow {
		t.Fatalf("got %s, want =>", tok.Kind)
	}
}

func TestLexAnnotation(t *testing.T) {
	input := `@minlen(8)`
	l := New("test.gox", input)
	tok := l.NextToken()
	if tok.Kind != TokAt {
		t.Fatalf("got %s, want @", tok.Kind)
	}
	tok = l.NextToken()
	if tok.Kind != TokIdent || tok.Value != "minlen" {
		t.Fatalf("got %s %q, want IDENT minlen", tok.Kind, tok.Value)
	}
	tok = l.NextToken()
	if tok.Kind != TokLParen {
		t.Fatalf("got %s, want (", tok.Kind)
	}
	tok = l.NextToken()
	if tok.Kind != TokNumber || tok.Value != "8" {
		t.Fatalf("got %s %q, want NUMBER 8", tok.Kind, tok.Value)
	}
}
