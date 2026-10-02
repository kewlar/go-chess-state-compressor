package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/huffman"
)

type HuffmanCoder struct {
	huffmanTree *huffman.Node
	huffmanDict map[rune]*huffman.Node
}

func NewHuffmanCoder() *HuffmanCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	huffmanTree := huffman.BuildHuffmanTree(frequencies)

	return &HuffmanCoder{
		huffmanTree: huffmanTree,
		huffmanDict: huffmanTree.BuildDictionary(),
	}
}

func (c *HuffmanCoder) Name() string {
	return "Huffman"
}

func (c *HuffmanCoder) FromFen(fen string) *bitfield.BitField {
	state, _ := chess.ParseFen(fen)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy := uint64(0)
	color := uint32(0)
	pieces := []rune{}
	for cell, r := range state.Board {
		if r == 0 {
			continue
		}

		pieceType := rune(int(r) | 32)

		occupancy |= 1 << cell

		isWhite := int(r)&32 == 0
		if isWhite {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, pieceType)
	}

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteBoolSlice(c.huffmanDict[piece].Bits)
	}

	// Side to move.
	bf.WriteBool(state.SideToMove == 'w')

	// Castling availability.
	for _, option := range chess.CastlingOptions {
		bf.WriteBool(state.Castling[option])
	}

	// En passant square.
	bf.WriteInt(uint64(state.EnPassant), 7)

	// Halfmove clock.
	bf.WriteInt(uint64(state.Halfmove), 8)

	// Fullmove number.
	bf.WriteInt(uint64(state.Fullmove), 14)

	return bf
}

func (c *HuffmanCoder) ToFen(data *bitfield.BitField) string {
	state := chess.NewState()

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
	if isWhite {
		state.SideToMove = 'w'
	} else {
		state.SideToMove = 'b'
	}

	// Castling availability.
	for _, option := range chess.CastlingOptions {
		state.Castling[option] = data.ReadBool()
	}

	// En passant square.
	state.EnPassant = int(data.ReadInt(7))

	// Halfmove clock.
	state.Halfmove = int(data.ReadInt(8))

	// Fullmove number.
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement.
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			piece, color := pieces[idx], colors[idx]
			if color { // White
				piece &= 0b1101_1111
			}
			state.Board[i] = piece
			idx++
		}
	}

	return state.ToFen()
}
