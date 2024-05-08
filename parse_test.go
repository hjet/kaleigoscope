package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func treeString(e ExprAST) string {
	var buf bytes.Buffer
	writeTree(&buf, reflect.ValueOf(e))
	return buf.String()
}

func writeTree(buf *bytes.Buffer, x reflect.Value) {
	switch x.Kind() {
	case reflect.String, reflect.Int, reflect.Bool:
		fmt.Fprintf(buf, "%v", x.Interface())
	case reflect.Int32: // rune hack
		fmt.Fprintf(buf, "%c", x.Interface())
	case reflect.Ptr, reflect.Interface:
		if elem := x.Elem(); elem.Kind() == 0 {
			buf.WriteString("nil")
		} else {
			writeTree(buf, elem)
		}
	case reflect.Struct:
		fmt.Fprintf(buf, "(%s", strings.TrimPrefix(x.Type().String(), "main."))
		for i, n := 0, x.NumField(); i < n; i++ {
			f := x.Field(i)
			fName := x.Type().Field(i).Name
			switch f.Kind() {
			case reflect.Slice:
				if n := f.Len(); n > 0 {
					fmt.Fprintf(buf, " %s=(", fName)
					for i := 0; i < n; i++ {
						if i > 0 {
							buf.WriteByte(' ')
						}
						writeTree(buf, f.Index(i))
					}
					fmt.Fprintf(buf, ")")
				}
				continue
			case reflect.Ptr, reflect.Interface:
				if f.IsNil() {
					continue
				}
			}
			fmt.Fprintf(buf, " %s=", fName)
			writeTree(buf, f)
		}
		fmt.Fprintf(buf, ")")
	default:
		fmt.Fprintf(buf, "%T", x.Interface())
	}
}

func TestParse(t *testing.T) {
	for _, test := range []struct {
		input, want string
	}{
		{`print(1)`, `(FunctionAST Proto=(PrototypeAST Name=__anon) Body=(CallExprAST Callee=print Args=((NumberExprAST Val=1))))`},
	} {
		p := &Parser{}
		p.Init([]byte(test.input))
		e := p.ParseTopLevel()
		got := treeString(e)
		if test.want != got {
			t.Errorf("parse `%s` = %s, want %s", test.input, got, test.want)
		}

	}
}

//func main() {
//	test := "def foo(a b) b + foo(a,3)"
//	p := &Parser{}
//	p.Init([]byte(test))
//	res := p.ParseTopLevel()
//	fmt.Println(treeString(res))
//}
