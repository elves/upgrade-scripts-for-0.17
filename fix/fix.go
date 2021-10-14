package fix

import (
	"fmt"
	"sort"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

const (
	fnSuffix = "~"
	nsSuffix = ":"
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Builtin namespace.
	builtin staticNs
	// Lexical namespaces.
	scopes []staticNs
	// Information about the source.
	srcMeta parse.Source

	inserts []insert
	deletes []diag.Ranging
}

type insert struct {
	pos  int
	text string
}

func Fix(src parse.Source) (string, error) {
	t, err := parse.Parse(src, parse.Config{})
	if err != nil {
		return "", err
	}
	inserts, deletes, err := compile(builtinNs, t)
	if err != nil {
		return "", err
	}
	// TODO: Apply replacement
	return applyDiff(src.Code, inserts, deletes), nil
}

func applyDiff(s string, inserts []insert, deletes []diag.Ranging) string {
	var sb strings.Builder
	insertIdx := 0
	deleteIdx := 0
	for i, r := range s {
		for insertIdx < len(inserts) && i == inserts[insertIdx].pos {
			sb.WriteString(inserts[insertIdx].text)
			insertIdx++
		}
		if deleteIdx < len(deletes) {
			if d := deletes[deleteIdx]; d.From <= i && i < d.To {
				continue
			} else if i == d.To {
				deleteIdx++
			}
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func compile(b staticNs, tree parse.Tree) (inserts []insert, deletes []diag.Ranging, err error) {
	cp := &compiler{b, []staticNs{makeStaticNs("edit:")}, tree.Source, nil, nil}
	defer func() {
		r := recover()
		if r == nil {
			return
		} else if e := getCompilationError(r); e != nil {
			// Save the compilation error and stop the panic.
			err = e
		} else {
			// Resume the panic; it is not supposed to be handled here.
			panic(r)
		}
	}()
	cp.visit(tree.Root)
	sort.Slice(cp.inserts, func(i, j int) bool {
		return cp.inserts[i].pos < cp.inserts[j].pos
	})
	sort.Slice(cp.deletes, func(i, j int) bool {
		return cp.deletes[i].From < cp.deletes[j].From
	})
	return cp.inserts, cp.deletes, nil
}

const compilationErrorType = "compilation error"

func (cp *compiler) errorpf(r diag.Ranger, format string, args ...interface{}) {
	// The panic is caught by the recover in compile above.
	panic(&diag.Error{
		Type:    compilationErrorType,
		Message: fmt.Sprintf(format, args...),
		Context: *diag.NewContext(cp.srcMeta.Name, cp.srcMeta.Code, r)})
}

// Returns a *diag.Error if the given value is a compilation error. Otherwise it
// returns nil.
func getCompilationError(e interface{}) *diag.Error {
	if e, ok := e.(*diag.Error); ok && e.Type == compilationErrorType {
		return e
	}
	return nil
}

func (cp *compiler) insert(pos int, text string) {
	cp.inserts = append(cp.inserts, insert{pos, text})
}

func (cp *compiler) delete(from, to int) {
	cp.deletes = append(cp.deletes, diag.Ranging{From: from, To: to})
}

func (cp *compiler) thisScope() staticNs {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() staticNs {
	sc := make(staticNs)
	cp.scopes = append(cp.scopes, sc)
	return sc
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
}

type staticNs map[string]struct{}

var builtinNs = makeStaticNs(
	"!=s~", "!=~", "%~", "*~", "+~",
	"-gc~", "-ifaddrs~", "-log~", "-override-wcwidth~", "-stack~",
	"-~", "/~", "<=s~", "<=~", "<s~", "<~", "==s~", "==~", ">=s~", ">=~", ">s~", ">~",
	"_",
	"after-chdir", "all~", "args", "assoc~",
	"base~", "before-chdir", "bool~", "break~", "buildinfo",
	"cd~", "constantly~", "continue~", "count~",
	"deprecate~", "dir-history~", "dissoc~", "drop~",
	"each~", "eawk~", "echo~", "eq~", "eval~", "exact-num~", "exec~", "exit~", "external~",
	"fail~", "false", "fg~", "float64~", "from-json~", "from-lines~", "from-terminated~",
	"get-env~", "has-env~",
	"has-external~", "has-key~", "has-value~",
	"is~",
	"keys~", "kind-of~",
	"make-map~", "multi-error~",
	"nil", "nop~", "not-eq~", "notify-bg-job-success", "not~", "ns~", "num-bg-jobs", "num~",
	"ok",
	"one~", "only-bytes~", "only-values~", "order~",
	"paths", "peach~", "pid", "pprint~", "printf~", "print~", "put~", "pwd",
	"randint~", "rand~", "range~", "read-line~", "read-upto~", "repeat~", "repr~", "resolve~", "return~", "run-parallel~",
	"search-external~", "set-env~", "show~", "sleep~", "slurp~", "src~", "styled-segment~", "styled~",
	"take~", "tilde-abbr~", "time~", "to-json~", "to-lines~", "to-string~", "to-terminated~", "true",
	"unset-env~", "use-mod~",
	"value-out-indicator", "version",
	"wcswidth~",
)

func (ns staticNs) del(k string) {
	delete(ns, k)
}

func (ns staticNs) add(k string) {
	ns[k] = struct{}{}
}

func (ns staticNs) has(k string) bool {
	_, ok := ns[k]
	return ok
}

func makeStaticNs(names ...string) staticNs {
	ns := make(staticNs)
	for _, name := range names {
		ns.add(name)
	}
	return ns
}
