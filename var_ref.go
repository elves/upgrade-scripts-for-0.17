package fix

import (
	"strings"

	"src.elv.sh/pkg/diag"
)

// This file implements variable resolution. Elvish has fully static lexical
// scopes, so variable resolution involves some work in the compilation phase as
// well.
//
// During compilation, a qualified variable name (whether in lvalue, like "x
// = foo", or in variable use, like "$x") is searched in compiler's staticNs
// tables to determine which scope they belong to, as well as their indices in
// that scope. This step is just called "resolve" in the code, and it stores
// information in a varRef struct.
//
// During evaluation, the varRef is then used to look up the Var for the
// variable. This step is called "deref" in the code.
//
// The resolve phase can take place during evaluation as well for introspection.

// Keeps all the information statically about a variable referenced by a
// qualified name.
type varRef struct {
	scope    varScope
	subNames []string
}

type varScope int

const (
	localScope varScope = 1 + iota
	captureScope
	builtinScope
	envScope
	externalScope
)

// An interface satisfied by both *compiler and *Frame. Used to implement
// resolveVarRef as a function that works for both types.
type scopeSearcher interface {
	searchLocal(k string) bool
	searchCapture(k string) bool
	searchBuiltin(k string) bool
}

// Resolves a qname into a varRef.
func resolveVarRef(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	qname = strings.TrimPrefix(qname, ":")
	if ref := resolveVarRefLocal(s, qname); ref != nil {
		return ref
	}
	if ref := resolveVarRefCapture(s, qname); ref != nil {
		return ref
	}
	if ref := resolveVarRefBuiltin(s, qname, r); ref != nil {
		return ref
	}
	return nil
}

func resolveVarRefLocal(s scopeSearcher, qname string) *varRef {
	first, rest := SplitQName(qname)
	if s.searchLocal(first) {
		return &varRef{scope: localScope, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefCapture(s scopeSearcher, qname string) *varRef {
	first, rest := SplitQName(qname)
	if s.searchCapture(first) {
		return &varRef{scope: captureScope, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefBuiltin(s scopeSearcher, qname string, r diag.Ranger) *varRef {
	first, rest := SplitQName(qname)
	if rest != "" {
		// Try special namespace first.
		switch first {
		case "local:":
			return resolveVarRefLocal(s, rest)
		case "up:":
			return resolveVarRefCapture(s, rest)
		case "e:":
			if strings.HasSuffix(rest, FnSuffix) {
				return &varRef{scope: externalScope, subNames: []string{rest[:len(rest)-1]}}
			}
		case "E:":
			return &varRef{scope: envScope, subNames: []string{rest}}
		}
	}
	if s.searchBuiltin(first) {
		return &varRef{scope: builtinScope, subNames: SplitQNameSegs(rest)}
	}
	return nil
}

func (cp *compiler) searchLocal(k string) bool {
	return cp.thisScope().has(k)
}

func (cp *compiler) searchCapture(k string) bool {
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		if cp.scopes[i].has(k) {
			return true
		}
	}
	return false
}

func (cp *compiler) searchBuiltin(k string) bool {
	return cp.builtin.has(k)
}
