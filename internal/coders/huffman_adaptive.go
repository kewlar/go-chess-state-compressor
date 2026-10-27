package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanAdaptiveCoder struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCoder() *HuffmanAdaptiveCoder {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &HuffmanAdaptiveCoder{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCoder) Name() string {
	return "Huffman (adaptive)"
}

func (c *HuffmanAdaptiveCoder) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	// Huffman coding.
	fullOccupancy := bitfield.IntToBoolSlice(occupancy, 64)
	colorSlice := bitfield.IntToBoolSlice(uint64(color), len(pieces))
	bits.WriteBoolSlice(huffman.EncodeAdaptive(piecesToRunes(pieces), fullOccupancy, c.frequencies, colorSlice, state))

	return bits
}

func (c *HuffmanAdaptiveCoder) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := huffman.DecodeAdaptive(func() bool { return bits.ReadBool() }, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
