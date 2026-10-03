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

	// Arithmetic coding.
	interval := arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1))
	for _, piece := range pieces {
		interval.Mul(c.intervals[piece])
	}
	bits := interval.ToBits()
	if len(bits) > 255 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), 8)
	bf.WriteBoolSlice(bits)

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
	pieces := make([]rune, 0, pieceCount)
	l := int(data.ReadInt(8))
	bits := data.ReadBoolSlice(l)
	x := arithmetic.NewRationalFromBits(bits)
	for range pieceCount {
		piece := rune(0)
		for r, i := range c.intervals {
			if i.Cover(x) {
				piece = r
				break
			}
		}
		if piece == 0 {
			log.Fatalf("covering interval not found")
		}
		pieces = append(pieces, piece)
		x.ScaleWithInterval(c.intervals[piece])
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
