package goexpression

import "fmt"

// scanner 源码扫描器, 按顺序逐字符读取源码
type scanner struct {
	Raw   string
	Index int
}

// NextChar 读取一个字符
func (s *scanner) NextChar() (rune, bool) {
	if s.Index >= len(s.Raw) {
		return 0, false
	}
	r := rune(s.Raw[s.Index])
	s.Index++
	return r, true
}

// Peek 查看当前字符
func (s *scanner) Peek() (rune, bool) {
	if s.Index >= len(s.Raw) {
		return 0, false
	}
	return rune(s.Raw[s.Index]), true
}

// Rewind 倒带step个字符
func (s *scanner) Rewind(step int) error {
	if s.Index < step || s.Index-step > len(s.Raw) {
		return fmt.Errorf("lexer: scanner Index: %d - step: %d < 0 || > len(s.Raw) %d", s.Index, step, len(s.Raw))
	}
	s.Index -= step
	return nil
}
