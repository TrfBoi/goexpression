package goexpression

import (
	"fmt"
	"testing"
)

func TestLexer_Parse(t *testing.T) {
	type fields struct {
		srcScanner *scanner
		Tokens     []*Token
	}
	type args struct {
		functions map[string]Function
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestLexer_Parse-NormalAll",
			fields: fields{
				srcScanner: &scanner{
					Raw: "1 2.0 '3' true false a b ()[], ? : || && == != < <= > >= in + - | ^ * / % & &^ << >> ** ++ -- - ! ~",
				},
			},
			args: args{
				functions: map[string]Function{
					"b": nil,
				},
			},
			wantErr: false,
		},
		{
			name: "TestLexer_Parse-Normal1",
			fields: fields{
				srcScanner: &scanner{
					Raw: "1.0 > 0.2 + .3 || a=='11111\\'' && true != false || b &^ 1",
				},
			},
			args: args{
				functions: map[string]Function{
					"b": nil,
				},
			},
			wantErr: false,
		},
		{
			name: "TestLexer_Parse-Normal2",
			fields: fields{
				srcScanner: &scanner{
					Raw: "1++1 ++2 == 3 % 4 && b(c, d) == a && 1 in [1, 2.0, '3']",
				},
			},
			args: args{
				functions: map[string]Function{
					"b": nil,
				},
			},
			wantErr: false,
		},
		{
			name: "TestLexer_Parse-error",
			fields: fields{
				srcScanner: &scanner{
					Raw: "+ +++1 =*= 2-- || 1 | 2 && 3 & 4 || b(1, 2) && 2 in [3, 3, 3]",
				},
			},
			args: args{
				functions: map[string]Function{
					"b": nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				scanner: tt.fields.srcScanner,
				Tokens:  tt.fields.Tokens,
			}
			err := l.Parse(tt.args.functions)
			fmt.Printf("%v\n", l.Tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
