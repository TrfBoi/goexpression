package goexpression

import (
	"fmt"
	"math"
)

// opFunc 执行函数格式定义
type opFunc func(l any, r any, params map[string]any) (any, error)

var opFuncArray = [OpSize]opFunc{
	nil,
	ternaryTFunc, // ?
	ternaryFFunc, // :

	orOrFunc, // ||

	andAndFunc, // &&

	eqlFunc, // ==
	neqFunc, // !=
	lssFunc, // <
	leqFunc, // <=
	gtrFunc, // >
	geqFunc, // >=
	inFunc,  // in

	addFunc, // +
	subFunc, // -
	orFunc,  // |
	xorFunc, // ^

	mulFunc,    // *
	divFunc,    // /
	remFunc,    // %
	andFunc,    // &
	andNotFunc, // &^
	shlFunc,    // <<
	shrFunc,    // >>

	exponentFunc, // **

	addAddFunc, // ++, 注意++和--只设计成只可前置
	subSubFunc, // --
	minusFunc,  // -
	notFunc,    // !
	bitNotFunc, // ~
}

// a ? b : c
// a == true 时 return right(b)
func ternaryTFunc(left, right any, _ map[string]any) (any, error) {
	if left.(bool) {
		return right, nil
	}
	return nil, nil
}

// left == nil 即 a == false。return right(c)
func ternaryFFunc(left, right any, _ map[string]any) (any, error) {
	if left != nil {
		return left, nil
	}
	return right, nil
}

func orOrFunc(left, right any, _ map[string]any) (any, error) {
	return left.(bool) || right.(bool), nil
}

func andAndFunc(left, right any, _ map[string]any) (any, error) {
	return left.(bool) && right.(bool), nil
}

// 只支持基本类型
func eqlFunc(left, right any, _ map[string]any) (any, error) {
	return left == right, nil
}

func neqFunc(left, right any, _ map[string]any) (any, error) {
	return left != right, nil
}

// str float64
func lssFunc(left, right any, _ map[string]any) (any, error) {
	if IsString(left) && IsString(right) {
		return left.(string) < right.(string), nil
	}
	return left.(float64) < right.(float64), nil
}

func leqFunc(left, right any, _ map[string]any) (any, error) {
	if IsString(left) && IsString(right) {
		return left.(string) <= right.(string), nil
	}
	return left.(float64) <= right.(float64), nil
}

func gtrFunc(left, right any, _ map[string]any) (any, error) {
	if IsString(left) && IsString(right) {
		return left.(string) > right.(string), nil
	}
	return left.(float64) > right.(float64), nil
}

func geqFunc(left, right any, _ map[string]any) (any, error) {
	if IsString(left) && IsString(right) {
		return left.(string) >= right.(string), nil
	}
	return left.(float64) >= right.(float64), nil
}

func inFunc(left, right any, _ map[string]any) (any, error) {
	if right == nil {
		return false, nil
	}
	switch right.(type) {
	case []any:
		for _, v := range right.([]any) {
			if v == left {
				return true, nil
			}
		}
		return false, nil
	default:
		return left == right, nil
	}
}

func addFunc(left, right any, _ map[string]any) (any, error) {
	if IsString(left) && IsString(right) {
		return left.(string) + right.(string), nil
	}
	return left.(float64) + right.(float64), nil
}

func subFunc(left, right any, _ map[string]any) (any, error) {
	return left.(float64) - right.(float64), nil
}

func orFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) | int64(right.(float64))), nil
}

func xorFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) ^ int64(right.(float64))), nil
}

func mulFunc(left, right any, _ map[string]any) (any, error) {
	return left.(float64) * right.(float64), nil
}

func divFunc(left, right any, _ map[string]any) (any, error) {
	return left.(float64) / right.(float64), nil
}

func remFunc(left, right any, _ map[string]any) (any, error) {
	return math.Mod(left.(float64), right.(float64)), nil
}

func andFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) & int64(right.(float64))), nil
}

func andNotFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) &^ int64(right.(float64))), nil
}

func shlFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) << int64(right.(float64))), nil
}

func shrFunc(left, right any, _ map[string]any) (any, error) {
	return float64(int64(left.(float64)) >> int64(right.(float64))), nil
}

func exponentFunc(left, right any, _ map[string]any) (any, error) {
	return math.Pow(left.(float64), right.(float64)), nil
}

func addAddFunc(left, _ any, _ map[string]any) (any, error) {
	return left.(float64) + 1, nil
}

func subSubFunc(left, _ any, _ map[string]any) (any, error) {
	return left.(float64) - 1, nil
}

func minusFunc(left, _ any, _ map[string]any) (any, error) {
	return -left.(float64), nil
}

func notFunc(left, _ any, _ map[string]any) (any, error) {
	return !left.(bool), nil
}

func bitNotFunc(left, _ any, _ map[string]any) (any, error) {
	return float64(^int64(left.(float64))), nil
}

func makeLitFunc(lit any) opFunc {
	return func(l any, r any, params map[string]any) (any, error) {
		return lit, nil
	}
}

func makeVarFunc(name string) opFunc {
	return func(l, r any, params map[string]any) (any, error) {
		ret, ok := params[name]
		if !ok {
			return nil, fmt.Errorf("execute: %s param not in the passed parameter list", name)
		}
		return ret, nil
	}
}

func makeFuncFunc(function Function) opFunc {
	return func(left, right any, params map[string]any) (any, error) {
		if right == nil {
			return function()
		}
		switch right.(type) {
		case []any:
			return function(right.([]any)...)
		default:
			return function(right)
		}
	}
}

func commaFunc(left, right any, _ map[string]any) (any, error) {
	var ret []any

	switch left.(type) {
	case []any:
		ret = append(left.([]any), right)
	default:
		ret = []any{left, right}
	}
	return ret, nil
}
