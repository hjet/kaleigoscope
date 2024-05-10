package main

import (
	"fmt"
	"tinygo.org/x/go-llvm"
)

type CodeGen struct {
	ctx         *llvm.Context
	module      *llvm.Module
	builder     *llvm.Builder
	symbolTable map[string]llvm.Value
	protos      map[string]*PrototypeAST
}

func (cg *CodeGen) Init() {
	ctx := llvm.NewContext()
	mod := ctx.NewModule("")
	builder := ctx.NewBuilder()

	cg.ctx = &ctx
	cg.module = &mod
	cg.builder = &builder

	cg.symbolTable = make(map[string]llvm.Value)
	cg.protos = make(map[string]*PrototypeAST)
}

func (cg *CodeGen) error(e string) {
	panic(e)
}

func (cg *CodeGen) resetModule() {
	mod := cg.ctx.NewModule("")
	cg.module = &mod
	bu := cg.ctx.NewBuilder()
	cg.builder = &bu
	cg.symbolTable = make(map[string]llvm.Value)
}

func (cg *CodeGen) getFunction(name string) llvm.Value {
	fun := cg.module.NamedFunction(name) // check if fn in module
	if !fun.IsNil() {
		return fun
	}

	if proto, ok := cg.protos[name]; ok { // check global proto map
		return proto.codeGen(cg)
	}

	return llvm.Value{}
}

func (n *NumberExprAST) codeGen(cg *CodeGen) llvm.Value {
	return llvm.ConstFloatFromString(cg.ctx.DoubleType(), n.Val)
}

func (v *VariableExprAST) codeGen(cg *CodeGen) llvm.Value {
	val, ok := cg.symbolTable[v.Name]
	if !ok {
		cg.error(fmt.Sprintf("Unknown var: %s", v.Name))
	}
	return val
}

func (b *BinaryExprAST) codeGen(cg *CodeGen) llvm.Value {
	l := b.LHS.codeGen(cg)
	r := b.RHS.codeGen(cg)

	switch b.Op {
	case '+':
		return cg.builder.CreateFAdd(l, r, "addtmp")
	case '-':
		return cg.builder.CreateFSub(l, r, "subtmp")
	case '*':
		return cg.builder.CreateFMul(l, r, "multmp")
	case '<':
		l := cg.builder.CreateFCmp(llvm.FloatULT, l, r, "cmptmp")
		return cg.builder.CreateUIToFP(l, cg.ctx.DoubleType(), "booltmp")
	default:
		cg.error(fmt.Sprintf("Invalid binop: %c", b.Op))
	}
	return llvm.Value{}
}

func (c *CallExprAST) codeGen(cg *CodeGen) llvm.Value {
	fun := cg.getFunction(c.Callee)
	if fun.IsNil() {
		cg.error(fmt.Sprintf("Error: Unknown function referenced: %s", c.Callee))
	}
	funType := fun.GlobalValueType() // see https://llvm.org/docs/OpaquePointers.html

	if len(fun.Params()) != len(c.Args) {
		cg.error(fmt.Sprintf("Error: Expected %d args in call to %s; recieved %d", len(fun.Params()), c.Callee, len(c.Args)))
	}

	args := make([]llvm.Value, 0, len(c.Args))
	for _, arg := range c.Args {
		args = append(args, arg.codeGen(cg))
	}
	return cg.builder.CreateCall(funType, fun, args, "calltmp")
}

func (p *PrototypeAST) codeGen(cg *CodeGen) llvm.Value {
	paramTypes := make([]llvm.Type, 0, len(p.Args))
	for range p.Args {
		paramTypes = append(paramTypes, cg.ctx.DoubleType())
	}
	funcType := llvm.FunctionType(cg.ctx.DoubleType(), paramTypes, false)
	newFunc := llvm.AddFunction(*cg.module, p.Name, funcType)
	newFunc.SetLinkage(llvm.ExternalLinkage)

	for i, argName := range p.Args {
		newFunc.Params()[i].SetName(argName)
	}

	return newFunc
}

func (f *FunctionAST) codeGen(cg *CodeGen) llvm.Value {
	fnName := f.Proto.Name
	fun := cg.module.NamedFunction(fnName) // get fn from module
	if fun.IsNil() {
		fun = f.Proto.codeGen(cg)
	}
	if fun.IsNil() {
		cg.error(fmt.Sprintf("Error: could not get or create function: %s", fnName))
	}
	if fun.BasicBlocksCount() != 0 { // TODO: allow redefining fn
		cg.error(fmt.Sprintf("Error: cannot redefine function: %s", fnName))
	}

	cg.protos[fnName] = f.Proto // add proto to global map (see ch. 4)

	blk := cg.ctx.AddBasicBlock(fun, "entry")
	cg.builder.SetInsertPointAtEnd(blk)

	cg.symbolTable = make(map[string]llvm.Value) // reset function symbol table
	for _, param := range fun.Params() {
		cg.symbolTable[param.Name()] = param
	}

	body := f.Body.codeGen(cg)
	if body.IsNil() {
		fun.EraseFromParentAsFunction()
		cg.error(fmt.Sprintf("Error: could not codegen function body: %s", fnName))
	}
	cg.builder.CreateRet(body)

	if err := llvm.VerifyFunction(fun, llvm.PrintMessageAction); err != nil {
		fun.EraseFromParentAsFunction()
		cg.error(fmt.Sprintf("Error: function %s didn't pass verifier", fnName))
	}

	return fun
}
