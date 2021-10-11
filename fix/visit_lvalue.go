package fix

import (
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

// Parsed group of lvalues.
type lvaluesGroup struct {
	lvalues []lvalue
	// Index of the rest variable within lvalues. If there is no rest variable,
	// the index is -1.
	rest int
}

// Parsed lvalue.
type lvalue struct {
	diag.Ranging
	newName string
}

type lvalueFlag uint

const (
	setLValue lvalueFlag = 1 << iota
	newLValue
)

func (cp *compiler) parseCompoundLValues(ns []*parse.Compound, f lvalueFlag) lvaluesGroup {
	g := lvaluesGroup{nil, -1}
	for _, n := range ns {
		if len(n.Indexings) != 1 {
			cp.errorpf(n, "lvalue may not be composite expressions")
		}
		more := cp.parseIndexingLValue(n.Indexings[0], f)
		if more.rest == -1 {
			g.lvalues = append(g.lvalues, more.lvalues...)
		} else if g.rest != -1 {
			cp.errorpf(n, "at most one rest variable is allowed")
		} else {
			g.rest = len(g.lvalues) + more.rest
			g.lvalues = append(g.lvalues, more.lvalues...)
		}
	}
	return g
}

func (cp *compiler) parseIndexingLValue(n *parse.Indexing, f lvalueFlag) lvaluesGroup {
	if n.Head.Type == parse.Braced {
		// Braced list of lvalues may not have indices.
		if len(n.Indices) > 0 {
			cp.errorpf(n, "braced list may not have indices when used as lvalue")
		}
		return cp.parseCompoundLValues(n.Head.Braced, f)
	}
	// A basic lvalue.
	if !parse.ValidLHSVariable(n.Head, true) {
		cp.errorpf(n.Head, "lvalue must be valid literal variable names")
	}
	varUse := n.Head.Value
	sigil, qname := SplitSigil(varUse)

	var foundSet bool
	if f&setLValue != 0 {
		foundSet = resolveVarRef(cp, qname, n) != nil
	}
	var newName string
	if !foundSet {
		if f&newLValue == 0 {
			cp.errorpf(n, "cannot find variable $%s", qname)
		}
		if len(n.Indices) > 0 {
			cp.errorpf(n, "name for new variable must not have indices")
		}
		segs := SplitQNameSegs(qname)
		if len(segs) == 1 {
			// Unqualified name - implicit local
			cp.thisScope().add(segs[0])
			newName = segs[0]
		} else if len(segs) == 2 && (segs[0] == "local:" || segs[0] == ":") {
			// Qualified local name
			cp.thisScope().add(segs[1])
			newName = segs[1]
		} else {
			cp.errorpf(n, "cannot create variable $%s; new variables can only be created in the local scope", qname)
		}
	}

	ends := make([]int, len(n.Indices)+1)
	ends[0] = n.Head.Range().To
	for i, idx := range n.Indices {
		ends[i+1] = idx.Range().To
	}
	lv := lvalue{n.Range(), newName}
	restIndex := -1
	if sigil == "@" {
		restIndex = 0
	}
	// TODO: Support % (and other sigils?) if https://b.elv.sh/584 is implemented for map explosion.
	return lvaluesGroup{[]lvalue{lv}, restIndex}
}
