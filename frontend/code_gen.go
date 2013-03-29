package frontend

import (
	"fmt"
	"github.com/axw/gollvm/llvm"
	"os"
)

type CodeGen struct {
	curFunc     llvm.Value
	variableMap map[string]llvm.Value
	module      llvm.Module
	builder     llvm.Builder
}

func NewCodeGen(c llvm.Context) *CodeGen {
	builder := c.NewBuilder()

	return &CodeGen{builder: builder, variableMap: make(map[string]llvm.Value)}
}

func (c *CodeGen) DoCodeGen(tunit *TranslationUnitAST, name string) bool {
	return c.generateTranslationUnit(tunit, name)
}

func (c *CodeGen) GetModule() (module llvm.Module) {
	if !c.module.IsNil() {
		module = c.module
	} else {
		module = c.context().NewModule("null")
	}

	return
}

func (c *CodeGen) context() llvm.Context {
	return llvm.GlobalContext()
}

func (c *CodeGen) generateTranslationUnit(tunit *TranslationUnitAST, name string) bool {
	c.module = c.context().NewModule(name)

	for i, _ := range tunit.Prototypes {
		proto := tunit.Prototypes[i]

		if proto == nil {
			break
		} else if _, ok := c.generatePrototype(proto, c.module); !ok {
			return false
		}
	}

	for i, _ := range tunit.Functions {
		funcAST := tunit.Functions[i]

		if funcAST == nil {
			break

		} else if _, ok := c.generateFunctionDefinition(funcAST, c.module); !ok {
			return false
		}
	}

	return true
}

func (c *CodeGen) generatePrototype(proto *PrototypeAST, module llvm.Module) (fun llvm.Value, ok bool) {
	fun = module.NamedFunction(proto.Name)

	if !fun.IsNil() {
		if fun.ParamsCount() == len(proto.Params) && fun.BasicBlocksCount() == 0 {
			return fun, true
		} else {
			fmt.Fprintf(os.Stderr, "error::function %s is redefined", proto.Name)
			return fun, false
		}
	}

	intArgs := []llvm.Type{}

	for i := 0; i < len(proto.Params); i++ {
		intArgs = append(intArgs, c.context().Int32Type())
	}

	intType := llvm.FunctionType(c.context().Int32Type(), intArgs, false)

	fun = llvm.AddFunction(module, proto.Name, intType)
	fun.SetLinkage(llvm.ExternalLinkage)

	for i, param := range fun.Params() {
		param.SetName(proto.Params[i] + "_arg")
	}

	return fun, true
}

func (c *CodeGen) generateFunctionDefinition(funcAST *FunctionAST, module llvm.Module) (value llvm.Value, ok bool) {
	fun, ok := c.generatePrototype(funcAST.Proto, module)

	if !ok {
		return value, false
	}

	c.curFunc = fun
	c.variableMap = make(map[string]llvm.Value)

	basicBlock := c.context().AddBasicBlock(fun, "entry")
	c.builder.SetInsertPointAtEnd(basicBlock)

	c.generateFunctionStatement(funcAST.Body)

	return fun, true
}

func (c *CodeGen) generateFunctionStatement(funcStmt *FunctionStmtAST) llvm.Value {
	var v llvm.Value

	for i, _ := range funcStmt.VariableDecls {
		v = c.generateVariableDeclaration(funcStmt.VariableDecls[i])
	}

	for i, _ := range funcStmt.StmtLists {
		v = c.generateStatement(funcStmt.StmtLists[i])
	}

	return v
}

func (c *CodeGen) generateVariableDeclaration(vdeclAST *VariableDeclAST) llvm.Value {
	alloca := c.builder.CreateAlloca(c.context().Int32Type(), vdeclAST.Name)

	c.variableMap[vdeclAST.Name] = alloca

	if vdeclAST.Type == Decl_param {
		for _, param := range c.curFunc.Params() {
			if param.Name() == vdeclAST.Name+"_arg" {
				c.builder.CreateStore(param, alloca)
				break
			}
		}
	}

	return alloca
}

func (c *CodeGen) generateStatement(stmt AST) (value llvm.Value) {
	switch stmt.GetID() {
	case BinaryExprID:
		value = c.generateBinaryExpression(stmt.(*BinaryExprAST))
	case CallExprID:
		value = c.generateCallExpression(stmt.(*CallExprAST))
	case JumpStmtID:
		value = c.generateJumpStatement(stmt.(*JumpStmtAST))
	}

	return
}

func (c *CodeGen) generateBinaryExpression(binExpr *BinaryExprAST) (value llvm.Value) {
	var lhsV llvm.Value
	var rhsV llvm.Value

	lhs := binExpr.LHS
	rhs := binExpr.RHS

	if binExpr.Op == "=" {
		fmt.Println("genStore")

		lhsVar := lhs.(*VariableAST)
		lhsV = c.variableMap[lhsVar.Name]
	} else {
		// lhs?
		// binary?
		switch lhs.GetID() {
		case BinaryExprID:
			lhsV = c.generateBinaryExpression(lhs.(*BinaryExprAST))
		case VariableID:
			lhsV = c.generateVariable(lhs.(*VariableAST))
		case NumberID:
			lhsV = c.generateNumber(lhs.(*NumberAST).Val)
		}
	}

	switch rhs.GetID() {
	case BinaryExprID:
		rhsV = c.generateBinaryExpression(rhs.(*BinaryExprAST))
	case VariableID:
		rhsV = c.generateVariable(rhs.(*VariableAST))
	case NumberID:
		rhsV = c.generateNumber(rhs.(*NumberAST).Val)
	}

	switch binExpr.Op {
	case "=":
		value = c.builder.CreateStore(rhsV, lhsV)
	case "+":
		value = c.builder.CreateAdd(lhsV, rhsV, "add_tmp")
	case "-":
		value = c.builder.CreateSub(lhsV, rhsV, "sub_tmp")
	case "*":
		value = c.builder.CreateMul(lhsV, rhsV, "mul_tmp")
	case "/":
		value = c.builder.CreateSDiv(lhsV, rhsV, "div_tmp")
	}

	return
}

func (c *CodeGen) generateCallExpression(callExpr *CallExprAST) llvm.Value {
	var argV llvm.Value

	argVec := []llvm.Value{}

	for _, arg := range callExpr.Args {
		switch arg.GetID() {
		case CallExprID:
			argV = c.generateCallExpression(arg.(*CallExprAST))
		case BinaryExprID:
			binExpr := arg.(*BinaryExprAST)
			argV = c.generateBinaryExpression(binExpr)

			if binExpr.Op == "=" {
				variable := binExpr.LHS
				argV = c.builder.CreateLoad(c.variableMap[variable.(*VariableAST).Name], "arg_val")
			}
		case VariableID:
			argV = c.generateVariable(arg.(*VariableAST))
		case NumberID:
			argV = c.generateNumber(arg.(*NumberAST).Val)
		}

		argVec = append(argVec, argV)
	}

	return c.builder.CreateCall(c.module.NamedFunction(callExpr.Callee), argVec, "call_tmp")
}

func (c *CodeGen) generateJumpStatement(jumpStmt *JumpStmtAST) llvm.Value {
	var retV llvm.Value

	expr := jumpStmt.Expr

	switch expr.GetID() {
	case BinaryExprID:
		retV = c.generateBinaryExpression(expr.(*BinaryExprAST))
	case VariableID:
		retV = c.generateVariable(expr.(*VariableAST))
	case NumberID:
		retV = c.generateNumber(expr.(*NumberAST).Val)
	}

	return c.builder.CreateRet(retV)
}

func (c *CodeGen) generateVariable(variable *VariableAST) llvm.Value {
	return c.builder.CreateLoad(c.variableMap[variable.Name], "var_tmp")
}

func (c *CodeGen) generateNumber(value int) llvm.Value {
	return llvm.ConstInt(c.context().Int32Type(), uint64(value), false)
}
