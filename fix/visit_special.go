package fix

import (
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

type visitSpecial func(*compiler, *parse.Form)

var builtinSpecials map[string]visitSpecial

func init() {
	builtinSpecials = map[string]visitSpecial{
		"var": visitVar,
		"set": visitSet,
		"del": visitDel,
		"fn":  visitFn,

		"use": visitUse,

		"for": visitFor,
		"try": visitTry,

		"and":      ordinary,
		"or":       ordinary,
		"coalesce": ordinary,
		"if":       ordinary,
		"while":    ordinary,
		"pragma":   ordinary,
	}
}

func ordinary(cp *compiler, n *parse.Form) {
	cp.visit(n.Head)
	for _, a := range n.Args {
		cp.visit(a)
	}
}

// VarForm = 'var' { VariablePrimary } [ '=' { Compound } ]
func visitVar(cp *compiler, fn *parse.Form) {
	eqIndex := -1
	for i, cn := range fn.Args {
		if parse.SourceText(cn) == "=" {
			eqIndex = i
			break
		}
	}

	if eqIndex == -1 {
		cp.parseCompoundLValues(fn.Args, newLValue)
		// Just create new variables, nothing extra to do at runtime.
		return
	}
	// Visit rhs before lhs before many potential shadowing
	for _, a := range fn.Args[eqIndex+1:] {
		cp.visit(a)
	}
	cp.parseCompoundLValues(fn.Args[:eqIndex], newLValue)
}

// SetForm = 'set' { LHS } '=' { Compound }
func visitSet(cp *compiler, fn *parse.Form) {
	eqIndex := -1
	for i, cn := range fn.Args {
		if parse.SourceText(cn) == "=" {
			eqIndex = i
			break
		}
	}
	if eqIndex == -1 {
		cp.errorpf(diag.PointRanging(fn.Range().To), "need = and right-hand-side")
	}
	// 0.16 has a buggy version of "set" that has the semantics of legacy
	// assignment, i.e. it can also create new variable. Pre-declare new
	// variables with "var", if any.
	lvGroup := cp.parseCompoundLValues(fn.Args[:eqIndex], setLValue|newLValue)
	hasNew := false
	var declBuilder strings.Builder
	declBuilder.WriteString("var")
	for _, lv := range lvGroup.lvalues {
		if lv.newName != "" {
			declBuilder.WriteString(" " + lv.newName)
			hasNew = true
		}
	}
	if hasNew {
		cp.insert(fn.Head.From, declBuilder.String()+"; ")
	}

	for _, a := range fn.Args[eqIndex+1:] {
		cp.visit(a)
	}
}

const delArgMsg = "arguments to del must be variable or variable elements"

// DelForm = 'del' { LHS }
func visitDel(cp *compiler, fn *parse.Form) {
	for _, cn := range fn.Args {
		if len(cn.Indexings) != 1 {
			cp.errorpf(cn, delArgMsg)
			continue
		}
		head, indices := cn.Indexings[0].Head, cn.Indexings[0].Indices
		if head.Type == parse.Variable {
			cp.errorpf(cn, "arguments to del must drop $")
		} else if !parse.ValidLHSVariable(head, false) {
			cp.errorpf(cn, delArgMsg)
		}

		qname := head.Value
		ref := resolveVarRef(cp, qname, nil)
		if ref == nil {
			cp.errorpf(cn, "no variable $%s", head.Value)
			continue
		}
		if len(indices) == 0 {
			if ref.local && len(ref.subNames) == 0 {
				cp.thisScope().del(qname)
			}
		}
	}
}

// FnForm = 'fn' StringPrimary LambdaPrimary
//
// fn f []{foobar} is a shorthand for set '&'f = []{foobar}.
func visitFn(cp *compiler, fn *parse.Form) {
	args := cp.walkArgs(fn)
	nameNode := args.next()
	name := stringLiteralOrError(cp, nameNode, "function name")
	bodyNode := args.nextMustLambda("function body")
	args.mustEnd()

	// Define the variable before compiling the body, so that the body may refer
	// to the function itself.
	cp.thisScope().add(name + fnSuffix)
	cp.visitLambda(bodyNode)
}

// UseForm = 'use' StringPrimary
func visitUse(cp *compiler, fn *parse.Form) {
	var name string

	switch len(fn.Args) {
	case 0:
		end := fn.Head.Range().To
		cp.errorpf(diag.PointRanging(end), "lack module name")
	case 1:
		spec := stringLiteralOrError(cp, fn.Args[0], "module spec")
		// Use the last path component as the name; for instance, if path =
		// "a/b/c/d", name is "d". If path doesn't have slashes, name = path.
		name = spec[strings.LastIndexByte(spec, '/')+1:]
	case 2:
		name = stringLiteralOrError(cp, fn.Args[1], "module name")
	default: // > 2
		cp.errorpf(diag.MixedRanging(fn.Args[2], fn.Args[len(fn.Args)-1]),
			"superfluous argument(s)")
	}

	cp.thisScope().add(name + nsSuffix)
}

func visitFor(cp *compiler, fn *parse.Form) {
	args := cp.walkArgs(fn)
	varNode := args.next()
	iterNode := args.next()
	bodyNode := args.nextMustLambda("for body")
	elseNode := args.nextMustLambdaIfAfter("else")
	args.mustEnd()

	cp.compileOneLValue(varNode, setLValue|newLValue)

	cp.visit(iterNode)
	cp.visit(bodyNode)
	if elseNode != nil {
		cp.visit(elseNode)
	}
}

func visitTry(cp *compiler, fn *parse.Form) {
	args := cp.walkArgs(fn)
	bodyNode := args.nextMustLambda("try body")
	var exceptVarNode *parse.Compound
	var exceptNode *parse.Primary
	if args.nextIs("except") {
		// Parse an optional lvalue into exceptVarNode.
		n := args.peek()
		if _, ok := cmpd.StringLiteral(n); ok {
			exceptVarNode = n
			args.next()
		}
		exceptNode = args.nextMustLambda("except body")
	}
	elseNode := args.nextMustLambdaIfAfter("else")
	finallyNode := args.nextMustLambdaIfAfter("finally")
	args.mustEnd()

	cp.visit(bodyNode)
	if exceptVarNode != nil {
		cp.compileOneLValue(exceptVarNode, setLValue|newLValue)
	}
	if exceptNode != nil {
		cp.visit(exceptNode)
	}
	if elseNode != nil {
		cp.visit(elseNode)
	}
	if finallyNode != nil {
		cp.visit(finallyNode)
	}
}

func (cp *compiler) compileOneLValue(n *parse.Compound, f lvalueFlag) lvalue {
	if len(n.Indexings) != 1 {
		cp.errorpf(n, "must be valid lvalue")
	}
	lvalues := cp.parseIndexingLValue(n.Indexings[0], f)
	if lvalues.rest != -1 {
		cp.errorpf(lvalues.lvalues[lvalues.rest], "rest variable not allowed")
	}
	if len(lvalues.lvalues) != 1 {
		cp.errorpf(n, "must be exactly one lvalue")
	}
	return lvalues.lvalues[0]
}
