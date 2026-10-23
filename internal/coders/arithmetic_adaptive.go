package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticAdaptiveCoder struct {
	frequencies map[rune]int
}

func NewArithmeticAdaptiveCoder() *ArithmeticAdaptiveCoder {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &ArithmeticAdaptiveCoder{
		frequencies: frequencies,
	}
}

func (c *ArithmeticAdaptiveCoder) Name() string {
	return "Arithmetic (adaptive)"
}

func (c *ArithmeticAdaptiveCoder) Encode(state *chess.State) *bitfield.BitField {
	bf := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))

	bf.WriteBool(bool(state.SideToMove))
	bf.WriteBoolSlice(state.Castling.BoolSlice())
	bf.WriteInt(uint64(state.EnPassant), 7)
	bf.WriteInt(uint64(state.Halfmove), 8)
	bf.WriteInt(uint64(state.Fullmove), 14)

	// Arithmetic coding.
	code := arithmetic.EncodeAdaptive(piecesToRunes(pieces), bitfield.IntToBoolSlice(occupancy, 64), c.frequencies, bitfield.IntToBoolSlice(uint64(color), len(pieces)), state)
	if len(code) > maxBitsForArithmeticCode {
		log.Fatalf("encoded data too long (%d bits)", len(code))
	}
	bf.WriteInt(uint64(len(code)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(code)

	return bf
}

func (c *ArithmeticAdaptiveCoder) Decode(data *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	state.SideToMove = chess.Color(data.ReadBool())
	state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
	state.EnPassant = chess.Square(data.ReadInt(7))
	state.Halfmove = int(data.ReadInt(8))
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	code := data.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(code, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
