# ibtree

ibtree is an implementation of generic immutable balanced binary trees in Go. 
This packages provides generic immutable AVL trees. 

## Why

Mutable btrees are all well and good, but sometimes you need a btree that can be updated
and accessed at the same time in multiple different goroutines.  Sure, you could invent
a complicated locking scheme using ever more finegrained and deadlock prone locking schemes.
There are libraries out there that provide that.  This is not one of them.

Instead, this library provides immutable btrees.  Any operation that would mutate the tree will
instead make copies of any nodes that would be changed and return a new tree.  The new tree and
the old tree will share unchanged nodes.  This library also provides bulk insert and delete operations
that minimize the amount of node copying that happens under the hood.

## Installation

go get https://github.com/VictorLowther/ibtree

## Example

    package main
    import github.com/VictorLowther/ibtree
    import fmt

    func main() {
        tree := ibtree.New[int](func(a,b int) {return a < b})
        tree = tree.Insert(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
        tree = tree.Reverse()
        iter := tree.Iterate(nil, nil)
        for iter.Next() {
            fmt.Println(iter.Item())
        }
    }

## Benchmarks:

On a Macbook Pro M1 Max:

    % go test -bench .
    goos: darwin
    goarch: arm64
    pkg: github.com/VictorLowther/ibtree
    BenchmarkInsertIntSeqNocow-10           	 6304920	       214.8 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntSeqCow-10             	 7593422	       176.1 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntSeqReverseNocow-10    	 6726402	       225.3 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntSeqReverseCow-10      	 7352782	       175.6 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntRandCow-10            	 2879599	       731.1 ns/op	      32 B/op	       1 allocs/op
    BenchmarkDeleteIntSeq-10                	 8109351	       159.6 ns/op	      32 B/op	       1 allocs/op
    BenchmarkDeleteIntRand-10               	 2684017	       640.3 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertStringSeq-10             	 6578706	       189.6 ns/op	      48 B/op	       1 allocs/op
    BenchmarkInsertStringRand-10            	 1586451	       933.3 ns/op	      48 B/op	       1 allocs/op
    BenchmarkDeleteStringSeq-10             	 6984127	       187.3 ns/op	      48 B/op	       1 allocs/op
    BenchmarkDeleteStringRand-10            	 1548794	       877.8 ns/op	      48 B/op	       1 allocs/op
    BenchmarkIntIterAll-10                     181415104	       6.618 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/btree_size_16-10            100000000	       10.49 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_16-10              201185232	       6.035 ns/op
    BenchmarkFetch/btree_size_256-10        	50292850	       21.89 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_256-10             149541414	       8.055 ns/op
    BenchmarkFetch/btree_size_65536-10      	11741773	       105.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_65536-10        	61229953	       19.14 ns/op
    BenchmarkFetch/btree_size_16777216-10   	 1714615	       691.6 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_16777216-10     	21412989	       57.05 ns/op
    PASS
    ok  	github.com/VictorLowther/ibtree	86.026s

Interestingly enough, the slowdown on the random benchmarks appears to be due to
branch misprediction rather than tree rebalancing performing more work -- dealing
with sorted data actually performs more rebalancing than random data.
