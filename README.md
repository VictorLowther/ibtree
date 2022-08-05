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
    BenchmarkInsertIntSeqNocow-10           	 6181240	       224.5 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntSeqCow-10             	 3437576	       387.5 ns/op	      89 B/op	       1 allocs/op
    BenchmarkInsertIntSeqReverseNocow-10    	 5719492	       233.2 ns/op	      32 B/op	       1 allocs/op
    BenchmarkInsertIntSeqReverseCow-10      	 3429450	       381.8 ns/op	      89 B/op	       1 allocs/op
    BenchmarkInsertIntRandCow-10            	 1542794	       837.5 ns/op	      66 B/op	       1 allocs/op
    BenchmarkDeleteIntSeq-10                	 4191504	       295.0 ns/op	      32 B/op	       1 allocs/op
    BenchmarkDeleteIntRand-10               	 1610540	       839.3 ns/op	      49 B/op	       1 allocs/op
    BenchmarkInsertStringSeq-10             	 7300251	       193.7 ns/op	      48 B/op	       1 allocs/op
    BenchmarkInsertStringRand-10            	 1592608	       926.7 ns/op	      48 B/op	       1 allocs/op
    BenchmarkDeleteStringSeq-10             	 3384838	       348.7 ns/op	      48 B/op	       1 allocs/op
    BenchmarkDeleteStringRand-10            	 1000000	        1066 ns/op	      73 B/op	       1 allocs/op
    BenchmarkIntIterAll-10                  	203883426	       5.866 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/btree_size_16-10         	100000000	       11.15 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_16-10           	177492420	       6.677 ns/op
    BenchmarkFetch/btree_size_256-10        	68811284	       17.10 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_256-10          	140613560	       7.681 ns/op
    BenchmarkFetch/btree_size_65536-10      	35249148	       34.58 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_65536-10        	60658269	       19.38 ns/op
    BenchmarkFetch/btree_size_16777216-10   	16971104	       70.74 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFetch/map_size_16777216-10     	21009904	       60.14 ns/op
    PASS
    ok  	github.com/VictorLowther/ibtree	80.137s

Interestingly enough, the slowdown on the random benchmarks appears to be due to
branch misprediction rather than tree rebalancing performing more work -- dealing
with sorted data actually performs more rebalancing than random data.
