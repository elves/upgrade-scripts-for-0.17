package fix

import (
	"fmt"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

const (
	FnSuffix = "~"
	NsSuffix = ":"
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Builtin namespace.
	builtin *staticNs
	// Lexical namespaces.
	scopes []*staticNs
	// Sources of captured variables.
	captures []*staticUpNs
	// Information about the source.
	srcMeta parse.Source

	replacement map[diag.Ranging]string
}

func compile(b, g *staticNs, tree parse.Tree) (err error) {
	g = g.clone()
	cp := &compiler{
		b, []*staticNs{g}, []*staticUpNs{new(staticUpNs)},
		tree.Source, make(map[diag.Ranging]string)}
	defer func() {
		r := recover()
		if r == nil {
			return
		} else if e := GetCompilationError(r); e != nil {
			// Save the compilation error and stop the panic.
			err = e
		} else {
			// Resume the panic; it is not supposed to be handled here.
			panic(r)
		}
	}()
	cp.visit(tree.Root)
	return nil
}

const compilationErrorType = "compilation error"

func (cp *compiler) errorpf(r diag.Ranger, format string, args ...interface{}) {
	// The panic is caught by the recover in compile above.
	panic(&diag.Error{
		Type:    compilationErrorType,
		Message: fmt.Sprintf(format, args...),
		Context: *diag.NewContext(cp.srcMeta.Name, cp.srcMeta.Code, r)})
}

// GetCompilationError returns a *diag.Error if the given value is a compilation
// error. Otherwise it returns nil.
func GetCompilationError(e interface{}) *diag.Error {
	if e, ok := e.(*diag.Error); ok && e.Type == compilationErrorType {
		return e
	}
	return nil
}

func (cp *compiler) thisScope() *staticNs {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) pushScope() (*staticNs, *staticUpNs) {
	sc := new(staticNs)
	up := new(staticUpNs)
	cp.scopes = append(cp.scopes, sc)
	cp.captures = append(cp.captures, up)
	return sc, up
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
	cp.captures[len(cp.captures)-1] = nil
	cp.captures = cp.captures[:len(cp.captures)-1]
}
