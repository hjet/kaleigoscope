package main

import (
	"fmt"
)

// ASMGen generates ARMv8 A64 from an AST
// TODO: cleanup
// TODO: add tests
type ASMGen struct {
	symbolTable map[string]int
	stkSize     int
}

func (ag *ASMGen) error(e string) {
	panic(e)
}

func newASMGen() *ASMGen {
	ag := &ASMGen{}
	ag.symbolTable = make(map[string]int)
	return ag
}

func (ag *ASMGen) addPrologue() {
	fmt.Printf("	stp	x29, x30, [sp, #-16]!\n")
	fmt.Printf("	mov	x29, sp\n")
}

func (ag *ASMGen) addEpilogue() {
	fmt.Printf("	ldp     x29, x30, [sp], #16\n")
	fmt.Printf("	ret\n\n")
}

func (ag *ASMGen) push() {
	fmt.Printf("	str d0, [sp, #-16]!\n")
}

func (ag *ASMGen) pop(reg string) {
	fmt.Printf("	ldr	%s, [sp], #16\n", reg)
}

func (n *NumberExprAST) asmGen(ag *ASMGen) {
	// TODO: this won't work for many vals...
	fmt.Printf("	fmov d0, #%s\n", n.Val)
}

func (v *VariableExprAST) asmGen(ag *ASMGen) {
	offset, ok := ag.symbolTable[v.Name]
	if !ok {
		ag.error(fmt.Sprintf("Unknown var: %s", v.Name))
	}
	fmt.Printf("	ldr d0, [x29, #-%d]\n", offset)
}

func (b *BinaryExprAST) asmGen(ag *ASMGen) {
	b.RHS.asmGen(ag)
	ag.push()
	b.LHS.asmGen(ag)
	ag.pop("d1")

	switch b.Op {
	case '+':
		fmt.Printf("	fadd d0, d0, d1\n")
	case '-':
		fmt.Printf("	fsub d0, d0, d1\n")
	case '*':
		fmt.Printf("	fmul d0, d0, d1\n")
	case '<':
		ag.error("< not implemented yet... :(")
	default:
		ag.error(fmt.Sprintf("Invalid binop: %c", b.Op))
	}
}

func (c *CallExprAST) asmGen(ag *ASMGen) {
	funName := c.Callee
	for ix := len(c.Args) - 1; ix >= 0; ix-- {
		c.Args[ix].asmGen(ag)
		fmt.Printf("	fmov	d%d, d0\n", ix)
	}
	// TODO: support passing args via stack
	fmt.Printf("	bl _%s\n", funName)
}

func (p *PrototypeAST) asmGen(ag *ASMGen) {}

func (f *FunctionAST) asmGen(ag *ASMGen) {

	funName := f.Proto.Name
	fmt.Printf("	.globl  _%s\n", funName)
	fmt.Printf("	.p2align 2\n")

	fmt.Printf("_%s:\n", funName)

	offset := 0
	for _, arg := range f.Proto.Args {
		offset += 8
		ag.symbolTable[arg] = offset
	}
	ag.stkSize = offset
	ag.addPrologue()
	fmt.Printf("	sub sp, sp, %d\n", offset)

	// store args in stack at offset
	for ix, arg := range f.Proto.Args {
		fmt.Printf("	str d%d, [x29, #-%d]\n", ix, ag.symbolTable[arg])
	}

	f.Body.asmGen(ag)

	fmt.Printf("	add sp, sp, %d\n", offset)
	ag.addEpilogue()
}

func main() {
	ag := newASMGen()
	parser := &Parser{}
	parser.Init([]byte("def foo(x y) x - y"))
	e := parser.ParseTopLevel()
	fmt.Printf("	.section        __TEXT,__text,regular,pure_instructions\n")
	fmt.Printf("	.build_version macos, 14, 0\n")
	e.asmGen(ag)
	parser.Init([]byte("def bar(x y) x * y"))
	e = parser.ParseTopLevel()
	e.asmGen(ag)

}
