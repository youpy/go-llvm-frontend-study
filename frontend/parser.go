package frontend

import (
	"fmt"
	"os"
)

type Parser struct {
	*TokenSet
	TU             *TranslationUnitAST
	VariableTable  []string
	PrototypeTable map[string]int
	FunctionTable  map[string]int
}

func NewParser(filename string) *Parser {
	tokens := LexicalAnalysis(filename)

	return &Parser{
		TokenSet:       tokens,
		VariableTable:  []string{},
		PrototypeTable: make(map[string]int),
		FunctionTable:  make(map[string]int)}
}

func (p *Parser) GetAST() (tu *TranslationUnitAST) {
	if p.TU != nil {
		tu = p.TU
	} else {
		tu = &TranslationUnitAST{[]*PrototypeAST{}, []*FunctionAST{}}
	}

	return
}

func (p *Parser) DoParse() (result bool) {
	if p.TokenSet != nil {
		result = p.visitTranslationUnit()
	} else {
		panic("error at lexer\n")
	}

	return
}

func (p *Parser) visitTranslationUnit() bool {
	p.TU = &TranslationUnitAST{[]*PrototypeAST{}, []*FunctionAST{}}

	// printnum
	paramList := []string{"i"}
	p.TU.Prototypes = append(p.TU.Prototypes, &PrototypeAST{"printnum", paramList})
	p.PrototypeTable["printnum"] = 1

	for {
		if !p.visitExternalDeclaration(p.TU) {
			return false
		}

		if p.getCurType() == TOK_EOF {
			break
		}
	}

	return true
}

func (p *Parser) visitExternalDeclaration(tunit *TranslationUnitAST) bool {
	debug("visitExternalDeclaration")
	proto := p.visitFunctionDeclaration()

	if proto != nil {
		tunit.Prototypes = append(tunit.Prototypes, proto)
		return true
	}

	funcDef := p.visitFunctionDefinition()

	if funcDef != nil {
		tunit.Functions = append(tunit.Functions, funcDef)
		return true
	}

	return false
}

func (p *Parser) visitFunctionDeclaration() (result *PrototypeAST) {
	debug("visitFunctionDeclaration")

	bkup := p.getCurIndex()
	proto := p.visitPrototype()

	if proto == nil {
		return nil
	}

	if p.getCurString() == ";" {
		_, isInPrototypeTable := p.PrototypeTable[proto.Name]
		_, isInFunctionTable := p.FunctionTable[proto.Name]

		if isInPrototypeTable ||
			(isInFunctionTable && p.FunctionTable[proto.Name] != len(proto.Params)) {
			fmt.Fprintf(os.Stderr, "Function: %s is redefined\n", proto.Name)
			return nil
		}

		p.PrototypeTable[proto.Name] = len(proto.Params)

		p.getNextToken()
		result = proto
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	return
}

func (p *Parser) visitFunctionDefinition() *FunctionAST {
	debug("visitFunctionDefinition")

	proto := p.visitPrototype()

	if proto == nil {
		return nil
	} else {
		_, isInPrototypeTable := p.PrototypeTable[proto.Name]
		_, isInFunctionTable := p.FunctionTable[proto.Name]

		if (isInPrototypeTable && p.PrototypeTable[proto.Name] != len(proto.Params)) ||
			isInFunctionTable {
			fmt.Fprintf(os.Stderr, "Function: %s is redefined\n", proto.Name)
			return nil
		}
	}

	p.VariableTable = []string{}

	funcStmt := p.visitFunctionStatement(proto)

	if funcStmt == nil {
		return nil
	}

	p.FunctionTable[proto.Name] = len(proto.Params)

	return &FunctionAST{proto, funcStmt}
}

func (p *Parser) visitPrototype() *PrototypeAST {
	debug("visitPrototype")

	var name string

	bkup := p.getCurIndex()
	isFirstParam := true
	paramList := []string{}

	if p.getCurType() == TOK_INT {
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	if p.getCurType() == TOK_IDENTIFIER {
		name = p.getCurString()
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	if p.getCurType() == TOK_SYMBOL && p.getCurString() == "(" {
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	for {
		if !isFirstParam &&
			p.getCurType() == TOK_SYMBOL &&
			p.getCurString() == "," {
			p.getNextToken()
		}

		if p.getCurType() == TOK_INT {
			p.getNextToken()
		} else {
			break
		}

		if p.getCurType() == TOK_IDENTIFIER {
			for _, param := range paramList {
				if param == p.getCurString() {
					p.applyTokenIndex(bkup)
					return nil
				}
			}

			isFirstParam = false
			paramList = append(paramList, p.getCurString())
			p.getNextToken()
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.getCurType() == TOK_SYMBOL && p.getCurString() == ")" {
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	return &PrototypeAST{name, paramList}
}

func (p *Parser) visitFunctionStatement(proto *PrototypeAST) (funcStmt *FunctionStmtAST) {
	debug("visitFunctionStatement")

	bkup := p.getCurIndex()

	if p.getCurString() == "{" {
		p.getNextToken()
	} else {
		return nil
	}

	funcStmt = &FunctionStmtAST{[]*VariableDeclAST{}, []AST{}}

	for i, _ := range proto.Params {
		vdecl := &VariableDeclAST{proto.Params[i], Decl_param, &BaseAST{VariableDeclID}}
		p.VariableTable = append(p.VariableTable, vdecl.Name)
		funcStmt.VariableDecls = append(funcStmt.VariableDecls, vdecl)
	}

	for {
		if vdecl := p.visitVariableDeclaration(); vdecl != nil {
			vdecl.Type = Decl_local

			for _, availableVdecl := range p.VariableTable {
				if availableVdecl == vdecl.Name {
					return nil
				}
			}

			p.VariableTable = append(p.VariableTable, vdecl.Name)
			funcStmt.VariableDecls = append(funcStmt.VariableDecls, vdecl)
		} else {
			break
		}
	}

	for {
		if stmt := p.visitStatement(); stmt != nil {
			funcStmt.StmtLists = append(funcStmt.StmtLists, stmt)
		} else {
			break
		}
	}

	if len(funcStmt.StmtLists) > 0 {
		lastStmt := funcStmt.StmtLists[len(funcStmt.StmtLists)-1]

		if lastStmt.GetID() != JumpStmtID {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.getCurString() == "}" {
		p.getNextToken()
		return
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	return
}

func (p *Parser) visitVariableDeclaration() *VariableDeclAST {
	debug("visitVariableDeclaration")

	var name string

	bkup := p.getCurIndex()

	if p.getCurType() == TOK_INT {
		p.getNextToken()
	} else {
		return nil
	}

	if p.getCurType() == TOK_IDENTIFIER {
		name = p.getCurString()
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == ";" {
		p.getNextToken()
	} else {
		p.applyTokenIndex(bkup)
		return nil
	}

	return &VariableDeclAST{Name: name, BaseAST: &BaseAST{VariableDeclID}}
}

func (p *Parser) visitStatement() (result AST) {
	debug("visitStatement")

	bkup := p.getCurIndex()

	for {
		if expr := p.visitExpressionStatement(); expr != nil {
			result = expr
			return
		} else if jump := p.visitJumpStatement(); jump != nil {
			result = jump
			return
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	return
}

func (p *Parser) visitExpressionStatement() AST {
	debug("visitExpressionStatement")

	if p.getCurString() == ";" {
		p.getNextToken()
		return &NullExprAST{&BaseAST{NullExprID}}
	} else if assignExpr := p.visitAssignmentExpression(); assignExpr != nil {
		if p.getCurString() == ";" {
			p.getNextToken()
			return assignExpr
		}
	}

	return nil
}

func (p *Parser) visitAssignmentExpression() AST {
	debug("visitAssignmentExpression")

	bkup := p.getCurIndex()

	if p.getCurType() == TOK_IDENTIFIER {
		found := false

		for _, availableVdecl := range p.VariableTable {
			if availableVdecl == p.getCurString() {
				found = true
			}
		}

		if found {
			lhs := &VariableAST{p.getCurString(), &BaseAST{VariableID}}
			p.getNextToken()

			if p.getCurType() == TOK_SYMBOL && p.getCurString() == "=" {
				p.getNextToken()

				if rhs := p.visitAdditiveExpression(nil); rhs != nil {
					return &BinaryExprAST{"=", lhs, rhs, &BaseAST{BinaryExprID}}
				} else {
					p.applyTokenIndex(bkup)
				}
			} else {
				p.applyTokenIndex(bkup)
			}
		} else {
			p.applyTokenIndex(bkup)
		}
	}

	addExpr := p.visitAdditiveExpression(nil)

	if addExpr != nil {
		return addExpr
	}

	return nil
}

func (p *Parser) visitAdditiveExpression(lhs AST) AST {
	debug("visitAdditiveExpression")

	bkup := p.getCurIndex()

	if lhs == nil {
		lhs = p.visitMultiplicativeExpression(nil)
	}

	if lhs == nil {
		return nil
	}

	if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == "+" {
		p.getNextToken()

		rhs := p.visitMultiplicativeExpression(nil)

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"+", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == "-" {
		p.getNextToken()

		rhs := p.visitMultiplicativeExpression(nil)

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"-", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	return lhs
}

func (p *Parser) visitMultiplicativeExpression(lhs AST) AST {
	debug("visitMultiplicativeExpression")

	bkup := p.getCurIndex()

	if lhs == nil {
		lhs = p.visitPostfixExpression()
	}

	if lhs == nil {
		return nil
	}

	if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == "*" {
		p.getNextToken()

		rhs := p.visitPostfixExpression()

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"*", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == "/" {
		p.getNextToken()

		rhs := p.visitPostfixExpression()

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"/", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	return lhs
}

func (p *Parser) visitPostfixExpression() (result AST) {
	debug("visitPostfixExpression")

	var paramNum int
	var isInTable bool

	bkup := p.getCurIndex()

	if p.getCurType() == TOK_IDENTIFIER {
		callee := p.getCurString()
		p.getNextToken()

		if paramNum, isInTable = p.PrototypeTable[callee]; !isInTable {
			if paramNum, isInTable = p.FunctionTable[callee]; !isInTable {
				p.applyTokenIndex(bkup)

				if priExpr := p.visitPrimaryExpression(); priExpr != nil {
					result = priExpr
				}

				return
			}
		}

		if p.getCurType() != TOK_SYMBOL ||
			p.getCurString() != "(" {
			p.applyTokenIndex(bkup)

			if priExpr := p.visitPrimaryExpression(); priExpr != nil {
				result = priExpr
			}

			return
		}

		p.getNextToken()

		args := []AST{}

		for {
			if assignExpr := p.visitAssignmentExpression(); assignExpr != nil {
				args = append(args, assignExpr)

				if p.getCurType() == TOK_SYMBOL && p.getCurString() == "," {
					p.getNextToken()
				} else {
					break
				}
			}
		}

		if len(args) != paramNum {
			p.applyTokenIndex(bkup)
			return nil
		}

		if p.getCurType() == TOK_SYMBOL && p.getCurString() == ")" {
			p.getNextToken()
			result = &CallExprAST{callee, args, &BaseAST{CallExprID}}
		} else {
			p.applyTokenIndex(bkup)
		}
	}

	if result == nil {
		if priExpr := p.visitPrimaryExpression(); priExpr != nil {
			result = priExpr
			return
		}
	}

	return
}

func (p *Parser) visitPrimaryExpression() AST {
	debug("visitPrimaryExpression")

	bkup := p.getCurIndex()

	if p.getCurType() == TOK_IDENTIFIER {
		found := false

		for _, availableVdecl := range p.VariableTable {
			if availableVdecl == p.getCurString() {
				found = true
			}
		}

		if found {
			name := p.getCurString()
			p.getNextToken()
			return &VariableAST{name, &BaseAST{VariableID}}
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	} else if p.getCurType() == TOK_DIGIT {
		val := p.getCurNumVal()
		p.getNextToken()
		return &NumberAST{val, &BaseAST{NumberID}}
	} else if p.getCurType() == TOK_SYMBOL &&
		p.getCurString() == "-" {
		p.getNextToken()
		if p.getCurType() == TOK_DIGIT {
			val := p.getCurNumVal()
			p.getNextToken()
			return &NumberAST{-val, &BaseAST{NumberID}}
		} else {
			p.applyTokenIndex(bkup)
			return nil
		}
	}

	return nil
}

func (p *Parser) visitJumpStatement() AST {
	debug("visitJumpStatement")

	bkup := p.getCurIndex()

	if p.getCurType() == TOK_RETURN {
		p.getNextToken()

		if assignExpr := p.visitAssignmentExpression(); assignExpr != nil {
			if p.getCurType() == TOK_SYMBOL &&
				p.getCurString() == ";" {
				p.getNextToken()
				return &JumpStmtAST{assignExpr, &BaseAST{JumpStmtID}}
			}
		}
	}

	p.applyTokenIndex(bkup)
	return nil
}

func debug(msg string) {
	if os.Getenv("DEBUG") != "" {
		fmt.Println(msg)
	}
}
