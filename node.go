package ibtree

// node is a generic type that represents a node in the AVL Tree.
type node[T any] struct {
	l    *node[T] // left child
	r    *node[T] // right child
	genH uint64   // Generation and height of the node.
	i    T        // The item the node is holding.
}

// nodeStack keeps track of nodes that are modified during insert and delete operations.
// The node at position 0 is the root of the tree.
type nodeStack[T any] struct {
	s   []*node[T] // The stack of nodes we are currently manipulating.
	gen uint64
}

func (ns *nodeStack[T]) clear() {
	ns.s = ns.s[:0]
}

func (ns *nodeStack[T]) newNode(v T) *node[T] {
	return &node[T]{i: v, genH: (ns.gen << 8) | 0x01}
}

func (ns *nodeStack[T]) copy(n *node[T]) *node[T] {
	if n.gen() == ns.gen {
		return n
	}
	return &node[T]{l: n.l, r: n.r, i: n.i, genH: (ns.gen << 8) | (n.h())}
}

func (ns *nodeStack[T]) add(n *node[T]) {
	ns.s = append(ns.s, ns.copy(n))
}

func (ns *nodeStack[T]) addLeft(n *node[T]) {
	i := len(ns.s)
	ns.s = append(ns.s, ns.copy(n))
	ns.s[i-1].l = ns.s[i]
}

func (ns *nodeStack[T]) addRight(n *node[T]) {
	i := len(ns.s)
	ns.s = append(ns.s, ns.copy(n))
	ns.s[i-1].r = ns.s[i]
}

func (ns *nodeStack[T]) pos(i int) int {
	if i >= 0 {
		return i
	}
	return len(ns.s) + i
}

func (ns *nodeStack[T]) at(i int) *node[T] {
	return ns.s[ns.pos(i)]
}

func (ns *nodeStack[T]) set(at int, v *node[T]) {
	ns.s[ns.pos(at)] = v
}

func (ns *nodeStack[T]) drop() {
	ns.set(ns.pos(-1), nil)
	ns.s = ns.s[:ns.pos(-1)]
}

// gen returns this node's generation.
func (n *node[T]) gen() uint64 {
	return n.genH >> 8
}

// h returns the node's height in the tree from the least significant byte of genH.
// This limits the tree height to 255, but given that the wost case height of
// an AVL tree is 1.44(log(n)) we will never overflow it on a 64 bit system
func (n *node[T]) h() uint64 {
	return n.genH & 0xff
}

// balance calculates the relative balance of a node.
// Negative numbers indicate a subtree that is left-heavy,
// and positive numbers indicate a Tree that is right-heavy.
func (n *node[T]) balance() (res int) {
	if n.l != nil {
		res -= int(n.l.h())
	}
	if n.r != nil {
		res += int(n.r.h())
	}
	return
}

// setHeight calculates the height of this node.
func (n *node[T]) setHeight() {
	h := uint64(0)
	if n.l != nil {
		h = n.l.h()
	}
	if n.r != nil {
		if rh := n.r.h(); rh >= h {
			h = rh
		}
	}
	h++
	n.genH &= ^uint64(0xff)
	n.genH |= h
	return
}

// rotateLeft transforms
//
//	 |
//	 a
//	/ \
//
// x   b
//
//	 / \
//	y   z
//
// to
//
//	   |
//	   b
//	  / \
//	 a   z
//	/ \
//
// x   y
func (a *node[T]) rotateLeft() (b *node[T]) {
	b = a.r
	a.r = b.l
	b.l = a
	return
}

// rotateRight is the inverse of rotateLeft. it transforms
//
//	   |
//	   a(h)
//	  / \
//	 b   z
//	/ \
//
// x   y
//
// to
//
//	 |
//	 b
//	/ \
//
// x   a
//
//	 / \
//	y   z
func (a *node[T]) rotateRight() (b *node[T]) {
	b = a.l
	a.l = b.r
	b.r = a
	return
}

func (t *Tree[T]) getExact(ins *nodeStack[T], n *node[T], v T) int {
	ins.clear()
	ins.add(n)
	for n != nil {
		if t.less(n.i, v) {
			// I expect the common case to be inserting things in ascending order.
			if n.r == nil {
				return Greater
			}
			ins.addRight(n.r)
			n = n.r
		} else if t.less(v, n.i) {
			if n.l == nil {
				return Less
			}
			ins.addLeft(n.l)
			n = n.l
		} else {
			break
		}
	}
	return Equal
}

func (n *node[T]) getLeftmost(res *nodeStack[T]) {
	res.addRight(n.r)
	n = n.r
	for n.l != nil {
		res.addLeft(n.l)
		n = n.l
	}
}

func (n *node[T]) getRightmost(res *nodeStack[T]) {
	res.addLeft(n.l)
	n = n.l
	for n.r != nil {
		res.addRight(n.r)
		n = n.r
	}
}

// min finds the minimal child of h
func min[T any](n *node[T]) *node[T] {
	for n.l != nil {
		n = n.l
	}
	return n
}

// max finds the maximal child of h
func max[T any](n *node[T]) *node[T] {
	for n.r != nil {
		n = n.r
	}
	return n
}

func (n *node[T]) swapChild(was, is *node[T]) *node[T] {
	if n.l == was {
		n.l = is
	} else if n.r == was {
		n.r = is
	} else {
		panic(`Impossible`)
	}
	return is
}

// rebalance walks up the Tree starting at node n, rebalancing nodes
// that no longer meet the AVL balance criteria. rebalance will continue until
// it either walks all the way up the Tree, or the node has the
// same height it started with.
func rebalance[T any](ins *nodeStack[T]) {
	var n *node[T]
	for i := len(ins.s) - 1; i >= 0; i-- {
		n = ins.s[i]
		oh := n.h()
		switch n.balance() {
		case Less, Equal, Greater:
		case rightHeavy:
			// Tree is excessively right-heavy, rotate it to the left.
			n.r = ins.copy(n.r)
			if n.r.balance() < 0 {
				n.r.l = ins.copy(n.r.l)
				// Right Tree is left-heavy, which would cause the next rotation to result in overall left-heaviness.
				// Rotate the right Tree to the right to counteract this.
				n.r = n.r.rotateRight()
				n.r.r.setHeight()
			}
			if i > 0 {
				n = ins.s[i-1].swapChild(n, n.rotateLeft())
			} else {
				n = n.rotateLeft()
			}
			n.l.setHeight()
		case leftHeavy:
			// Tree is excessively left-heavy, rotate it to the right
			n.l = ins.copy(n.l)
			if n.l.balance() > 0 {
				n.l.r = ins.copy(n.l.r)
				// The left Tree is right-heavy, which would cause the next rotation to result in overall right-heaviness.
				// Rotate the left Tree to the left to compensate.
				n.l = n.l.rotateLeft()
				n.l.l.setHeight()
			}
			if i > 0 {
				n = ins.s[i-1].swapChild(n, n.rotateRight())
			} else {
				n = n.rotateRight()
			}
			n.r.setHeight()
		default:
			panic("Tree too far out of shape!")
		}
		ins.s[i] = n
		n.setHeight()
		if oh == n.h() {
			break
		}
	}
}
