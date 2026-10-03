package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type ArithmeticAdaptiveCoder struct {
	frequencies      map[rune]int
	intervals        map[rune]*arithmetic.Interval
	homeRowIntervals map[rune]*arithmetic.Interval
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
	homeRowFrequencies := map[rune]int{
		'p': 0,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	return &ArithmeticAdaptiveCoder{
		frequencies:      frequencies,
		intervals:        arithmetic.FrequenciesToIntervals(frequencies),
		homeRowIntervals: arithmetic.FrequenciesToIntervals(homeRowFrequencies),
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

	// Arithmetic coding.
	bits := arithmetic.Encode(pieces, c.intervals)

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
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

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.Decode(bits, pieceCount, c.intervals)

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
