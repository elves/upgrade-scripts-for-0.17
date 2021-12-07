package fix

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

var fixTests = []struct {
	name   string
	before string
	opts   Opts
	after  string
}{
	{
		name:   "new variable",
		before: "a = foo",
		after:  "var a = foo",
	},
	{
		name:   "new variable in lambda",
		before: "{ a = foo }",
		after:  "{ var a = foo }",
	},
	{
		name:   "set local",
		before: "var a = foo; a = bar",
		after:  "var a = foo; set a = bar",
	},
	{
		name:   "set explicit local",
		before: "var a = foo; local:a = bar",
		after:  "var a = foo; set local:a = bar",
	},
	{
		name:   "mix of var and set",
		before: "var a = foo; a b = x y",
		after:  "var a = foo; var b; set a b = x y",
	},
	{
		name:   "set builtin",
		before: "value-out-indicator = x",
		after:  "set value-out-indicator = x",
	},
	{
		name:   "set captured",
		before: "var a; { a = foo }",
		after:  "var a; { set a = foo }",
	},
	{
		name:   "set explicit upvalue",
		before: "var a; { up:a = foo }",
		after:  "var a; { set up:a = foo }",
	},
	{
		name:   "set env",
		before: "E:a = b",
		after:  "set E:a = b",
	},
	{
		name:   "set variable declared by fn",
		before: "fn f { }; f~ = { }",
		after:  "fn f { }; set f~ = { }",
	},
	{
		name:   "set variable declared by use",
		before: "use ns; ns: = x",
		after:  "use ns; set ns: = x",
	},
	{
		name:   "create deleted variable",
		before: "var x; del x; x = foo",
		after:  "var x; del x; var x = foo",
	},
	{
		name:   "set argument",
		before: "fn f [a]{ a = foo }",
		after:  "fn f [a]{ set a = foo }",
	},
	{
		name:   "set option",
		before: "fn f [&a=b]{ a = foo }",
		after:  "fn f [&a=b]{ set a = foo }",
	},
	{
		name:   "set for variable",
		before: "for x [] { x = foo }",
		after:  "for x [] { set x = foo }",
	},
	{
		name:   "set except variable",
		before: "try { } except x { x = foo }",
		after:  "try { } except x { set x = foo }",
	},
	{
		name:   "set temp variable",
		before: "a=b a = c",
		after:  "a=b set a = c",
	},
	{
		name:   "set leftover temp variable",
		before: "a=b nop; a = c",
		after:  "a=b nop; set a = c",
	},
	{
		name:   "rest variable",
		before: "a @b = foo bar baz",
		after:  "var a @b = foo bar baz",
	},
	{
		name:   "create rest variable",
		before: "var a; a @b = foo bar baz",
		after:  "var a; var b; set a @b = foo bar baz",
	},
	{
		name:   "complex lvalue group",
		before: "{a,@b} = foo bar",
		after:  "var {a,@b} = foo bar",
	},
	{
		name:   "buggy set",
		before: "set a = foo",
		after:  "var a; set a = foo",
	},
	{
		name:   "buggy set, mix of existing and new variables",
		before: "var a; set a b = foo bar",
		after:  "var a; var b; set a b = foo bar",
	},
	{
		name:   "strip local: from variable to declare",
		before: "local:a = foo",
		after:  "var a = foo",
	},
	{
		name:   "strip local: from variable to declare",
		before: "var a; a local:b = foo bar",
		after:  "var a; var b; set a local:b = foo bar",
	},

	{
		name:   "legacy lambda",
		opts:   Opts{MigrateLambda: true},
		before: "fn f [a b &k=v]{ ... }",
		after:  "fn f {|a b &k=v| ... }",
	},
	{
		name:   "legacy lambda with empty arg list",
		opts:   Opts{MigrateLambda: true},
		before: "fn f []{ ... }",
		after:  "fn f {|| ... }",
	},
	{
		name:   "legacy lambda in temp assignment",
		opts:   Opts{MigrateLambda: true},
		before: "f=[a]{ ... } nop",
		after:  "f={|a| ... } nop",
	},
}

func TestFix(t *testing.T) {
	for _, tc := range fixTests {
		t.Run(tc.name, func(t *testing.T) {
			after, err := Fix(parse.Source{Name: tc.name, Code: tc.before}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if after != tc.after {
				t.Errorf("got after %q, want %q", after, tc.after)
			}
		})
	}
}
