package coders

import (
	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticCoder struct {
	frequencies map[rune]int
	intervals   map[rune]*arithmetic.Interval
}

func NewArithmeticCoder() *ArithmeticCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	return &ArithmeticCoder{
		frequencies: frequencies,
		intervals:   arithmetic.FrequenciesToIntervals(frequencies),
	}
}

func (c *ArithmeticCoder) Name() string {
	return "Arithmetic"
}

func (c *ArithmeticCoder) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))

	// Arithmetic coding.
	code := arithmetic.Encode(piecesToRunes(pieces), c.intervals)
	bits.WriteInt(uint64(len(code)), arithmetic.BitsForArithmeticCodeLength)
	bits.WriteBoolSlice(code)

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	return bits
}

func (c *ArithmeticCoder) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	l := int(bits.ReadInt(arithmetic.BitsForArithmeticCodeLength))
	code := bits.ReadBoolSlice(l)
	pieces := arithmetic.Decode(code, pieceCount, c.intervals)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	return state
}
