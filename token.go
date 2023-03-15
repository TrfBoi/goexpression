package goexpression

import "fmt"

// TokenKind token类型
type TokenKind int

const (
	FloatLit TokenKind = iota
	StrLit
	BoolLit
	Var // var

	Lparen // (
	Rparen // )
	Lbrack // [
	Rbrack // ]
	Comma  // ,

	Op

	Func
)

// Operator 操作符
type Operator int

// 按优先级由低到高排列
const (
	NotOperator Operator = iota

	TernaryT // ?
	TernaryF // :

	OrOr // ||

	AndAnd // &&

	Eql // ==
	Neq // !=
	Lss // <
	Leq // <=
	Gtr // >
	Geq // >=
	In  // in

	Add // +
	Sub // -
	Or  // |
	Xor // ^

	Mul    // *
	Div    // /
	Rem    // %
	And    // &
	AndNot // &^
	Shl    // <<
	Shr    // >>

	Exponent // **

	AddAdd // ++, 注意++和--只设计成只可前置
	SubSub // --
	Minus  // -
	Not    // !
	BitNot // ~
	OpSize
)

// IsOperator 是否为正确表达式
func (o Operator) IsOperator() bool {
	return o > NotOperator && o < OpSize
}

// IsBinaryOperator 是否为二元表达式
func (o Operator) IsBinaryOperator() bool {
	return o >= TernaryT && o <= Exponent
}

var opMap = map[string]Operator{
	"?":  TernaryT,
	":":  TernaryF,
	"||": OrOr,
	"&&": AndAnd,
	"==": Eql,
	"!=": Neq,
	"<":  Lss,
	"<=": Leq,
	">":  Gtr,
	">=": Geq,
	"in": In,
	"+":  Add,
	"-":  Sub,
	"|":  Or,
	"^":  Xor,
	"*":  Mul,
	"/":  Div,
	"%":  Rem,
	"&":  And,
	"&^": AndNot,
	"<<": Shl,
	">>": Shr,
	"**": Exponent,
	"++": AddAdd,
	"--": SubSub,
	"!":  Not,
	"~":  BitNot,
	// "-": Minus, // 特殊处理
}

// var opPrec = [OpSize]int{0, 1, 1, 2, 3, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 7, 8, 8, 8, 8, 8}

// GetPrec 获取操作符优先级
func (o Operator) GetPrec() int {
	switch o {
	case TernaryT:
		fallthrough
	case TernaryF:
		return 1
	case OrOr:
		return 2
	case AndAnd:
		return 3
	case Eql:
		fallthrough
	case Neq:
		fallthrough
	case Lss:
		fallthrough
	case Leq:
		fallthrough
	case Gtr:
		fallthrough
	case Geq:
		fallthrough
	case In:
		return 4
	case Add:
		fallthrough
	case Sub:
		fallthrough
	case Or:
		fallthrough
	case Xor:
		return 5
	case Mul:
		fallthrough
	case Div:
		fallthrough
	case Rem:
		fallthrough
	case And:
		fallthrough
	case AndNot:
		fallthrough
	case Shl:
		fallthrough
	case Shr:
		return 6
	case Exponent:
		return 7
	case AddAdd:
		fallthrough
	case SubSub:
		fallthrough
	case Minus:
		fallthrough
	case Not:
		fallthrough
	case BitNot:
		return 8
	default:
		return 0
	}
}

// Token Token
type Token struct {
	Type     TokenKind
	Operator Operator
	Raw      any
}

// String return Token's string representation
func (t *Token) String() string {
	return fmt.Sprintf("{Type: %d, Raw: %v, Operator: %d}", t.Type, t.Raw, t.Operator)
}

var stateTransferMap = map[TokenKind]map[TokenKind]bool{
	FloatLit: {
		Op:     true, // 1 + 1
		Rparen: true, // (1 + 1)
		Rbrack: true, // [1, 2]
		Comma:  true, // [1, 1]
	},
	StrLit: {
		Op:     true, // '1' + '1'
		Rparen: true,
		Rbrack: true,
		Comma:  true,
	},
	BoolLit: {
		Op:     true, // true == false
		Rparen: true,
		Rbrack: true,
		Comma:  true,
	},
	Var: {
		Op:     true, // a == b
		Rparen: true,
		Rbrack: true,
		Comma:  true,
	},
	Lparen: {
		Op:       true, // (!a & b)
		FloatLit: true, // (1 + 1)
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Lparen:   true, // ((1+1)+1)
		Rparen:   true, // func()
		Lbrack:   true, // ([1, 2])
		Func:     true,
	},
	Rparen: {
		Op:     true, // (1 + 1) + 1
		Rparen: true, // (1 + (1 + 1))
		Rbrack: true,
		Comma:  true, // [(1+1), (2+1)]
	},
	Lbrack: {
		Op:       true, // [-1, 2]
		FloatLit: true, // [1, 2]
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Lparen:   true, // [(1+1), 1]
		Lbrack:   true, // [[1],2]
		Rbrack:   true, // []
		Func:     true, // [func(), 2]
	},
	Rbrack: {
		Op:     true, // 1 in [1, 2, 3] || a > b
		Rparen: true, // (1 in [1, 2, 3])
		Rbrack: true, // [2, [1]]
		Comma:  true, // [[1],2]
	},
	Comma: {
		Op:       true, // [-1, 2]
		FloatLit: true, // [1, 2]
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Lparen:   true, // [1, (1+1)]
		Lbrack:   true, //  [2, [1]]
		Func:     true,
	},
	Op: {
		Op:       true, // 1 + -1
		FloatLit: true,
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Lparen:   true,
		Lbrack:   true,
		Func:     true,
	},
	Func: {
		Lparen: true, // func()
	},
}

// GotTokenKinds 当前Token后面期望的Token类型
func (t *Token) GotTokenKinds() map[TokenKind]bool {
	return stateTransferMap[t.Type]
}

var (
	startTokens = map[TokenKind]bool{
		FloatLit: true,
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Lparen:   true,
		Lbrack:   true,
		Op:       true,
		Func:     true,
	}
	endTokens = map[TokenKind]bool{
		FloatLit: true,
		StrLit:   true,
		BoolLit:  true,
		Var:      true,
		Rparen:   true,
		Rbrack:   true,
	}
)

// CanStart 判断当前Token是否可以作为表达式开头
func (t *Token) CanStart() bool { return startTokens[t.Type] }

// CanEnd 判断当前Token是否可以作为表达式结尾
func (t *Token) CanEnd() bool { return endTokens[t.Type] }
