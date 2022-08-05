package ibtree

import (
	"math"
)

// getKeyHeight returns an item in the Tree with key @key, and it's height in the Tree
func (t *Tree[T]) getKeyHeight(key CompareAgainst[T]) (result T, depth int) {
	return t.getHeight(t.root, key)
}

func (t *Tree[T]) getHeight(h *node[T], item CompareAgainst[T]) (T, int) {
	if h == nil {
		var ref T
		return ref, 0
	}
	switch item(h.i) {
	case -1:
		result, depth := t.getHeight(h.l, item)
		return result, depth + 1
	case 1:
		result, depth := t.getHeight(h.l, item)
		return result, depth + 1
	default:
		return h.i, 0
	}
}

// heightStats returns the average and standard deviation of the height
// of elements in the Tree
func (t *Tree[T]) heightStats() (avg, stddev float64) {
	av := &AvgVar{}
	heightStats(t.root, 0, av)
	return av.GetAvg(), av.GetStdDev()
}

func heightStats[T any](h *node[T], d int, av *AvgVar) {
	if h == nil {
		return
	}
	av.Add(float64(d))
	if h.l != nil {
		heightStats(h.l, d+1, av)
	}
	if h.r != nil {
		heightStats(h.r, d+1, av)
	}
}

// AvgVar maintains the average and variance of a stream of numbers
// in a space-efficient manner.
type AvgVar struct {
	count      int64
	sum, sumsq float64
}

func (av *AvgVar) Init() {
	av.count = 0
	av.sum = 0.0
	av.sumsq = 0.0
}

func (av *AvgVar) Add(sample float64) {
	av.count++
	av.sum += sample
	av.sumsq += sample * sample
}

func (av *AvgVar) GetCount() int64 { return av.count }

func (av *AvgVar) GetAvg() float64 { return av.sum / float64(av.count) }

func (av *AvgVar) GetTotal() float64 { return av.sum }

func (av *AvgVar) GetVar() float64 {
	a := av.GetAvg()
	return av.sumsq/float64(av.count) - a*a
}

func (av *AvgVar) GetStdDev() float64 { return math.Sqrt(av.GetVar()) }
