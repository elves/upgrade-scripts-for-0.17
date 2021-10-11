package fix

// The compile-time representation of a namespace. Called "static" namespace
// since it contains information that are known without executing the code.
// The data structure itself, however, is not static, and gets mutated as the
// compiler gains more information about the namespace. The zero value of
// staticNs is an empty namespace.
type staticNs struct {
	names   []string
	deleted []bool
}

func (ns *staticNs) clone() *staticNs {
	return &staticNs{
		append([]string{}, ns.names...), append([]bool{}, ns.deleted...)}
}

func (ns *staticNs) del(k string) {
	if i := ns.lookup(k); i != -1 {
		ns.deleted[i] = true
	}
}

// Adds a name, shadowing any existing one.
func (ns *staticNs) add(k string) int {
	ns.del(k)
	return ns.addInner(k)
}

// Adds a name, assuming that it either doesn't exist yet or has been deleted.
func (ns *staticNs) addInner(k string) int {
	ns.names = append(ns.names, k)
	ns.deleted = append(ns.deleted, false)
	return len(ns.names) - 1
}

func (ns *staticNs) lookup(k string) int {
	for i, name := range ns.names {
		if name == k && !ns.deleted[i] {
			return i
		}
	}
	return -1
}

type staticUpNs struct {
	names []string
	// For each name, whether the upvalue comes from the immediate outer scope,
	// i.e. the local scope a lambda is evaluated in.
	local []bool
	// Index of the upvalue variable, either into the local scope (if
	// the corresponding value in local is true) or the up scope (if the
	// corresponding value in local is false).
	index []int
}

func (up *staticUpNs) add(k string, local bool, index int) int {
	for i, name := range up.names {
		if name == k {
			return i
		}
	}
	up.names = append(up.names, k)
	up.local = append(up.local, local)
	up.index = append(up.index, index)
	return len(up.names) - 1
}
