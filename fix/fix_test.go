package fix

import (
	"testing"

	"src.elv.sh/pkg/parse"
)

var fixTests = []struct {
	name, before, after string
}{
	{
		"var",
		"a = foo",
		"var a = foo",
	},
	{
		"var in lambda",
		"{ a = foo }",
		"{ var a = foo }",
	},
	{
		"set",
		"var a = foo; a = bar",
		"var a = foo; set a = bar",
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
		"set fn",
		"fn f { }; f~ = { }",
		"fn f { }; set f~ = { }",
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
