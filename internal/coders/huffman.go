package coders

import (
	"slices"
	"strconv"
	"strings"

	"example.com/go-chess/internal/bitfield"
)

type HuffmanCoder struct {
	castlingOptions []rune
	huffmanTree     *HuffmanNode
	huffmanDict     map[rune]*HuffmanNode
}

func NewHuffmanCoder() *HuffmanCoder {
	var (
		castlingOptions = []rune{'K', 'Q', 'k', 'q'}
		frequencies     = map[rune]int{
			'p': 16,
			'k': 2,
			'q': 2,
			'r': 4,
			'b': 4,
			'n': 4,
		}
	)

	huffmanTree := buildHuffmanTree(frequencies)
	huffmanDict := buildHuffmanDict(huffmanTree)

	return &HuffmanCoder{
		castlingOptions: castlingOptions,
		huffmanTree:     huffmanTree,
		huffmanDict:     huffmanDict,
	}
}

func (c *HuffmanCoder) Name() string {
	return "Huffman"
}

type HuffmanNode struct {
	Bits      []bool
	Value     rune
	Frequency int
	Children  map[bool]*HuffmanNode
	Parent    *HuffmanNode
}

func buildHuffmanTree(frequencies map[rune]int) *HuffmanNode {
	leaves := []*HuffmanNode{}
	for r, n := range frequencies {
		leaf := &HuffmanNode{
			Value:     r,
			Frequency: n,
		}
		leaves = append(leaves, leaf)
	}
	for len(leaves) > 1 {
		slices.SortFunc(leaves, func(a, b *HuffmanNode) int {
			if a.Frequency < b.Frequency {
				return -1
			}
			if a.Frequency > b.Frequency {
				return 1
			}
			return 0
		})
		a, b := leaves[0], leaves[1]
		parent := &HuffmanNode{
			Frequency: a.Frequency + b.Frequency,
			Children:  map[bool]*HuffmanNode{},
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

func (node *HuffmanNode) prependBit(bit bool) {
	if node.IsLeaf() {
		node.Bits = append([]bool{bit}, node.Bits...)
	} else {
		node.Children[false].prependBit(bit)
		node.Children[true].prependBit(bit)
	}
}

func (node *HuffmanNode) IsLeaf() bool {
	return node.Children == nil
}

func buildHuffmanDict(tree *HuffmanNode) map[rune]*HuffmanNode {
	dict := map[rune]*HuffmanNode{}
	var node *HuffmanNode
	queue := []*HuffmanNode{tree}
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

func (c *HuffmanCoder) FromFen(fen string) *bitfield.BitField {
	// PARSE FEN

	parts := strings.Fields(fen)

	// Piece placement
	occupancy := uint64(0)
	color := uint32(0)
	pieces := []rune{}
	cell := 0
	for _, r := range parts[0] {
		if r == '/' {
			continue
		}

		if '1' <= r && r <= '8' {
			cell += int(r - '0')
			continue
		}

		pieceType := rune(int(r) | 32)

		occupancy |= 1 << cell

		isWhite := int(r)&32 == 0
		if isWhite {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, pieceType)

		cell++
	}

	// Side to move.
	whiteToMove := parts[1] == "w"

	// Castling availability.
	castling := []bool{}
	for _, option := range c.castlingOptions {
		castling = append(castling, strings.ContainsRune(parts[2], option))
	}

	// En passant square.
	enPassant := byte(64) // 64 == no en passant
	if parts[3] != "-" {
		file := parts[3][0] - 'a'
		rank := parts[3][1] - '1'
		enPassant = (7-rank)*8 + file
	}

	// Halfmove clock.
	halfMoves, _ := strconv.Atoi(parts[4])

	// Fullmove number.
	fullMoves, _ := strconv.Atoi(parts[5])

	// GENERATE BITS

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteBoolSlice(c.huffmanDict[piece].Bits)
	}

	// Side to move.
	bf.WriteBool(whiteToMove)

	// Castling availability.
	for idx := range c.castlingOptions {
		bf.WriteBool(castling[idx])
	}

	// En passant square.
	bf.WriteInt(uint64(enPassant), 7)

	// Halfmove clock.
	bf.WriteInt(uint64(halfMoves), 8)

	// Fullmove number.
	bf.WriteInt(uint64(fullMoves), 14)

	return bf
}

func (c *HuffmanCoder) ToFen(data *bitfield.BitField) string {
	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := 0
	for _, occupied := range occupancy {
		if occupied {
			pieceCount++
		}
	}

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		// Read Huffman code.
		node := c.huffmanTree
		for !node.IsLeaf() {
			bit := data.ReadBool()
			node = node.Children[bit]
		}
		pieces = append(pieces, node.Value)
	}

	// Side to move.
	isWhite := data.ReadBool()

	// Castling availability.
	castling := map[rune]bool{}
	for _, option := range c.castlingOptions {
		castling[option] = data.ReadBool()
	}

	// En passant square.
	enPassant := data.ReadInt(7)

	// Halfmove clock.
	halfMoves := data.ReadInt(8)

	// Fullmove number.
	fullMoves := data.ReadInt(14)

	// GENERATE FEN

	fen := strings.Builder{}

	// Piece placement.
	empty := 0
	idx := 0
	for i, occupied := range occupancy {
		if i&7 == 0 && i != 0 {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			fen.WriteRune('/')
		}
		if occupied {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			piece, color := pieces[idx], colors[idx]
			if color { // White
				piece &= 0b1101_1111
			}
			fen.WriteRune(piece)
			idx++
		} else {
			empty++
		}
	}
	if empty > 0 {
		fen.WriteRune('0' + rune(empty))
	}
	fen.WriteRune(' ')

	// Side to move.
	if isWhite {
		fen.WriteRune('w')
	} else {
		fen.WriteRune('b')
	}
	fen.WriteRune(' ')

	// Castling availability.
	none := true
	for _, option := range c.castlingOptions {
		if castling[option] {
			fen.WriteRune(option)
			none = false
		}
	}
	if none {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// En passant square.
	if enPassant <= 63 {
		rank := 7 - (enPassant >> 3)
		file := enPassant & 7
		fen.WriteRune('a' + rune(file))
		fen.WriteRune('1' + rune(rank))
	} else {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// Halfmove clock.
	fen.WriteString(strconv.Itoa(int(halfMoves)))
	fen.WriteRune(' ')

	// Fullmove number.
	fen.WriteString(strconv.Itoa(int(fullMoves)))

	return fen.String()
}
