package goexpression

import (
	"context"
	"fmt"
	"testing"
)

func TestExpression_Bool(t *testing.T) {
	exp1, err := NewExpression("b(c, d) == a && 100 in ([[b(c, d)], 1, '3', []]) || 1 in []", true, map[string]Function{
		"b": func(params ...any) (any, error) {
			if len(params) != 2 {
				return nil, fmt.Errorf("error")
			}
			// 测试类型转换不检查
			return params[0].(float64) + params[1].(float64), nil
		},
	})
	fmt.Println(err)

	exp2, err := NewExpression("age(get(ctx)) == 1", true, map[string]Function{
		"age": func(params ...any) (any, error) {
			if len(params) != 1 {
				return nil, fmt.Errorf("error")
			}
			return params[0].(float64), nil
		},
		"get": func(params ...any) (any, error) {
			if len(params) != 1 {
				return nil, fmt.Errorf("error")
			}
			ctx := params[0].(context.Context)
			return float64(ctx.Value("ctx").(int)), nil
		},
	})
	fmt.Println(err)

	exp3, err := NewExpression(`1.0 == 1.0 && 2 == 2 && true == t &&
									a == b && ++1 == --3 && -2 == (-3 + 1) && !b == False &&
									~0 == -1 && 2**2 == 4 && 1<<1 == 4 >> 1 && 2 & 1 == 0 && 2 % 3 == 2 &&
									6.0 / 3.0 == 2 && 1.5 * 2 == 3 && 3 &^ 1 == 2 && 1 ^ 2 == 3 && 1 | 2 == 3 &&
									10.5 - 2.5 == 8 && 1.2 + 2.3 == 3.5 && 2 >= 2 && 2 > 1 && 1 < 2 && 2 <= 2 &&
									3 != 2 && 0 == 0 && (1 == 2 ? true : false || 1 == 1 ? true : false)`, true, nil)
	fmt.Println(err)

	type fields struct {
		Root      *astNode
		NeedCheck bool
	}
	type args struct {
		params map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:   "TestExpression_Bool-Normal1",
			fields: fields{Root: exp1.root, NeedCheck: exp1.NeedCheck},
			args: args{map[string]any{
				"a": 100.0, // 传参数值型一律为 float64
				"c": 50.0,
				"d": 50.0,
			}},
			want:    true,
			wantErr: false,
		},
		{
			name:   "TestExpression_Bool-Normal2",
			fields: fields{Root: exp2.root, NeedCheck: exp2.NeedCheck},
			args: args{map[string]any{
				"ctx": context.WithValue(context.Background(), "ctx", 1),
			}},
			want:    true,
			wantErr: false,
		},
		{
			name:   "TestExpression_Bool-Normal3",
			fields: fields{Root: exp3.root, NeedCheck: exp3.NeedCheck},
			args: args{map[string]any{
				"a": true,
				"b": true,
			}},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Expression{
				NeedCheck: tt.fields.NeedCheck,
				root:      tt.fields.Root,
			}
			got, err := e.Bool(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bool() got = %v, want %v", got, tt.want)
			}
		})
	}
}
