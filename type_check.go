package goexpression

// typeCheck 类型检查, 检查对应 astNode 左右子节点执行结果是否符合预期
type typeCheck func(left, right any) bool

var typeCheckArray = [OpSize]typeCheck{
	nil,
	leftBool, // ?
	nil,      // :

	logic, // ||

	logic, // &&

	eqlOrNeq, // ==
	eqlOrNeq, // !=
	canCmp,   // <
	canCmp,   // <=
	canCmp,   // >
	canCmp,   // >=
	nil,      // in

	canCmp,    // +
	isFloat64, // -
	isFloat64, // |
	isFloat64, // ^

	isFloat64, // *
	isFloat64, // /
	isFloat64, // %
	isFloat64, // &
	isFloat64, // &^
	isFloat64, // <<
	isFloat64, // >>

	isFloat64, // **

	leftFloat64RightNil, // ++, 注意++和--只设计成只可前置
	leftFloat64RightNil, // --
	leftFloat64RightNil, // -
	leftBoolRightNil,    // !
	leftFloat64RightNil, // ~
}

func canCmp(left, right any) bool {
	return IsFloat64(left) && IsFloat64(right) || IsString(left) && IsString(right)
}

func eqlOrNeq(left, right any) bool {
	return IsBool(left) && IsBool(right) || canCmp(left, right)
}

func logic(left, right any) bool {
	return IsBool(left) && IsBool(right)
}

func isFloat64(left, right any) bool {
	return IsFloat64(left) && IsFloat64(right)
}

func leftFloat64RightNil(left, right any) bool {
	return IsFloat64(left) && right == nil
}

func leftBoolRightNil(left, right any) bool {
	return IsBool(left) && right == nil
}

func leftBool(left, _ any) bool {
	return IsBool(left)
}
