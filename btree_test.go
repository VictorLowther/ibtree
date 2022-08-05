package ibtree

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func (n *node[T]) height() uint {
	if n == nil {
		return 0
	}
	return n.h
}

// balanced checks a Tree to ensure it is AVL compliant.
// Only for use when running tests.
func (n *node[T]) balanced(t *testing.T) {
	if n == nil {
		return
	}
	if n.h == 0 {
		panic("Zero height")
	}
	if n.h == 1 && !(n.r == nil && n.l == nil) {
		panic("Height 1 node has children")
	}
	if n.h > 1 && n.r == nil && n.l == nil {
		panic("Interior node has no children")
	}
	lh, rh := n.l.height(), n.r.height()
	if lh >= n.h || rh >= n.h {
		panic("Child height greater than ours")
	}
	if !(n.h-lh == 1 || n.h-rh == 1) {
		panic("Height not max(lh,rh)+1")
	}
	b := n.balance()
	rb := int(rh) - int(lh)
	if b != rb {
		panic("Balance calculated incorrectly")
	}
	if b > 1 {
		panic("Too heavy to the right!")
	} else if b < -1 {
		panic("Too heavy to the left!")
	}
	if n.l != nil {
		n.l.balanced(t)
	}
	if n.r != nil {
		n.r.balanced(t)
	}
}

func il(a, b int) bool { return a < b }

func TestRotate(t *testing.T) {
	tree := New[int](il, 1, 0, 3, 2, 4)
	if tree.root.i != 1 {
		t.Fatalf("Tree.root.i %d, not 1", tree.root.i)
	}
	if tree.root.l.i != 0 {
		t.Fatalf("Tree root.l.i %d, not 0", tree.root.l.i)
	}
	if tree.root.r.i != 3 {
		t.Fatalf("Tree.root.r.i %d, not 3", tree.root.r.i)
	}
	if tree.root.r.l.i != 2 {
		t.Fatalf("Tree.root.r.l.i %d, not 2", tree.root.r.l.i)
	}
	if tree.root.r.r.i != 4 {
		t.Fatalf("Tree.root.r.r.i %d, not 4", tree.root.r.r.i)
	}
	t.Logf("Tree populated correctly")
	tree.root.balanced(t)
	tree.root = tree.root.rotateLeft()
	tree.root.l.setHeight()
	tree.root.setHeight()
	if tree.root.i != 3 {
		t.Fatalf("Tree.root.i %d, not 3", tree.root.i)
	}
	if tree.root.l.i != 1 {
		t.Fatalf("Tree root.l.i %d, not 1", tree.root.l.i)
	}
	if tree.root.l.l.i != 0 {
		t.Fatalf("Tree.root.l.l.i %d, not 0", tree.root.l.l.i)
	}
	if tree.root.l.r.i != 2 {
		t.Fatalf("Tree.root.l.r.i %d, not 2", tree.root.l.r.i)
	}
	if tree.root.r.i != 4 {
		t.Fatalf("Tree.root.r.i %d, not 4", tree.root.r.i)
	}
	t.Logf("Tree rotated left correctly")
	tree.root.balanced(t)
	tree.root = tree.root.rotateRight()
	tree.root.r.setHeight()
	tree.root.setHeight()
	if tree.root.i != 1 {
		t.Fatalf("Tree.root.i %d, not 1", tree.root.i)
	}
	if tree.root.l.i != 0 {
		t.Fatalf("Tree root.l.i %d, not 0", tree.root.l.i)
	}
	if tree.root.r.i != 3 {
		t.Fatalf("Tree.root.r.i %d, not 3", tree.root.r.i)
	}
	if tree.root.r.l.i != 2 {
		t.Fatalf("Tree.root.r.l.i %d, not 2", tree.root.r.l.i)
	}
	if tree.root.r.r.i != 4 {
		t.Fatalf("Tree.root.r.r.i %d, not 4", tree.root.r.r.i)
	}
	tree.root.balanced(t)
	t.Logf("Tree rotated right correctly")
	tree.Reverse()
	tree.root.balanced(t)
}

func TestCases(t *testing.T) {
	tree := New[int](il, 1)
	cmp := tree.Cmp(1)
	if tree.Len() != 1 {
		t.Fatalf("expecting len 1")
	}
	if !tree.Has(cmp) {
		t.Fatalf("expecting to find key=1")
	}

	tree, _, _ = tree.Delete(1)
	if tree.Len() != 0 {
		t.Fatalf("expecting len 0")
	}
	if tree.Has(cmp) {
		t.Fatalf("not expecting to find key=1")
	}

	tree, _, _ = tree.Delete(1)
	if tree.Len() != 0 {
		t.Fatalf("expecting len 0")
	}
	if tree.Has(cmp) {
		t.Fatalf("not expecting to find key=1")
	}
}

func sl(a, b string) bool { return a < b }
func TestRange(t *testing.T) {
	tree := New[string](sl, "ab", "aba", "abc", "a", "aa", "aaa", "b", "a-", "a!")
	expect := []string{"ab", "aba", "abc"}
	res := []string{}
	tree.Range(Lt(tree.Cmp("ab")), Gt(tree.Cmp("ac")), func(idx string) bool {
		res = append(res, idx)
		return true
	})
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	tree.Range(Lte(tree.Cmp("aaa")), Gte(tree.Cmp("b")), func(idx string) bool {
		res = append(res, idx)
		return true
	})
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
}

func TestIter(t *testing.T) {
	tree := New[string](sl, "ab", "aba", "abc", "a", "aa", "aaa", "b", "a-", "a!")
	expect := []string{"ab", "aba", "abc"}
	res := []string{}
	iter := tree.Iterator(Lt(tree.Cmp("ab")), Gt(tree.Cmp("ac")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	iter = tree.Iterator(Lte(tree.Cmp("aaa")), Gte(tree.Cmp("b")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	res = nil
	expect = nil
	iter = tree.Iterator(Lt(tree.Cmp("z")), nil)
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
	iter = tree.Iterator(nil, Gt(tree.Cmp("0")))
	for iter.Next() {
		res = append(res, iter.Item())
	}
	if !reflect.DeepEqual(expect, res) {
		t.Fatalf("Range failed: expected %v, got %v", expect, res)
	}
}

func TestIterDirection(t *testing.T) {
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < 100; i++ {
			t(i)
		}
	})
	for _, idx := range []int{0, 10, 90} {
		iter := tree.Iterator(Lt(tree.Cmp(idx)), nil)
		i := idx
		for iter.Next() {
			if iter.Item() != i {
				t.Fatalf("%d != %d", iter.Item(), i)
			}
			i++
		}
		iter = tree.Iterator(nil, Gt(tree.Cmp(idx)))
		i = idx
		for iter.Prev() {
			if iter.Item() != i {
				t.Fatalf("%d != %d", iter.Item(), i)
			}
			i--
		}
	}
	iter := tree.Iterator(nil, nil)
	i := -1
	for iter.Next() && i <= 90 {
		i++
		if iter.Item() != i {
			t.Fatalf("%d != %d", iter.Item(), i)
		}
	}
	for iter.Prev() && i >= 20 {
		i--
		if iter.Item() != i {
			t.Fatalf("%d != %d", iter.Item(), i)
		}
	}
	for iter.Next() {
		i++
		if iter.Item() != i {
			t.Fatalf("%d != %d", iter.Item(), i)
		}
	}
}

func TestReverse(t *testing.T) {
	src := rand.New(rand.NewSource(55))
	n := 1000
	tree := New[int](il, src.Perm(n)...)
	tree.root.balanced(t)
	j := 0
	iter := tree.Iterator(nil, nil)
	for iter.Next() {
		if iter.Item() != j {
			t.Fatalf("bad order")
		}
		j++
	}
	tree = tree.Reverse()
	j = n
	iter = tree.Iterator(nil, nil)
	for iter.Next() {
		j--
		if iter.Item() != j {
			t.Fatalf("bad order")
		}
	}
}

func TestRandomInsertOrder(t *testing.T) {
	src := rand.New(rand.NewSource(0))
	n := 10000
	tree := New[int](il, src.Perm(n)...)
	tree.root.balanced(t)
	j := 0
	tree.Walk(func(idx int) bool {
		if idx != j {
			t.Fatalf("bad order")
		}
		j++
		return true
	})
}

func TestRandomInsertDelete(t *testing.T) {
	n := 10000
	src := rand.New(rand.NewSource(0))
	backing := src.Perm(n)
	tree := New[int](il, backing...)
	tree.root.balanced(t)
	var idx int
	var found bool
	for i := 0; i < n; i++ {
		tree, idx, found = tree.Delete(backing[i])
		tree.root.balanced(t)
		if !found {
			t.Fatalf("Did not find %d in the Tree at %d", backing[i], i)
		}
		if idx != backing[i] {
			t.Fatalf("Error deleting: wanted %d, got %d", backing[i], idx)
		}
	}
}

func TestRandomInsertDeleteNonExistent(t *testing.T) {
	n := 100
	backing := rand.Perm(n)
	tree := New[int](il, backing...)
	tree.root.balanced(t)
	var v int
	var found bool
	if tree, v, found = tree.Delete(200); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if tree, v, found = tree.Delete(-2); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	for i := 0; i < n; i++ {
		if tree, _, found = tree.Delete(i); !found {
			t.Fatalf("remove failed for %d", i)
		}
		tree.root.balanced(t)
	}
	if tree, v, found = tree.Delete(200); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if tree, v, found = tree.Delete(-2); found {
		t.Fatalf("deleted non-existent item %d", v)
	}
	if tree.Len() != 0 {
		t.Fatalf("Failed to remove %d items!", tree.Len())
	}
}

func TestRandomInsertPartialDeleteOrder(t *testing.T) {
	n := 1000
	backing := rand.Perm(n)
	tree := New[int](il, backing...)
	tree.root.balanced(t)
	var found bool
	for i := 0; i < n; i++ {
		if tree, _, found = tree.Delete(i); !found {
			t.Fatalf("remove failed")
		}
		tree.root.balanced(t)
	}
	if tree.Len() != 0 {
		t.Fatalf("Failed to remove %d items!", tree.Len())
	}
}

func TestRandomInsertStats(t *testing.T) {
	n := 100000
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tree := New[int](il, r.Perm(n)...)
	avg, _ := tree.heightStats()
	expAvg := math.Log2(float64(n)) - 1.5
	if math.Abs(avg-expAvg) >= 1.44 {
		t.Errorf("too much deviation from expected average height")
	}
}

func TestSeqInsertStats(t *testing.T) {
	n := 100000
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < n; i++ {
			t(i)
		}
	})
	avg, _ := tree.heightStats()
	expAvg := math.Log2(float64(n)) - 1.5
	if math.Abs(avg-expAvg) >= 1.44 {
		t.Errorf("too much deviation from expected average height")
	}
}

func BenchmarkInsertIntSeqNocow(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	b.StartTimer()
	CreateWith[int](il, func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(i)
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntSeqCow(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(i)
		}
	})
	b.StartTimer()
	tree = tree.InsertWith(func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(i)
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntSeqReverseNocow(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	b.StartTimer()
	CreateWith[int](il, func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(b.N - i)
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntSeqReverseCow(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(b.N - i)
		}
	})
	b.StartTimer()
	tree = tree.InsertWith(func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(b.N - i)
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertIntRandCow(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	tree := New[int](il, 0)
	rs := rand.New(rand.NewSource(seed))
	backing := rs.Perm(b.N)
	b.StartTimer()
	tree = tree.Insert(backing...)
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteIntSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < b.N; i++ {
			t(i)
		}
	})
	b.StartTimer()
	tree.DeleteWith(func(f func(int) (int, bool)) {
		for i := 0; i < b.N; i++ {
			f(i)
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteIntRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	rs := rand.New(rand.NewSource(seed))
	vals := rs.Perm(b.N)
	tree := New[int](il, vals...)
	b.StartTimer()
	tree.DeleteItems(vals...)
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkInsertStringSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	sort.Strings(backing)
	b.StartTimer()
	New[string](sl, backing...)
	b.StopTimer()
}

func BenchmarkInsertStringRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	b.StartTimer()
	New[string](sl, backing...)
	b.StopTimer()
}

func BenchmarkDeleteStringSeq(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	sort.Strings(backing)
	tree := New[string](sl, backing...)
	b.StartTimer()
	tree.DeleteWith(func(f func(string) (string, bool)) {
		for i := 0; i < b.N; i++ {
			f(backing[i])
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkDeleteStringRand(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	seed := time.Now().Unix()
	rs := rand.New(rand.NewSource(seed))
	backing := make([]string, b.N)
	buf := [32]byte{}
	for i := range backing {
		rs.Read(buf[:])
		backing[i] = string(append([]byte{}, buf[:]...))
	}
	tree := New[string](sl, backing...)
	b.StartTimer()
	tree.DeleteWith(func(f func(string) (string, bool)) {
		for i := 0; i < b.N; i++ {
			f(backing[i])
		}
	})
	b.StopTimer()
	//ins, rms, rbi, rbr := Tree.RebalanceStats()
	//b.Logf("ins: %d, rebalances/ins: %f, rms: %d, rebalances/rm: %f", ins, rbi, rms, rbr)
}

func BenchmarkIntIterAll(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	tree := CreateWith[int](il, func(t func(int)) {
		for i := 0; i < 1<<16; i++ {
			t(i)
		}
	})
	b.StartTimer()
	i := 0
	for i < b.N {
		all := tree.Iterator(nil, nil)
		for all.Next() && i < b.N {
			if i%(1<<16) != all.Item() {
				b.Fatal(i, " != ", all.Item())
			}
			i++
		}
	}
	b.StopTimer()
}

func BenchmarkFetch(b *testing.B) {
	for _, sz := range []int{1 << 4, 1 << 8, 1 << 16, 1 << 24} {
		b.Run(fmt.Sprintf("btree size %d", sz), func(b *testing.B) {
			b.StopTimer()
			b.ReportAllocs()
			tree := CreateWith[int](il, func(t func(int)) {
				for i := 0; i < sz; i++ {
					t(i)
				}
			})
			fetched := 0
			items := rand.Perm(sz)
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := tree.Fetch(items[i%sz] << 1); ok {
					fetched++
				}
			}
			b.StopTimer()
		})
		b.Run(fmt.Sprintf("map size %d", sz), func(b *testing.B) {
			b.StopTimer()
			m := map[int]struct{}{}
			for i := 0; i < sz; i++ {
				m[i] = struct{}{}
			}
			items := rand.Perm(sz)
			fetched := 0
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := m[items[i%sz]<<1]; ok {
					fetched++
				}
			}

		})
	}
}

func TestAscendAfter(t *testing.T) {
	tree := New[int](il, 4, 6, 1, 3)
	var ary, expected []int
	ary = nil
	// inclusive
	tree.After(Lte(tree.Cmp(-1)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{1, 3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	// inclusive
	tree.After(Lt(tree.Cmp(3)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	// exclusive
	tree.After(Lte(tree.Cmp(3)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.After(Lt(tree.Cmp(2)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
}

func TestAscendBefore(t *testing.T) {
	tree := New[int](il, 4, 6, 1, 3)
	var ary []int
	tree.Before(Gt(tree.Cmp(10)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected := []int{1, 3, 4, 6}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.Before(Gte(tree.Cmp(4)), func(idx int) bool {
		ary = append(ary, idx)
		return true
	})
	expected = []int{1, 3}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
	ary = nil
	tree.Before(Gt(tree.Cmp(4)), func(i int) bool {
		ary = append(ary, i)
		return true
	})
	expected = []int{1, 3, 4}
	if !reflect.DeepEqual(ary, expected) {
		t.Fatalf("expected %v but got %v", expected, ary)
	}
}

type ovr struct {
	i, mark int
}

func ol(a, b ovr) bool { return a.i < b.i }

func TestCopyOnWriteRace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tree1 := CreateWith[ovr](ol, func(t func(ovr)) {
		for i := 0; i < 200; i++ {
			t(ovr{i: i, mark: 1})
		}
	})
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				i := 0
				iter := tree1.Iterator(nil, nil)
				for iter.Next() {
					item := iter.Item()
					if item.i != i || i >= 200 || item.mark != 1 {
						t.Errorf("Tree 1 has bleed over from %d at %d", item.mark, item.i)
						return
					}
					i++
				}
				if i != 200 {
					t.Errorf("Tree 1 Iteration ended at %d, not 200", i)
					return
				}
			}
		}
	}()
	tree2 := tree1.InsertWith(func(t func(ovr)) {
		for i := 100; i < 300; i++ {
			t(ovr{i: i, mark: 2})
		}
	})
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				iter := tree2.Iterator(nil, nil)
				i := 0
				for iter.Next() {
					item := iter.Item()
					if item.i != i || item.i >= 300 || item.mark == 3 {
						t.Errorf("Tree 2 has bleed over from %d at %d", item.mark, item.i)
						return
					}
					if i < 100 && item.mark != 1 {
						t.Errorf("Tree 2 Mark on item %d is %d, not 1", i, item.mark)
						return
					}
					if i >= 100 && item.mark != 2 {
						t.Errorf("Tree 2 Mark on item %d is %d, not 2", i, item.mark)
						return
					}
					i++
				}
				if i != 300 {
					t.Errorf("Tree 2 Iteration ended at %d, not 300", i)
					return
				}
			}
		}
	}()
	tree3 := tree1.InsertWith(func(t func(ovr)) {
		for i := -100; i < 100; i++ {
			t(ovr{i: i, mark: 3})
		}
	})
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				iter := tree3.Iterator(nil, nil)
				i := -100
				for iter.Next() {
					item := iter.Item()
					if item.i != i || item.i >= 200 || item.mark == 2 {
						t.Errorf("Tree 3 has bleed over from %d at %d", item.mark, item.i)
						return
					}
					if i < 100 && item.mark != 3 {
						t.Errorf("Tree 3 Item at %d has mark %d, not 3", i, item.mark)
						return
					}
					if i >= 100 && item.mark != 1 {
						t.Errorf("Tree 3 Item at %d has mark %d, not 1", i, item.mark)
						return
					}
					i++
				}
				if i != 200 {
					t.Errorf("Tree 3 Iteration ended at %d, not 200", i)
					return
				}
			}
		}
	}()
	time.Sleep(time.Second)
	tree3.InsertWith(func(t func(ovr)) {
		for i := -100; i < 200; i++ {
			t(ovr{i: i, mark: 4})
		}
	})
	time.Sleep(time.Second)
	for i := -100; i < 400; i++ {
		tree1.Delete(ovr{i: i})
		tree2.Delete(ovr{i: i})
		tree3.Delete(ovr{i: i})
	}
	time.Sleep(time.Second)
	cancel()
	wg.Wait()
}
