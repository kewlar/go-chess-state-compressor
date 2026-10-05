package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
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

func (c *ArithmeticCoder) FromFen(fen string) *bitfield.BitField {
	state, _ := chess.ParseFen(fen)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy := uint64(0)
	color := uint32(0)
	pieces := []rune{}
	for cell, piece := range state.Board {
		if piece.IsEmpty() {
			continue
		}

		occupancy |= 1 << cell
		if piece.Color() == chess.White {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, rune(piece.Type()))
	}

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))

	// Arithmetic coding.
	bits := arithmetic.Encode(pieces, c.intervals)

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

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

func (c *ArithmeticCoder) ToFen(data *bitfield.BitField) string {
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
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.Decode(bits, pieceCount, c.intervals)

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
