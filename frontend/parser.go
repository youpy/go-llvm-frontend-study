package frontend

import (
	"fmt"
)

type Parser struct {
	Tokens *TokenSet
	TU     *TranslationUnitAST
}

func NewParser(filename string) *Parser {
	tokens := LexicalAnalysis(filename)

	return &Parser{Tokens: tokens}
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
	if p.Tokens != nil {
		result = p.visitTranslationUnit()
	} else {
		panic("error at lexer\n")
	}

	return
}

func (p *Parser) visitTranslationUnit() bool {
	p.TU = &TranslationUnitAST{[]*PrototypeAST{}, []*FunctionAST{}}

	for {
		if !p.visitExternalDeclaration(p.TU) {
			return false
		}

		if p.Tokens.getCurType() == TOK_EOF {
			break
		}
	}

	return true
}

func (p *Parser) visitExternalDeclaration(tunit *TranslationUnitAST) bool {
	fmt.Println("visitExternalDeclaration")
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
	fmt.Println("visitFunctionDeclaration")

	bkup := p.Tokens.getCurIndex()
	proto := p.visitPrototype()

	if proto == nil {
		return nil
	}

	if p.Tokens.getCurString() == ";" {
		p.Tokens.getNextToken()

		result = proto
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	return
}

func (p *Parser) visitFunctionDefinition() *FunctionAST {
	fmt.Println("visitFunctionDefinition")

	proto := p.visitPrototype()

	if proto == nil {
		return nil
	}

	funcStmt := p.visitFunctionStatement(proto)

	if funcStmt == nil {
		return nil
	}

	return &FunctionAST{proto, funcStmt}
}

func (p *Parser) visitPrototype() *PrototypeAST {
	fmt.Println("visitPrototype")

	var name string

	bkup := p.Tokens.getCurIndex()
	isFirstParam := true
	paramList := []string{}

	if p.Tokens.getCurType() == TOK_INT {
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	if p.Tokens.getCurType() == TOK_IDENTIFIER {
		name = p.Tokens.getCurString()
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	if p.Tokens.getCurType() == TOK_SYMBOL && p.Tokens.getCurString() == "(" {
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	for {
		if p.Tokens.getCurType() == TOK_INT {
			p.Tokens.getNextToken()
		} else {
			break
		}

		if !isFirstParam &&
			p.Tokens.getCurType() == TOK_SYMBOL &&
			p.Tokens.getCurString() == "," {
			p.Tokens.getNextToken()
		}

		if p.Tokens.getCurType() == TOK_IDENTIFIER {
			paramList = append(paramList, p.Tokens.getCurString())
			p.Tokens.getNextToken()
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.Tokens.getCurType() == TOK_SYMBOL && p.Tokens.getCurString() == ")" {
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	return &PrototypeAST{name, paramList}
}

func (p *Parser) visitFunctionStatement(proto *PrototypeAST) (funcStmt *FunctionStmtAST) {
	fmt.Println("visitFunctionStatement")

	bkup := p.Tokens.getCurIndex()

	if p.Tokens.getCurString() == "{" {
		p.Tokens.getNextToken()
	} else {
		return nil
	}

	funcStmt = &FunctionStmtAST{[]*VariableDeclAST{}, []AST{}}

	for i, _ := range proto.Params {
		vdecl := &VariableDeclAST{proto.Params[i], Decl_param, &BaseAST{VariableDeclID}}
		funcStmt.VariableDecls = append(funcStmt.VariableDecls, vdecl)
	}

	for {
		if vdecl := p.visitVariableDeclaration(); vdecl != nil {
			vdecl.Type = Decl_local
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

	if p.Tokens.getCurString() == "}" {
		p.Tokens.getNextToken()
		return
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	return
}

func (p *Parser) visitVariableDeclaration() *VariableDeclAST {
	fmt.Println("visitVariableDeclaration")

	var name string

	bkup := p.Tokens.getCurIndex()

	if p.Tokens.getCurType() == TOK_INT {
		p.Tokens.getNextToken()
	} else {
		return nil
	}

	if p.Tokens.getCurType() == TOK_IDENTIFIER {
		name = p.Tokens.getCurString()
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == ";" {
		p.Tokens.getNextToken()
	} else {
		p.Tokens.applyTokenIndex(bkup)
		return nil
	}

	return &VariableDeclAST{Name: name, BaseAST: &BaseAST{VariableDeclID}}
}

func (p *Parser) visitStatement() (result AST) {
	fmt.Println("visitStatement")

	bkup := p.Tokens.getCurIndex()

	for {
		if expr := p.visitExpressionStatement(); expr != nil {
			result = expr
			return
		} else if jump := p.visitJumpStatement(); jump != nil {
			result = jump
			return
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	return
}

func (p *Parser) visitExpressionStatement() AST {
	fmt.Println("visitExpressionStatement")

	if p.Tokens.getCurString() == ";" {
		p.Tokens.getNextToken()
		return &NullExprAST{&BaseAST{NullExprID}}
	} else if assignExpr := p.visitAssignmentExpression(); assignExpr != nil {
		if p.Tokens.getCurString() == ";" {
			p.Tokens.getNextToken()
			return assignExpr
		}
	}

	return nil
}

func (p *Parser) visitAssignmentExpression() AST {
	fmt.Println("visitAssignmentExpression")

	bkup := p.Tokens.getCurIndex()

	if p.Tokens.getCurType() == TOK_IDENTIFIER {
		lhs := &VariableAST{p.Tokens.getCurString(), &BaseAST{VariableID}}
		p.Tokens.getNextToken()

		if p.Tokens.getCurType() == TOK_SYMBOL && p.Tokens.getCurString() == "=" {
			p.Tokens.getNextToken()

			if rhs := p.visitAdditiveExpression(nil); rhs != nil {
				return &BinaryExprAST{"=", lhs, rhs, &BaseAST{BinaryExprID}}
			} else {
				p.Tokens.applyTokenIndex(bkup)
			}
		} else {
			p.Tokens.applyTokenIndex(bkup)
		}
	}

	addExpr := p.visitAdditiveExpression(nil)

	if addExpr != nil {
		return addExpr
	}

	return nil
}

func (p *Parser) visitAdditiveExpression(lhs AST) AST {
	fmt.Println("visitAdditiveExpression")

	bkup := p.Tokens.getCurIndex()

	if lhs == nil {
		lhs = p.visitMultiplicativeExpression(nil)
	}

	if lhs == nil {
		return nil
	}

	if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == "+" {
		p.Tokens.getNextToken()

		rhs := p.visitMultiplicativeExpression(nil)

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"+", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == "-" {
		p.Tokens.getNextToken()

		rhs := p.visitMultiplicativeExpression(nil)

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"-", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	return lhs
}

func (p *Parser) visitMultiplicativeExpression(lhs AST) AST {
	fmt.Println("visitMultiplicativeExpression")

	bkup := p.Tokens.getCurIndex()

	if lhs == nil {
		lhs = p.visitPostfixExpression()
	}

	if lhs == nil {
		return nil
	}

	if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == "*" {
		p.Tokens.getNextToken()

		rhs := p.visitPostfixExpression()

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"*", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == "/" {
		p.Tokens.getNextToken()

		rhs := p.visitPostfixExpression()

		if rhs != nil {
			return p.visitMultiplicativeExpression(
				&BinaryExprAST{"/", lhs, rhs, &BaseAST{BinaryExprID}})
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	return lhs
}

func (p *Parser) visitPostfixExpression() (result AST) {
	fmt.Println("visitPostfixExpression")

	bkup := p.Tokens.getCurIndex()

	if priExpr := p.visitPrimaryExpression(); priExpr != nil {
		result = priExpr
		return
	}

	if p.Tokens.getCurType() == TOK_IDENTIFIER {
		callee := p.Tokens.getCurString()
		p.Tokens.getNextToken()

		if p.Tokens.getCurType() != TOK_SYMBOL ||
			p.Tokens.getCurString() != "(" {
			p.Tokens.applyTokenIndex(bkup)
			return
		}

		p.Tokens.getNextToken()

		args := []AST{}

		for assignExpr := p.visitAssignmentExpression(); assignExpr != nil; {
			args = append(args, assignExpr)
			if p.Tokens.getCurType() == TOK_SYMBOL && p.Tokens.getCurString() == "," {
				p.Tokens.getNextToken()
			} else {
				break
			}
		}

		if p.Tokens.getCurType() == TOK_SYMBOL && p.Tokens.getCurString() == ")" {
			p.Tokens.getNextToken()
			result = &CallExprAST{callee, args, &BaseAST{CallExprID}}
		} else {
			p.Tokens.applyTokenIndex(bkup)
		}
	}

	return
}

func (p *Parser) visitPrimaryExpression() AST {
	fmt.Println("visitPrimaryExpression")

	bkup := p.Tokens.getCurIndex()

	if p.Tokens.getCurType() == TOK_IDENTIFIER {
		name := p.Tokens.getCurString()
		p.Tokens.getNextToken()
		return &VariableAST{name, &BaseAST{VariableID}}
	} else if p.Tokens.getCurType() == TOK_DIGIT {
		val := p.Tokens.getCurNumVal()
		p.Tokens.getNextToken()
		return &NumberAST{val, &BaseAST{NumberID}}
	} else if p.Tokens.getCurType() == TOK_SYMBOL &&
		p.Tokens.getCurString() == "-" {
		p.Tokens.getNextToken()
		if p.Tokens.getCurType() == TOK_DIGIT {
			val := p.Tokens.getCurNumVal()
			p.Tokens.getNextToken()
			return &NumberAST{-val, &BaseAST{NumberID}}
		} else {
			p.Tokens.applyTokenIndex(bkup)
			return nil
		}
	}

	return nil
}

func (p *Parser) visitJumpStatement() AST {
	fmt.Println("visitJumpStatement")

	bkup := p.Tokens.getCurIndex()

	if p.Tokens.getCurType() == TOK_RETURN {
		p.Tokens.getNextToken()

		if assignExpr := p.visitAssignmentExpression(); assignExpr != nil {
			if p.Tokens.getCurType() == TOK_SYMBOL &&
				p.Tokens.getCurString() == ";" {
				p.Tokens.getNextToken()
				return &JumpStmtAST{assignExpr, &BaseAST{JumpStmtID}}
			}
		}
	}

	p.Tokens.applyTokenIndex(bkup)
	return nil
}
