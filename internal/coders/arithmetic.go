package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticCoder struct {
	frequencies map[rune]int
	intervals   map[rune]*arithmetic.Interval
}

const bitsForArithmeticCodeLength = 7

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
	bf := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))

	// Arithmetic coding.
	bits := arithmetic.Encode(piecesToRunes(pieces), c.intervals)
	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

	bf.WriteBool(bool(state.SideToMove))
	bf.WriteBoolSlice(state.Castling.BoolSlice())
	bf.WriteInt(uint64(state.EnPassant), 7)
	bf.WriteInt(uint64(state.Halfmove), 8)
	bf.WriteInt(uint64(state.Fullmove), 14)

	return bf
}

func (c *ArithmeticCoder) Decode(data *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.Decode(bits, pieceCount, c.intervals)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	state.SideToMove = chess.Color(data.ReadBool())
	state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
	state.EnPassant = chess.Square(data.ReadInt(7))
	state.Halfmove = int(data.ReadInt(8))
	state.Fullmove = int(data.ReadInt(14))

	return state
}
