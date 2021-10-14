package fix

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

var fixTests = []struct {
	name, before, after string
}{
	{
		"new variable",
		"a = foo",
		"var a = foo",
	},
	{
		"new variable in lambda",
		"{ a = foo }",
		"{ var a = foo }",
	},
	{
		"set local",
		"var a = foo; a = bar",
		"var a = foo; set a = bar",
	},
	{
		"set explicit local",
		"var a = foo; local:a = bar",
		"var a = foo; set local:a = bar",
	},
	{
		"mix of var and set",
		"var a = foo; a b = x y",
		"var a = foo; var b; set a b = x y",
	},
	{
		"set builtin",
		"value-out-indicator = x",
		"set value-out-indicator = x",
	},
	{
		"set captured",
		"var a; { a = foo }",
		"var a; { set a = foo }",
	},
	{
		"set explicit upvalue",
		"var a; { up:a = foo }",
		"var a; { set up:a = foo }",
	},
	{
		"set env",
		"E:a = b",
		"set E:a = b",
	},
	{
		"set variable declared by fn",
		"fn f { }; f~ = { }",
		"fn f { }; set f~ = { }",
	},
	{
		"set variable declared by use",
		"use ns; ns: = x",
		"use ns; set ns: = x",
	},
	{
		"create deleted variable",
		"var x; del x; x = foo",
		"var x; del x; var x = foo",
	},
	{
		"set argument; legacy lambda",
		"fn f [a]{ a = foo }",
		"fn f {|a| set a = foo }",
	},
	{
		"set option",
		"fn f [&a=b]{ a = foo }",
		"fn f {|&a=b| set a = foo }",
	},
	{
		"set for variable",
		"for x [] { x = foo }",
		"for x [] { set x = foo }",
	},
	{
		"set except variable",
		"try { } except x { x = foo }",
		"try { } except x { set x = foo }",
	},
	{
		"set temp variable",
		"a=b a = c",
		"a=b set a = c",
	},
	{
		"set leftover temp variable",
		"a=b nop; a = c",
		"a=b nop; set a = c",
	},
	{
		"rest variable",
		"a @b = foo bar baz",
		"var a @b = foo bar baz",
	},
	{
		"create rest variable",
		"var a; a @b = foo bar baz",
		"var a; var b; set a @b = foo bar baz",
	},
	{
		"complex lvalue group",
		"{a,@b} = foo bar",
		"var {a,@b} = foo bar",
	},
	{
		"buggy set",
		"set a = foo",
		"var a; set a = foo",
	},
	{
		"buggy set, mix of existing and new variables",
		"var a; set a b = foo bar",
		"var a; var b; set a b = foo bar",
	},
	{
		"strip local: from variable to declare",
		"local:a = foo",
		"var a = foo",
	},
	{
		"strip local: from variable to declare",
		"var a; a local:b = foo bar",
		"var a; var b; set a local:b = foo bar",
	},
}

func TestFix(t *testing.T) {
	for _, tc := range fixTests {
		t.Run(tc.name, func(t *testing.T) {
			after, err := Fix(parse.Source{Name: tc.name, Code: tc.before})
			if err != nil {
				t.Fatal(err)
			}
			if after != tc.after {
				t.Errorf("got after %q, want %q", after, tc.after)
			}
		})
	}
}
