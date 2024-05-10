package main

import (
	"testing"
)

func TestEval(t *testing.T) {
	e := NewEvaluator()

	for _, test := range []struct {
		input string
		want  float64
	}{
		{
			"def foo(a b) a+b",
			0,
		},
		{
			"foo(3,5)",
			8.0,
		},
		{
			"foo(3,8)",
			11.0,
		},
		{
			"def bar(c d) foo(c,d) + d",
			0,
		},
		{
			"foo(3,5) + bar(3,5);",
			21.0,
		},
		{
			"extern floor(x)",
			0,
		},
		{
			"floor(23.8)",
			23.0,
		},
	} {
		got := e.Evaluate(test.input)
		t.Logf("eval `%s` = %.2f, want %.2f", test.input, got, test.want)
		if test.want != got {
			t.Errorf("eval `%s` = %.2f, want %.2f", test.input, got, test.want)
		}
	}
}
