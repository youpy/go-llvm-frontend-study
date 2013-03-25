package frontend

type AstID int

const (
	BaseID         AstID = 0
	VariableDeclID AstID = 1
	NullExprID     AstID = 2
	VariableID     AstID = 3
	BinaryExprID   AstID = 4
	CallExprID     AstID = 5
	NumberID       AstID = 6
	JumpStmtID     AstID = 7
)

type DeclType int

const (
	Decl_local DeclType = 0
	Decl_param DeclType = 1
)

type AST interface {
	GetID() AstID
}

func classOf(a AST, b AST) bool {
	return a.GetID() == b.GetID()
}

type BaseAST struct {
	ID AstID
}

func (b *BaseAST) GetID() AstID {
	return b.ID
}

type VariableAST struct {
	Name string
	*BaseAST
}

type NumberAST struct {
	Val int
	*BaseAST
}

type BinaryExprAST struct {
	Op  string
	LHS AST
	RHS AST
	*BaseAST
}

type CallExprAST struct {
	Callee string
	Args   []AST
	*BaseAST
}

type VariableDeclAST struct {
	Name string
	Type DeclType
	*BaseAST
}

type JumpStmtAST struct {
	Expr AST
	*BaseAST
}

type FunctionStmtAST struct {
	VariableDecls []*VariableDeclAST
	StmtLists     []AST
}

type PrototypeAST struct {
	Name   string
	Params []string
}

type FunctionAST struct {
	Proto *PrototypeAST
	Body  *FunctionStmtAST
}

type TranslationUnitAST struct {
	Prototypes []*PrototypeAST
	Functions  []*FunctionAST
}

type NullExprAST struct {
	*BaseAST
}

func (ast *NullExprAST) GetID() AstID {
	return ast.BaseAST.ID
}

// func (t *TranslationUnitAST) empty() bool {
// }
