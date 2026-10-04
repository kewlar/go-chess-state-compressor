package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type ArithmeticAdaptiveCoder struct {
	frequencies map[rune]int
}

func NewArithmeticAdaptiveCoder() *ArithmeticAdaptiveCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	return &ArithmeticAdaptiveCoder{
		frequencies: frequencies,
	}
}

func (c *ArithmeticAdaptiveCoder) Name() string {
	return "Arithmetic (adaptive)"
}

func (c *ArithmeticAdaptiveCoder) FromFen(fen string) *bitfield.BitField {
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

	// Arithmetic coding.
	bits := arithmetic.EncodeAdaptive(pieces, bitfield.IntToBoolSlice(occupancy, 64), c.frequencies, bitfield.IntToBoolSlice(uint64(color), len(pieces)), state)

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

	return bf
}

func (c *ArithmeticAdaptiveCoder) ToFen(data *bitfield.BitField) string {
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

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(bits, pieceCount, occupancy, c.frequencies, colors, state)

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
