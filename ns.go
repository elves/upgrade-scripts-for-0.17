package fix

type staticNs map[string]struct{}

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
