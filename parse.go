package main

// TODO: better consume / operator checking

import (
	"fmt"
	//"tinygo.org/x/go-llvm"
)

// AST
type (
	ExprAST interface {
		exprNode()
		//codeGen(*CodeGen) llvm.Value
	}

	NumberExprAST struct {
		Val string
	}

	VariableExprAST struct {
		Name string
	}

	BinaryExprAST struct {
		Op       rune
		LHS, RHS ExprAST
	}

	CallExprAST struct {
		Callee string
		Args   []ExprAST
	}

	PrototypeAST struct {
		Name string
		Args []string
	}

	FunctionAST struct {
		Proto *PrototypeAST
		Body  ExprAST
	}
)

func (*NumberExprAST) exprNode()   {}
func (*VariableExprAST) exprNode() {}
func (*BinaryExprAST) exprNode()   {}
func (*CallExprAST) exprNode()     {}
func (*PrototypeAST) exprNode()    {}
func (*FunctionAST) exprNode()     {}

// Parser
type Parser struct {
	scanner *Scanner
	curTok  Token
	lit     string
}

func (p *Parser) Init(src []byte) {
	p.scanner = newScanner(src)
	p.getNextToken()
}

func (p *Parser) getNextToken() {
	p.curTok, p.lit = p.scanner.GetTok()
}

// numberexpr ::= number
func (p *Parser) ParseNumberExpr() *NumberExprAST {
	num := NumberExprAST{Val: p.lit}
	p.getNextToken()
	return &num
}

// parenexpr ::= '(' expression ')'
func (p *Parser) ParseParenExpr() ExprAST {
	p.getNextToken() // consume '('
	exp := p.ParseExpression()
	if p.lit != ")" {
		p.scanner.error("Expected ')'")
	}
	p.getNextToken()
	return exp
}

// identifierexpr
//
//	::= identifier
//	::= identifier '(' expression* ')'
func (p *Parser) ParseIdentifierExpr() ExprAST {
	idName := p.lit
	p.getNextToken() // consume identifier

	if p.lit != "(" {
		return &VariableExprAST{Name: idName}
	}

	p.getNextToken() // consume '('
	args := []ExprAST{}
	for p.lit != ")" {
		arg := p.ParseExpression()
		if arg == nil {
			p.scanner.error(fmt.Sprintf("Error parsing call arg: %s", idName))
		}
		args = append(args, arg)

		if p.lit == ")" {
			break
		}

		if p.lit != "," {
			p.scanner.error("Expected ',' or ')'")
		}
		p.getNextToken()
	}

	p.getNextToken() // consume ')'

	return &CallExprAST{Callee: idName, Args: args}
}

// primary
//
//	::= identifierexpr
//	::= numberexpr
//	::= parenexpr
func (p *Parser) ParsePrimary() ExprAST {
	switch p.curTok {
	case TokIdentifier:
		return p.ParseIdentifierExpr()
	case TokNumber:
		return p.ParseNumberExpr()
	case TokOperator:
		if p.lit == "(" {
			return p.ParseParenExpr()
		} else {
			p.scanner.error("Unexpected operator token. Expected '('")
		}
	default:
		p.scanner.error("Unknown token. Expected an expression.")
	}
	return nil
}

var BinopPrecedence = map[rune]int{
	'<': 10,
	'+': 20,
	'-': 20,
	'*': 40,
}

func (p *Parser) GetTokPrecendence() int {
	if p.curTok != TokOperator {
		return -1
	}
	prec, ok := BinopPrecedence[rune(p.lit[0])]
	if !ok {
		return -1
	}
	return prec
}

// expression
//
//	::= primary binoprhs
//
// see https://eli.thegreenplace.net/2012/08/02/parsing-expressions-by-precedence-climbing
func (p *Parser) ParseExpression() ExprAST {
	lhs := p.ParsePrimary()
	return p.ParseBinOpRhs(0, lhs)
}

// binoprhs
//
//	::= ('+' primary)*
func (p *Parser) ParseBinOpRhs(prec int, lhs ExprAST) ExprAST {
	for {
		tokPrec := p.GetTokPrecendence()

		if tokPrec < prec {
			return lhs
		}

		binOp := p.lit
		p.getNextToken() // eat binop

		rhs := p.ParsePrimary()

		nextPrec := p.GetTokPrecendence()
		if nextPrec > tokPrec {
			rhs = p.ParseBinOpRhs(tokPrec+1, rhs)
		}

		lhs = &BinaryExprAST{Op: rune(binOp[0]), LHS: lhs, RHS: rhs}
	}
}

// / prototype
// /   ::= id '(' id* ')'
func (p *Parser) ParsePrototype() *PrototypeAST {
	if p.curTok != TokIdentifier {
		p.scanner.error("Expected identifier")
	}
	fnName := p.lit
	p.getNextToken()

	if p.lit != "(" {
		p.scanner.error("Expected '('")
	}
	p.getNextToken()

	args := make([]string, 0)
	for p.curTok == TokIdentifier {
		args = append(args, p.lit)
		p.getNextToken()
	}

	if p.lit != ")" {
		p.scanner.error("Expected ')")
	}

	p.getNextToken()

	return &PrototypeAST{Name: fnName, Args: args}
}

// definition
//
//	::= 'def' prototype expression
func (p *Parser) ParseDefinition() *FunctionAST {
	p.getNextToken() // consume def
	proto := p.ParsePrototype()
	expr := p.ParseExpression()
	return &FunctionAST{Proto: proto, Body: expr}
}

// external
//
//	::= 'extern' prototype
func (p *Parser) ParseExtern() *PrototypeAST {
	p.getNextToken() // consume extern
	return p.ParsePrototype()
}

// toplevelexpr
//
//	::= expr
func (p *Parser) ParseTopLevelExpr() *FunctionAST {
	e := p.ParseExpression()
	proto := &PrototypeAST{"__anon", []string{}}
	return &FunctionAST{Proto: proto, Body: e}
}

// top
//
//	::= definition | external | expression | ';'
func (p *Parser) ParseTopLevel() ExprAST {
	for {
		switch p.curTok {
		case TokEOF:
			return nil
		case ';':
			p.getNextToken()
		case TokDef:
			return p.ParseDefinition()
		case TokExtern:
			return p.ParseExtern()
		default:
			return p.ParseTopLevelExpr()
		}
	}
}
