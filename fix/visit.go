package fix

import (
	"os"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

func (cp *compiler) visit(n parse.Node) {
	switch n := n.(type) {
	case *parse.Form:
		cp.visitForm(n)
		return
	case *parse.Primary:
		if n.Type == parse.Lambda {
			cp.visitLambda(n)
			return
		}
	}
	for _, ch := range parse.Children(n) {
		cp.visit(ch)
	}
}

func (cp *compiler) visitForm(n *parse.Form) {
	for _, a := range n.Assignments {
		cp.parseIndexingLValue(a.Left, setLValue|newLValue)
		cp.visit(a.Right)
	}
	for _, r := range n.Redirs {
		cp.visit(r)
	}

	if n.Head == nil {
		// Incomplete form; nothing to do.
		return
	}

	if head, ok := cmpd.StringLiteral(n.Head); ok {
		if special, ok := builtinSpecials[head]; ok {
			// A special form
			special(cp, n)
			return
		}
	}

	for i, arg := range n.Args {
		if parse.SourceText(arg) == "=" {
			// A legacy assignment form
			lhsNodes := make([]*parse.Compound, i+1)
			lhsNodes[0] = n.Head
			copy(lhsNodes[1:], n.Args[:i])
			lvGroup := cp.parseCompoundLValues(lhsNodes, setLValue|newLValue)
			newNames := 0
			for _, lv := range lvGroup.lvalues {
				if lv.newName != "" {
					newNames++
				}
			}
			at := n.Head.From
			switch newNames {
			case 0:
				// No new names: rewrite to set
				cp.insert(at, "set ")
			case len(lvGroup.lvalues):
				// All new names: rewrite to var
				cp.insert(at, "var ")
				for _, lv := range lvGroup.lvalues {
					if strings.HasPrefix(lv.source, "local:") {
						cp.delete(lv.From, lv.From+len("local:"))
					}
				}
			default:
				// Mix of existing and new names: rewrite to var + set
				var declBuilder strings.Builder
				declBuilder.WriteString("var")
				for _, lv := range lvGroup.lvalues {
					if lv.newName != "" {
						declBuilder.WriteString(" " + lv.newName)
					}
				}
				cp.insert(at, declBuilder.String()+"; set ")
			}

			for _, a := range n.Args[i+1:] {
				cp.visit(a)
			}
			return
		}
	}

	cp.visit(n.Head)
	for _, a := range n.Args {
		cp.visit(a)
	}
	for _, o := range n.Opts {
		cp.visit(o)
	}
}

func (cp *compiler) visitLambda(n *parse.Primary) {
	if cp.opts.MigrateLambda && n.LegacyLambda {
		lbracket, rbracket := -1, -1
	ch:
		for _, ch := range parse.Children(n) {
			switch parse.SourceText(ch) {
			case "[":
				lbracket = ch.Range().From
			case "]":
				rbracket = ch.Range().From
				break ch
			}
		}
		if lbracket == -1 || rbracket == -1 {
			diag.Complain(os.Stderr, "code bug: didn't find [ or ] in legacy lambda")
		} else {
			cp.delete(lbracket, lbracket+1)
			cp.insert(lbracket, "{|")
			cp.delete(rbracket, rbracket+2)
			cp.insert(rbracket, "|")
		}
	}

	// Parse signature.
	var argNames, optNames []string
	if len(n.Elements) > 0 {
		// Argument list.
		argNames = make([]string, len(n.Elements))
		for i, arg := range n.Elements {
			ref := stringLiteralOrError(cp, arg, "argument name")
			_, qname := splitSigil(ref)
			name, rest := splitQName(qname)
			if rest != "" {
				cp.errorpf(arg, "argument name must be unqualified")
			}
			if name == "" {
				cp.errorpf(arg, "argument name must not be empty")
			}
			argNames[i] = name
		}
	}
	if len(n.MapPairs) > 0 {
		optNames = make([]string, len(n.MapPairs))
		for i, opt := range n.MapPairs {
			qname := stringLiteralOrError(cp, opt.Key, "option name")
			name, rest := splitQName(qname)
			if rest != "" {
				cp.errorpf(opt.Key, "option name must be unqualified")
			}
			if name == "" {
				cp.errorpf(opt.Key, "option name must not be empty")
			}
			optNames[i] = name
			if opt.Value == nil {
				cp.errorpf(opt.Key, "option must have default value")
			} else {
				cp.visit(opt.Value)
			}
		}
	}

	local := cp.pushScope()
	for _, argName := range argNames {
		local.add(argName)
	}
	for _, optName := range optNames {
		local.add(optName)
	}
	cp.visit(n.Chunk)
	cp.popScope()
}
