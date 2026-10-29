package codecs

import (
	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticAdaptiveCodec struct {
	frequencies map[rune]int
}

func NewArithmeticAdaptiveCodec() *ArithmeticAdaptiveCodec {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &ArithmeticAdaptiveCodec{
		frequencies: frequencies,
	}
}

func (c *ArithmeticAdaptiveCodec) Name() string {
	return "Arithmetic (adaptive)"
}

func (c *ArithmeticAdaptiveCodec) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	// Arithmetic coding.
	code := arithmetic.EncodeAdaptive(piecesToRunes(pieces), bitfield.IntToBoolSlice(occupancy, 64), c.frequencies, bitfield.IntToBoolSlice(uint64(color), len(pieces)), state)
	bits.WriteInt(uint64(len(code)), arithmetic.BitsForArithmeticCodeLength)
	bits.WriteBoolSlice(code)

	return bits
}

func (c *ArithmeticAdaptiveCodec) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement: pieces.
	l := int(bits.ReadInt(arithmetic.BitsForArithmeticCodeLength))
	code := bits.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(code, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
