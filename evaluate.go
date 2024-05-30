package main

import (
	"fmt"
	"tinygo.org/x/go-llvm"
)

type Evaluator struct {
	cg            *CodeGen
	target        *llvm.Target
	targetMachine *llvm.TargetMachine
	jit           *llvm.ExecutionEngine
}

func NewEvaluator() *Evaluator {
	eval := Evaluator{}

	llvm.LinkInMCJIT()
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()

	cg := CodeGen{}
	cg.Init()
	eval.cg = &cg

	target, err := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())
	if err != nil {
		eval.error("Error: Could not get target")
	}
	eval.target = &target

	tm := target.CreateTargetMachine(llvm.DefaultTargetTriple(), "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)
	eval.targetMachine = &tm

	opts := llvm.NewMCJITCompilerOptions()
	ee, err := llvm.NewMCJITCompiler(*eval.cg.module, opts)
	if err != nil {
		eval.error("Error: Could not create MCJIT")
	}
	eval.jit = &ee

	return &eval
}

func (e *Evaluator) error(errStr string) {
	panic(errStr)
}

func (e *Evaluator) Optimize() {
	pbo := llvm.NewPassBuilderOptions()
	err := e.cg.module.RunPasses("default<Os>", *e.targetMachine, pbo)
	if err != nil {
		e.error("Error: Error running default optimization passes")
	}
}

func (e *Evaluator) PrintAsm() {
	mem, err := e.targetMachine.EmitToMemoryBuffer(*e.cg.module, llvm.AssemblyFile)
	if err != nil {
		e.error("Error: Error emitting ASM to MemoryBuffer")
	}

	bytes := mem.Bytes()
	fmt.Println("***** ASSEMBLY *****")
	fmt.Println(string(bytes))
}

func (e *Evaluator) Evaluate(code string) float64 {
	parser := &Parser{}
	parser.Init([]byte(code))
	expr := parser.ParseTopLevel()
	_ = expr.codeGen(e.cg)

	// check if it's an anonymous function, if not return
	fnExpr, ok := expr.(*FunctionAST)
	if !(ok && fnExpr.Proto.Name == "__anon") {
		return 0
	}

	//e.Optimize()
	e.PrintAsm()

	e.jit.AddModule(*e.cg.module)
	e.cg.resetModule()

	// execute last TLE -- TODO: fn *?
	fn := e.jit.FindFunction("__anon")
	ret := e.jit.RunFunction(fn, []llvm.GenericValue{})
	retFloat := ret.Float(e.cg.ctx.DoubleType())
	return retFloat
}
