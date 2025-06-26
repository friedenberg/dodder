package tridex

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func (n *node) Add(v string) {
	if len(v) == 0 {
		n.IncludesTerminus = true
		return
	}

	if v != n.Value {
		n.Count += 1
	}

	if n.Count == 1 {
		n.Value = v
		return
	} else if n.Value != "" && n.Value != v {
		n.Add(n.Value)
		n.Value = ""
	}

	c := v[0]

	var child node
	ok := false
	child, ok = n.Children[c]

	if !ok {
		child = node{Children: make(map[byte]node)}
	}

	child.Add(v[1:])

	n.Children[c] = child
}

func (n *node) Remove(v string) {
	if v == "" {
		n.Count -= 1
		n.IncludesTerminus = false
		return
	}

	if n.Value == v {
		n.Count -= 1
		n.Value = ""
		return
	}

	first := v[0]

	rest := ""

	if len(v) > 1 {
		rest = v[1:]
	}

	child, ok := n.Children[first]

	if ok {
		child.Remove(rest)
		n.Count -= 1

		if child.Count == 0 {
			delete(n.Children, first)
		} else {
			n.Children[first] = child
		}
	}
}

func (n node) Contains(v string) (ok bool) {
	if len(v) == 0 {
		ok = true
		return
	}

	if n.Count == 1 && n.Value != "" {
		ok = strings.HasPrefix(n.Value, v)
		return
	}

	c := v[0]

	var child node

	child, ok = n.Children[c]

	if ok {
		ok = child.Contains(v[1:])
	}

	return
}

func (n node) ContainsExactly(v string) (ok bool) {
	if len(v) == 0 {
		ok = n.IncludesTerminus
		return
	}

	if n.Value != "" {
		ok = n.Value == v
		return
	}

	c := v[0]

	var child node

	child, ok = n.Children[c]

	if ok {
		ok = child.ContainsExactly(v[1:])
	}

	return
}

func (n node) Any() byte {
	for c := range n.Children {
		return c
	}

	return 0
}

func (n node) Expand(v string, sb *strings.Builder) (ok bool) {
	var c byte
	var rem string

	if len(v) == 0 {
		switch n.Count {

		case 0:
			return true

		case 1:
			if !n.IncludesTerminus {
				sb.WriteString(n.Value)
			}

			return true
		}
	} else {
		switch n.Count {
		case 1:
			ok = strings.HasPrefix(n.Value, v)

			if ok {
				sb.WriteString(n.Value)
			}

			return

		default:
			rem = v[1:]
			c = v[0]
		}
	}

	var child node

	if child, ok = n.Children[c]; ok {
		sb.WriteByte(c)
		return child.Expand(rem, sb)
	}

	return
}

func (n node) Abbreviate(v string, loc int) string {
	if n.IsRoot && len(n.Children) == 0 {
		if n.Value != "" {
			return n.Value[0:1]
		} else {
			return ""
		}
	}

	if len(v)-1 < loc {
		return v
	}

	if n.Count == 1 && n.ContainsExactly(v[loc:]) && !n.IncludesTerminus {
		return v[0:loc]
	}

	c := v[loc]

	child, ok := n.Children[c]

	if ok {
		return child.Abbreviate(v, loc+1)
	} else {
		if len(v)-1 < loc {
			return v
		} else {
			return v[0 : loc+1]
		}
	}
}

func (a *node) Copy() (b node) {
	b = *a

	for i, c := range b.Children {
		b.Children[i] = c.Copy()
	}

	return
}

func (n *node) Each(f interfaces.FuncIter[string], acc string) (err error) {
	if n.Value != "" {
		if err = f(acc + n.Value); err != nil {
			return
		}
	}

	if n.IncludesTerminus {
		if err = f(acc); err != nil {
			return
		}
	}

	for r, c := range n.Children {
		if err = c.Each(f, acc+string(r)); err != nil {
			return
		}
	}

	return
}
