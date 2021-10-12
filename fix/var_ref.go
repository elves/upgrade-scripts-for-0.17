package fix

import (
	"strings"

	"src.elv.sh/pkg/diag"
)

type varRef struct {
	local    bool
	subNames []string
}

// Resolves a qname into a varRef.
func resolveVarRef(s *compiler, qname string, r diag.Ranger) *varRef {
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

func resolveVarRefLocal(s *compiler, qname string) *varRef {
	first, rest := splitQName(qname)
	if s.searchLocal(first) {
		return &varRef{local: true, subNames: splitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefCapture(s *compiler, qname string) *varRef {
	first, rest := splitQName(qname)
	if s.searchCapture(first) {
		return &varRef{subNames: splitQNameSegs(rest)}
	}
	return nil
}

func resolveVarRefBuiltin(s *compiler, qname string, r diag.Ranger) *varRef {
	first, rest := splitQName(qname)
	if rest != "" {
		// Try special namespace first.
		switch first {
		case "local:":
			return resolveVarRefLocal(s, rest)
		case "up:":
			return resolveVarRefCapture(s, rest)
		case "e:":
			if strings.HasSuffix(rest, fnSuffix) {
				return &varRef{subNames: []string{rest[:len(rest)-1]}}
			}
		case "E:":
			return &varRef{subNames: []string{rest}}
		}
	}
	if s.searchBuiltin(first) {
		return &varRef{subNames: splitQNameSegs(rest)}
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
