package goexpression

import "fmt"

// astNode 抽象语法树节点
type astNode struct {
	left, right *astNode
	op          Operator
	opFunc      opFunc
	typeCheck   typeCheck
}

func (a *astNode) dumpASTNode() {
	if a == nil {
		return
	}
	a.left.dumpASTNode()
	fmt.Printf("%+v\n", a)
	a.right.dumpASTNode()
}

// parse 语法分析器, 非并发安全, 不可重复利用, 只能解析一个表达式
type parse struct {
	*lexer
	root     *astNode
	curIndex int
}

// newParse 创建Parse
func newParse(raw string) *parse {
	return &parse{
		lexer: newLexer(raw),
		root:  &astNode{},
	}
}

func (p *parse) curToken() *Token {
	if len(p.Tokens) > p.curIndex {
		t := p.Tokens[p.curIndex]
		return t
	}
	return nil
}
func (p *parse) next() {
	p.curIndex++
}

func (p *parse) end() bool {
	return p.curIndex >= len(p.Tokens)
}

func (p *parse) pre() *Token {
	if p.curIndex == 0 || p.end() {
		return nil
	}
	return p.Tokens[p.curIndex-1]
}

// OnceParse 语法分析, 表达式只需要一次分析
func (p *parse) OnceParse(functions map[string]Function) (*astNode, error) {
	if err := p.Parse(functions); err != nil {
		return nil, err
	}

	// 提前检查
	if err := p.advanceCheck(); err != nil {
		return nil, err
	}
	// 语法检查
	if err := p.syntaxCheck(); err != nil {
		return nil, err
	}

	if err := p.doOnceParse(); err != nil {
		return nil, err
	}
	// 提前类型检查
	// 检查如 1 + true 这种错误
	// 暂时也不实现(错误将延时在运行时暴露), 应该由用户自行保证这些低级错误不会写出来
	//if err := p.advanceTypeCheck(); err != nil {
	//	return nil, err
	//}
	// 优化, 暂时不实现, 用户应该写出不需要优化的表达式
	// 如 1 + 2 + a 就是需要优化的表达式, 用户不应该写出这种表达式
	// p.optimization()

	return p.root, nil
}

func (p *parse) advanceCheck() error {
	var lParenNum, rParenNum, lBrack, rBrack int
	for _, token := range p.Tokens {
		if token.Type == Lparen {
			lParenNum++
		}
		if token.Type == Rparen {
			rParenNum++
		}
		if token.Type == Lbrack {
			lBrack++
		}
		if token.Type == Rbrack {
			rBrack++
		}
	}
	if lParenNum != rParenNum {
		return fmt.Errorf("syntax: ( num != ), please check")
	}
	if lBrack != rBrack {
		return fmt.Errorf("syntax: [ num != ], please check")
	}
	return nil
}

func (p *parse) syntaxCheck() error {
	length := len(p.Tokens)
	for i := 0; i < length-1; i++ {
		if !p.Tokens[i].GotTokenKinds()[p.Tokens[i+1].Type] {
			return fmt.Errorf("syntax: illegal %v after %v", p.Tokens[i+1].Raw, p.Tokens[i].Raw)
		}
	}
	if length > 0 && !p.Tokens[0].CanStart() {
		return fmt.Errorf("syntax: %v can't start as an expression", p.Tokens[0].Raw)
	}
	if length > 0 && !p.Tokens[length-1].CanEnd() {
		return fmt.Errorf("syntax: %v can't end as an expression", p.Tokens[length-1].Raw)
	}
	return nil
}

//func (p *parse) advanceTypeCheck() error {
//	return nil
//}

func (p *parse) doOnceParse() error {
	var err error
	p.root, err = p.binaryExpr(nil, 0)
	if err != nil {
		return err
	}
	if !p.end() {
		return fmt.Errorf("syntax: %v and it after tokens is illegal", p.curToken())
	}
	return nil
}

// func (p *parse) optimization() {}

// binaryExprs 解析表达式列表, 表达式用 ',' 分割
func (p *parse) binaryExprs(end *Token) (*astNode, error) {
	var (
		err  error
		left *astNode
	)
	if !p.end() && p.curToken().Type == end.Type {
		return nil, nil
	}
	left, err = p.binaryExpr(nil, 0)
	if err != nil {
		return nil, err
	}
	for !p.end() && p.curToken().Type == Comma {
		parent := &astNode{
			opFunc: commaFunc,
			// Type Check 由 func/in node 完成
		}
		p.next() // ,
		parent.left = left
		parent.right, err = p.binaryExpr(nil, 0)
		if err != nil {
			return nil, err
		}
		left = parent
	}
	return left, nil
}

// 优先级问题: 只会遍历大于传入优先级的操作符, 让大于传入优先级的成为一棵树, 且是子树, 所以先执行
// 结合性问题: 当有连续优先级相同操作符时, 永远是已经解析的成为新解析Node的左结点
// 左递归问题: binaryExpr 先调用 unaryExpr 当发现不仅仅只有 unaryExpr 时, 不是回溯, 而是 unaryExpr 成为 binaryExpr 的一部分
func (p *parse) binaryExpr(left *astNode, prec int) (*astNode, error) {
	var err error
	if left == nil {
		left, err = p.unaryExpr()
	}
	if err != nil {
		return nil, err
	}

	for curToken := p.curToken(); curToken != nil && curToken.Operator.IsBinaryOperator() &&
		curToken.Operator.GetPrec() > prec; curToken = p.curToken() {
		parent := &astNode{
			opFunc:    opFuncArray[curToken.Operator],
			typeCheck: typeCheckArray[curToken.Operator],
			op:        curToken.Operator,
		}
		curPrec := curToken.Operator.GetPrec()
		p.next() // op
		parent.left = left
		parent.right, err = p.binaryExpr(nil, curPrec)
		if err != nil {
			return nil, err
		}
		left = parent
	}
	return left, nil
}

// unaryExpr 解析可能携带一元操作符的表达式
// unaryExpr = primaryExpr | unary_op unaryExpr
func (p *parse) unaryExpr() (*astNode, error) {
	if p.end() {
		return nil, fmt.Errorf("syntax: need unaryExpr, expression premature end")
	}
	var (
		curToken = p.curToken()
		err      error
	)

	if curToken.Operator.IsOperator() {
		switch curToken.Operator {
		case AddAdd, SubSub, Not, BitNot:
			parent := &astNode{
				opFunc:    opFuncArray[curToken.Operator],
				typeCheck: typeCheckArray[curToken.Operator],
				op:        curToken.Operator,
			}
			p.next()
			parent.left, err = p.unaryExpr()
			return parent, err
		case Sub: // Sub == Minus
			parent := &astNode{
				opFunc:    opFuncArray[Minus],
				typeCheck: typeCheckArray[Minus],
				op:        Minus,
			}
			p.next()
			parent.left, err = p.unaryExpr()
			return parent, err
		default:
			return nil, fmt.Errorf("syntax: parse unaryExpr illegal operator %v", curToken.Raw)
		}
	}

	return p.primaryExpr()
}

// primaryExpr 基本表达式
// primaryExpr = Lit | Var | Func | ( binaryExpr ) | (unaryExpr, unaryExpr...) | [ unaryExpr, unaryExpr.... ]
func (p *parse) primaryExpr() (*astNode, error) {
	if p.end() {
		return nil, fmt.Errorf("syntax: need primaryExpr, expression premature end")
	}
	var (
		curToken = p.curToken()
		ret      = &astNode{}
		err      error
	)

	switch curToken.Type {
	case BoolLit, StrLit, FloatLit:
		ret.opFunc = makeLitFunc(curToken.Raw)
		p.next() // Lit
		return ret, nil
	case Var:
		ret.opFunc = makeVarFunc(curToken.Raw.(string))
		p.next() // var
		return ret, nil
	case Func:
		ret.opFunc = makeFuncFunc(curToken.Raw.(Function))
		p.next() // func name
		// 虽然已经在状态转移检查中做过了, 但是为了保证语法解析完整性, 随时可以去掉状态检查, 状态转移只是提前检查
		if p.end() || p.curToken().Type != Lparen {
			return nil, fmt.Errorf("syntax: func after need ( ")
		}
		p.next() // (
		ret.right, err = p.binaryExprs(&Token{Type: Rparen})
		if err != nil {
			return nil, err
		}
		if p.end() || p.curToken().Type != Rparen {
			return nil, fmt.Errorf("syntax: ( lack of ) ")
		}
		p.next() // )
		return ret, err
	case Lparen:
		p.next() // (
		ret, err = p.binaryExpr(nil, 0)
		if err != nil {
			return nil, err
		}
		if p.end() || p.curToken().Type != Rparen {
			return nil, fmt.Errorf("syntax: ( lack of ) ")
		}
		p.next() // )
		return ret, nil
	case Lbrack:
		p.next() // [
		ret, err = p.binaryExprs(&Token{Type: Rbrack})
		if err != nil {
			return nil, err
		}
		if p.end() || p.curToken().Type != Rbrack {
			return nil, fmt.Errorf("syntax: [ lack of ] ")
		}
		p.next() // ]
		return ret, nil
	default:
		return nil, fmt.Errorf("syntax: primaryExpr illegal Token %v", curToken)
	}
}
