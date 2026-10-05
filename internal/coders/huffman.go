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
	pieces := []chess.Piece{}
	for cell, piece := range state.Board {
		if piece.IsEmpty() {
			continue
		}

		occupancy |= 1 << cell
		if piece.Color() == chess.White {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, piece)
	}

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteBoolSlice(c.huffmanDict[rune(piece.Type())].Bits)
	}

	// Side to move.
	bf.WriteBool(bool(state.SideToMove))

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
	state.SideToMove = chess.Color(data.ReadBool())

	// Castling availability.
	for _, option := range chess.CastlingOptions {
		state.Castling[option] = data.ReadBool()
	}

	// En passant square.
	state.EnPassant = chess.Square(data.ReadInt(7))

	// Halfmove clock.
	state.Halfmove = int(data.ReadInt(8))

	// Fullmove number.
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement.
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			state.Board[i] = chess.NewPiece(pieces[idx], chess.Color(colors[idx]))
			idx++
		}
	}

	return state.ToFen()
}
