package goexpression

import (
	"fmt"
)

// Expression expression
type Expression struct {
	root      *astNode
	NeedCheck bool
}

// Execute 执行表达式
func (e *Expression) Execute(params map[string]any) (any, error) {
	if e.root == nil {
		return nil, fmt.Errorf("execute: parse result is nil")
	}
	return e.executeASTNode(e.root, params)
}

func (e *Expression) executeASTNode(root *astNode, params map[string]any) (any, error) {
	if root == nil {
		return nil, nil
	}
	var (
		left, right any
		err         error
	)

	if left, err = e.executeASTNode(root.left, params); err != nil {
		return nil, err
	}

	// 左边结果可能可以直接决定结果的
	switch root.op {
	case AndAnd:
		if left == false {
			return false, nil
		}
	case OrOr:
		if left == true {
			return true, nil
		}
	case TernaryT:
		if left == false {
			return nil, nil
		}
	case TernaryF:
		if left != nil {
			return left, nil
		}
	default:
	}

	if right, err = e.executeASTNode(root.right, params); err != nil {
		return nil, err
	}

	if e.NeedCheck && root.typeCheck != nil {
		if !root.typeCheck(left, right) {
			return nil, fmt.Errorf("execute: type check error")
		}
	}
	return root.opFunc(left, right, params)
}

// Bool 计算bool结果
func (e *Expression) Bool(params map[string]any) (bool, error) {
	ret, err := e.Execute(params)
	if err != nil {
		return false, err
	}
	b, ok := ret.(bool)
	if !ok {
		return false, fmt.Errorf("execute: the result( %+v ) is not of bool type", ret)
	}
	return b, nil
}

// Str returns str value
func (e *Expression) Str(params map[string]any) (string, error) {
	ret, err := e.Execute(params)
	if err != nil {
		return "", err
	}
	s, ok := ret.(string)
	if !ok {
		return "", fmt.Errorf("execute: the result( %+v ) is not of string type", ret)
	}
	return s, nil
}

// Int64 returns int64 value
func (e *Expression) Int64(params map[string]any) (int64, error) {
	ret, err := e.Execute(params)
	if err != nil {
		return 0, err
	}
	num, ok := ret.(float64)
	if !ok {
		return 0, fmt.Errorf("execute: the result( %+v ) is not of int64 type", ret)
	}
	return int64(num), nil
}

// Float64 returns float64 value
func (e *Expression) Float64(params map[string]any) (float64, error) {
	ret, err := e.Execute(params)
	if err != nil {
		return 0.0, err
	}
	num, ok := ret.(float64)
	if !ok {
		return 0.0, fmt.Errorf("execute: the result( %+v ) is not of float64 type", ret)
	}
	return num, nil
}

// NewExpression creates a new expression
func NewExpression(exp string, needCheck bool, functions map[string]Function) (*Expression, error) {
	var (
		p          = newParse(exp)
		expression = &Expression{NeedCheck: needCheck}
		err        error
	)
	expression.root, err = p.OnceParse(functions)
	return expression, err
}
