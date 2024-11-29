// Package huffman implements an n-ary Huffman coding algorithm.
//
// It can be used to generate prefix-free labels for any number of items.
// Prefix-free labels are labels where for any two labels X and Y,
// there's a guarantee that X is not a prefix of Y.
// This is a useful property because it allows for unambiguous matching.
// When receiving incremental input (e.g. keystrokes),
// as soon as the input matches a prefix-free label X,
// you can stop and process the item corresponding to X.
package huffman

import (
	"container/heap"
	"slices"
)

// Label generates unique prefix-free labels for items given their frequencies.
//
// base is the number of symbols in the alphabet.
// For example, for a binary alphabet, base should be 2.
// base=26 is a common choice for alphabetic labels.
// The mapping from base to alphabet is the caller's responsibility.
// base must be at least 2.
//
// For len(freqs) items, freqs[i] specifies the frequency of item i.
// Items with higher frequencies will be assigned shorter labels.
//
// The returned value is a list of labels for each item.
// labels[i] is the label for item i,
// specified as indexes into the alphabet.
// For example, given a binary alphabet {a b},
// the label {0 1 0} means "aba".
func Label(base int, freqs []int) (labels [][]int) {
	// This implements Huffman coding with a priority queue
	// using the method outlined on Wikipedia [1],
	// altering it for n-ary trees also using advice from the same page.
	// Further advice on optizing comes from rdbliss/huffman [2].
	//
	// [1]: https://en.wikipedia.org/wiki/Huffman_coding#Basic_technique
	// [2]: https://github.com/rdbliss/huffman/

	if base < 2 {
		panic("alphabet must have at least two elements")
	}

	switch len(freqs) {
	case 0:
		return nil
	case 1:
		// special-case:
		// If there's only one item, create a single letter label.
		return [][]int{{0}}
	}

	// MATHS //////////////////////////////////////////////////////////////
	//
	// 1. Number of iterations
	//
	// There are C = len(freqs) items.
	// Each item gets a node, so that's C nodes to start with.
	//
	// We will combine nodes in the heap. Each such combination:
	//
	//   - removes $base nodes from the heap
	//   - adds one new node to the heap
	//
	// Resulting in a net $base-1 reduction per iteration.
	//
	// We'll iterate until there's only one node left in the heap.
	//
	// Suppose the total number of iterations is I.
	// Starting with C nodes (one for each item),
	// and removing $base-1 nodes per iteration for I iterations,
	// we're left with one node in the heap.
	//
	//   1 = C - I($base-1)
	//
	// Solving for I:
	//
	//   1 = C - I($base-1)
	//   => I($base-1) = C - 1
	//   => I = (C - 1) / ($base - 1)
	//
	// I is also the maximum possible depth of the tree
	// since each iteration adds one level to the tree.
	//
	// 2. Numbef of nodes
	//
	// Total number of nodes we need to allocate (N)
	// is the original C plus one for each iteration.
	//
	//    N = C + I
	//    N = C + (C - 1) / ($base - 1)
	//
	// /However/ if the number of items is such that we remove fewer
	// than $base nodes in one iteration, we may have one extra iteration.
	// Easy way to do that is to pad the number of iterations by 1.
	//
	///////////////////////////////////////////////////////////////////////
	numIters := (len(freqs)-1)/(base-1) + 1
	numNodes := len(freqs) + numIters

	nodes := make([]node, numNodes) // one allocation for all nodes
	for i, f := range freqs {
		nodes[i] = node{
			Freq:         f,
			ParentIndex:  -1,
			SiblingIndex: -1,
		}
	}
	nextNodeIdx := len(freqs)

	// Fill the heap with leaf nodes for the user-provided elements.
	nodeHeap := make(nodeHeap, len(freqs))
	for i := range freqs {
		nodeHeap[i] = &nodes[i]
	}
	heap.Init(&nodeHeap)

	// This is the meat of the logic.
	//
	//  - Assign letters [0, $base) to the least frequent items
	//    and remove them from the heap.
	//  - Create a new node that represents these $base items
	//    and push it back into the heap.
	//  - Repeat until there's only one node left in the heap.
	//
	// We'll end up with a tree where each node has up to $base children.
	// The path from root down to leaf nodes will is the label for that item.
	combine := func(numChildren int) {
		parentIdx := nextNodeIdx
		nextNodeIdx++

		var freq, totalHops int
		for i := 0; i < numChildren && len(nodeHeap) > 0; i++ {
			child := heap.Pop(&nodeHeap).(*node)
			child.ParentIndex = parentIdx
			child.SiblingIndex = i
			freq += child.Freq
			totalHops += child.TotalHops
		}

		nodes[parentIdx] = node{
			ParentIndex:  -1,
			SiblingIndex: -1,
			Freq:         freq,
			TotalHops:    totalHops + numChildren,
		}
		heap.Push(&nodeHeap, &nodes[parentIdx])
	}

	// Special-case: for the first iteration,
	// assign fewer letters to the least frequent items.
	// This will ensure high frequency nodes don't unnecessarily
	// get longer labels because a couple extra nodes pushed them
	// over the edge of $base, requiring a new branch above.
	//
	// See https://github.com/rdbliss/huffman/blob/master/notes.md#generalization
	initial := 2 + (len(nodeHeap)-2)%(base-1)
	if initial > 0 {
		combine(initial)
	}

	for len(nodeHeap) > 1 {
		combine(base)
	}

	// nodeHeap only has one node left.
	totalHops := nodeHeap[0].TotalHops

	// The first len(freqs) nodes in nodes list refer to the leaf nodes.
	// These get labels assigned to them.
	labels = make([][]int, len(freqs))

	// One slice for all label data.
	// Labels for all items will be subslices of this.
	//
	//   len(freqs) items x max depth of the tree
	//   = len(freqs) x numIters
	labelData := make([]int, 0, totalHops)

	for idx, n := range nodes[:len(freqs)] {
		// The label for the item is the path from the root to the leaf.
		start := len(labelData)
		for c := n; c.ParentIndex != -1; c = nodes[c.ParentIndex] {
			labelData = append(labelData, c.SiblingIndex)
		}
		label := labelData[start:]

		// Reverse the label so it goes from root to leaf.
		slices.Reverse(label)
		labels[idx] = label
	}

	return labels
}

// node is used in two places:
//
//   - as part of min-heap sorted by Freq
//   - as part of a tree structure
//
// ParentID, and SiblingIndex form the tree structure.
// Freq is used for the min-heap.
type node struct {
	// Index of parent node in the nodes list.
	ParentIndex int

	// SiblingIndex of the node is its position
	// in the parent's children list.
	SiblingIndex int

	// TotalHops is the total number of hops from this node
	// to all leaf nodes under it.
	//
	// It is used to optimize the label generation.
	TotalHops int

	// Frequency of the leaf node,
	// or the combined frequency of the leaf nodes of a branch node.
	Freq int
}

type nodeHeap []*node

func (ns nodeHeap) Len() int { return len(ns) }

func (ns nodeHeap) Less(i, j int) bool {
	return ns[i].Freq < ns[j].Freq
}

func (ns nodeHeap) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns *nodeHeap) Push(e interface{}) {
	*ns = append(*ns, e.(*node))
}

func (ns *nodeHeap) Pop() interface{} {
	n := len(*ns) - 1
	v := (*ns)[n]
	*ns = (*ns)[:n]
	return v
}
