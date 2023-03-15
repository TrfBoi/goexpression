package goexpression

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// lexer 表达式词法分析器
type lexer struct {
	*scanner
	Tokens []*Token
}

// newLexer creates a new lexer
func newLexer(src string) *lexer {
	return &lexer{
		scanner: &scanner{
			Raw: src,
		},
	}
}

// Parse 词法解析
// 注意有相同前缀的 Token 的解析优先级
// eg: 1 ++ 2 不会被解析为 '1' '+' '+' '2'
func (l *lexer) Parse(functions map[string]Function) error {
	if len(l.Raw) == 0 {
		return fmt.Errorf("lexer: expression is empty")
	}
	var err error
	for char, hasNext := l.NextChar(); hasNext; char, hasNext = l.NextChar() {
		if unicode.IsSpace(char) {
			continue
		}
		if unicode.IsLetter(char) {
			name := l.letters(char)
			if ok := l.isKeyLetter(name); ok { // 关键字优先级最大
				continue
			}
			if f, ok := functions[name]; ok { // 注册的函数
				l.addToken(f, Func, false)
				continue
			}
			l.addToken(name, Var, false)
			continue
		}
		switch char {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			err = l.number(char)
		case '|':
			l.double('|')
		case '&':
			l.and()
		case '=':
			cur, ok := l.NextChar()
			if !ok || cur != '=' {
				return fmt.Errorf("lexer: index: %d need to be '=' ", l.Index-1)
			}
			l.addToken("==", Op, true)
		case '!':
			l.not()
		case '<':
			l.less()
		case '>':
			l.greater()
		case '(':
			l.addToken("(", Lparen, false)
		case ')':
			l.addToken(")", Rparen, false)
		case '+':
			l.double('+')
		case '-':
			// 负号逻辑在语法分析中区分
			l.double('-')
		case '^':
			l.addToken("^", Op, true)
		case '*':
			l.double('*')
		case '/':
			l.addToken("/", Op, true)
		case '%':
			l.addToken("%", Op, true)
		case '?':
			l.addToken("?", Op, true)
		case ':':
			l.addToken(":", Op, true)
		case '~':
			l.addToken("~", Op, true)
		case '[':
			l.addToken("[", Lbrack, false)
		case ']':
			l.addToken("]", Rbrack, false)
		case ',':
			l.addToken(",", Comma, false)
		case '\'':
			err = l.stdStr('\'')
		default:
			return fmt.Errorf("lexer: index: %d character illegal", l.Index-1)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lexer) addToken(raw any, typ TokenKind, isOp bool) {
	if isOp {
		op, ok := raw.(string)
		if !ok {
			panic(fmt.Sprintf("invalid operation: %v", raw))
		}
		l.Tokens = append(l.Tokens, &Token{Raw: raw, Type: Op, Operator: opMap[op]})
		return
	}
	l.Tokens = append(l.Tokens, &Token{Raw: raw, Type: typ})
}

func (l *lexer) stdStr(end rune) error {
	curIndex := l.Index
	builder := strings.Builder{}
	for char, hasNext := l.NextChar(); hasNext; char, hasNext = l.NextChar() {
		if char == '\\' {
			char, hasNext = l.NextChar()
			if !hasNext {
				return fmt.Errorf("lexer: index %d \\\\ after is end, need '", l.Index-2)
			}
			if char != end {
				return fmt.Errorf("lexer: index %d \\\\ after char need '", l.Index-2)
			}
			builder.WriteRune(char)
			continue
		}
		if char == end {
			l.addToken(builder.String(), StrLit, false)
			return nil
		}
		builder.WriteRune(char)
	}
	return fmt.Errorf("lexer: index %d \" missing right ' or \"", curIndex-1)
}

// TODO(bioit): 当有指数/进制等需求时改进
func (l *lexer) number(start rune) error {
	builder := strings.Builder{}
	builder.WriteRune(start)
	for char, ok := l.NextChar(); ok; char, ok = l.NextChar() {
		if !IsNumber(char) {
			_ = l.Rewind(1)
			break
		}
		builder.WriteRune(char)
	}
	numStr := builder.String()
	num, err := strconv.ParseFloat(numStr, 10)
	if err != nil {
		return fmt.Errorf("lexer: %s is not a num %+v", numStr, err)
	}
	l.addToken(num, FloatLit, false)
	return nil
}

func (l *lexer) letters(start rune) string {
	builder := strings.Builder{}
	builder.WriteRune(start)
	for char, ok := l.NextChar(); ok; char, ok = l.NextChar() {
		if !IsVar(char) {
			_ = l.Rewind(1)
			break
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

func (l *lexer) isKeyLetter(name string) bool {
	low := strings.ToLower(name)
	if low == "true" || low == "t" {
		l.addToken(true, BoolLit, false)
		return true
	}
	if low == "false" || low == "f" {
		l.addToken(false, BoolLit, false)
		return true
	}
	if low == "in" {
		l.addToken(low, Op, true)
		return true
	}
	return false
}

func (l *lexer) and() {
	cur, ok := l.Peek()
	if ok && cur == '^' {
		_, _ = l.NextChar()
		l.addToken("&^", Op, true)
		return
	}
	l.double('&')
}

func (l *lexer) not() {
	if cur, ok := l.Peek(); ok && cur == '=' {
		_, _ = l.NextChar()
		l.addToken("!=", Op, true)
	} else {
		l.addToken("!", Op, true)
	}
}

func (l *lexer) less() {
	cur, ok := l.Peek()
	if ok && cur == '=' {
		_, _ = l.NextChar()
		l.addToken("<=", Op, true)
		return
	}
	l.double('<')
}

func (l *lexer) greater() {
	cur, ok := l.Peek()
	if ok && cur == '=' {
		_, _ = l.NextChar()
		l.addToken(">=", Op, true)
		return
	}
	l.double('>')
}

func (l *lexer) double(c rune) {
	if cur, ok := l.Peek(); ok && cur == c {
		_, _ = l.NextChar()
		l.addToken(string(c)+string(c), Op, true)
	} else {
		l.addToken(string(c), Op, true)
	}
}
