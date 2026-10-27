package huffman

import (
	"fmt"
	"slices"

	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
)

type Node struct {
	Bits      []bool
	Value     rune
	Frequency int
	Children  map[bool]*Node
	Parent    *Node
}

func BuildHuffmanTree(frequencies map[rune]int) *Node {
	leaves := []*Node{}
	for r, n := range frequencies {
		if n > 0 {
			leaf := &Node{
				Value:     r,
				Frequency: n,
			}
			leaves = append(leaves, leaf)
		}
	}
	for len(leaves) > 1 {
		slices.SortFunc(leaves, func(a, b *Node) int {
			if a.Frequency < b.Frequency {
				return -1
			}
			if a.Frequency > b.Frequency {
				return 1
			}
			if a.Value < b.Value {
				return -1
			}
			if a.Value > b.Value {
				return 1
			}
			return 0
		})
		a, b := leaves[0], leaves[1]
		parent := &Node{
			Frequency: a.Frequency + b.Frequency,
			Children:  map[bool]*Node{},
		}
		a.prependBit(false)
		b.prependBit(true)
		a.Parent = parent
		b.Parent = parent
		parent.Children[false] = a
		parent.Children[true] = b
		leaves = append(leaves[2:], parent)
	}

	return leaves[0]
}

func (node *Node) prependBit(bit bool) {
	if node.IsLeaf() {
		node.Bits = append([]bool{bit}, node.Bits...)
	} else {
		node.Children[false].prependBit(bit)
		node.Children[true].prependBit(bit)
	}
}

func (node *Node) IsLeaf() bool {
	return node.Children == nil
}

func (tree *Node) BuildDictionary() map[rune]*Node {
	dict := map[rune]*Node{}
	var node *Node
	queue := []*Node{tree}
	for len(queue) > 0 {
		node, queue = queue[0], queue[1:]
		if node.IsLeaf() {
			dict[node.Value] = node
		} else {
			queue = append(queue, node.Children[false], node.Children[true])
		}
	}

	return dict
}

func EncodeAdaptive(data []rune, occupancy []bool, frequencies map[rune]int, color []bool, state *chess.State) []bool {
	bits := []bool{}
	n := len(data)
	idx := 0
	freqs := entropy.PrepareFrequencies(n, frequencies)
	for i := range 64 {
		if occupancy[i] {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(i), n, occupancy, color, state, data[:idx])
			piece := data[idx]
			if freqs2[piece] == 0 {
				panic(fmt.Sprintf("didn't expect '%s' at square %s", string(piece), chess.Square(i).String()))
			}
			tree := BuildHuffmanTree(freqs2)
			dict := tree.BuildDictionary()
			bits = append(bits, dict[piece].Bits...)
			idx++

			freqs = entropy.UpdateFrequencies(piece, n-idx, freqs)
		}
	}
	return bits
}

func DecodeAdaptive(readBit func() bool, n int, occupancy []bool, frequencies map[rune]int, color []bool, state *chess.State) []rune {
	pieces := make([]rune, 0, n)
	freqs := entropy.PrepareFrequencies(n, frequencies)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(square), n, occupancy, color, state, pieces)
			tree := BuildHuffmanTree(freqs2)

			node := tree
			for !node.IsLeaf() {
				bit := readBit()
				node = node.Children[bit]
			}
			piece := node.Value
			pieces = append(pieces, piece)

			idx++

			freqs = entropy.UpdateFrequencies(piece, n-idx, freqs)
		}
	}
	return pieces
}
