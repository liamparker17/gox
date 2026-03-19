package lexer

import "github.com/liamp/gox/ast"

type TokenKind int

const (
	TokEOF TokenKind = iota
	TokPackage
	TokImport
	TokSumtype
	TokContract
	TokMatch
	TokFunc
	TokInput
	TokOutput
	TokErrors
	TokRoute
	TokLBrace   // {
	TokRBrace   // }
	TokLParen   // (
	TokRParen   // )
	TokColon    // :
	TokArrow    // =>
	TokAt       // @
	TokComma    // ,
	TokIdent
	TokString
	TokNumber
	TokGoBlock
	TokComment
	TokNewline
)

var tokenNames = map[TokenKind]string{
	TokEOF: "EOF", TokPackage: "package", TokImport: "import",
	TokSumtype: "sumtype", TokContract: "contract", TokMatch: "match",
	TokFunc: "func", TokInput: "input", TokOutput: "output",
	TokErrors: "errors", TokRoute: "route",
	TokLBrace: "{", TokRBrace: "}", TokLParen: "(", TokRParen: ")",
	TokColon: ":", TokArrow: "=>", TokAt: "@", TokComma: ",",
	TokIdent: "IDENT", TokString: "STRING", TokNumber: "NUMBER",
	TokGoBlock: "GOBLOCK", TokComment: "COMMENT", TokNewline: "NEWLINE",
}

func (k TokenKind) String() string {
	if s, ok := tokenNames[k]; ok {
		return s
	}
	return "UNKNOWN"
}

var keywords = map[string]TokenKind{
	"package":  TokPackage,
	"import":   TokImport,
	"sumtype":  TokSumtype,
	"contract": TokContract,
	"match":    TokMatch,
	"func":     TokFunc,
	"input":    TokInput,
	"output":   TokOutput,
	"errors":   TokErrors,
	"route":    TokRoute,
}

type Token struct {
	Kind  TokenKind
	Value string
	Pos   ast.Position
}
