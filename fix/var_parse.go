package fix

import "strings"

// Splits any leading sigil from a qualified variable name.
func splitSigil(ref string) (sigil string, qname string) {
	if ref == "" {
		return "", ""
	}
	switch ref[0] {
	case '@':
		return ref[:1], ref[1:]
	default:
		return "", ref
	}
}

// Splits a qualified name into the first namespace segment and the rest.
func splitQName(qname string) (first, rest string) {
	colon := strings.IndexByte(qname, ':')
	if colon == -1 {
		return qname, ""
	}
	return qname[:colon+1], qname[colon+1:]
}

// Splits a qualified name into namespace segments.
func splitQNameSegs(qname string) []string {
	segs := strings.SplitAfter(qname, ":")
	if len(segs) > 0 && segs[len(segs)-1] == "" {
		segs = segs[:len(segs)-1]
	}
	return segs
}
