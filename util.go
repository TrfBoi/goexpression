package goexpression

import "unicode"

// IsNumber 是否为数字
func IsNumber(c rune) bool {
	return unicode.IsNumber(c) || c == '.'
}

// IsVar 是否是变量名
func IsVar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_'
}

// IsString 是否是字符串类型
func IsString(value any) bool {
	switch value.(type) {
	case string:
		return true
	}
	return false
}

// IsFloat64 是否是float64类型
func IsFloat64(value any) bool {
	switch value.(type) {
	case float64:
		return true
	}
	return false
}

// IsBool 是否是bool类型
func IsBool(value any) bool {
	switch value.(type) {
	case bool:
		return true
	}
	return false
}

func IsArray(value any) bool {
	switch value.(type) {
	case []any:
		return true
	}
	return false
}
