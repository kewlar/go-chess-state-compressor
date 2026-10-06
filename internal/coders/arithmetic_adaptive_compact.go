package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type ArithmeticAdaptiveCompactCoder struct {
	frequencies map[rune]int
}

func NewArithmeticAdaptiveCompactCoder() *ArithmeticAdaptiveCompactCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	return &ArithmeticAdaptiveCompactCoder{
		frequencies: frequencies,
	}
}

func (c *ArithmeticAdaptiveCompactCoder) Name() string {
	return "Arithmetic (adaptive, compact)"
}

func (c *ArithmeticAdaptiveCompactCoder) FromFen(fen string) *bitfield.BitField {
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

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bf.WriteInt(uint64(moveNum), 21)

	// Castling availability.
	hasCastling := false
	for _, option := range chess.CastlingOptions {
		hasCastling = hasCastling || state.Castling[option]
	}
	bf.WriteBool(hasCastling)
	if hasCastling {
		for _, option := range chess.CastlingOptions {
			bf.WriteBool(state.Castling[option])
		}
	}

	// En passant square.
	if state.EnPassant.IsValid() {
		bf.WriteBool(true)
		bf.WriteInt(uint64(state.EnPassant), 3) // Only the lowest 3 bits -- the file -- are necessary.
	} else {
		bf.WriteBool(false)
	}

	// Encode side to move in occupancy data, if possible.
	occupancyOut := occupancy
	if len(pieces) < 32 && state.SideToMove == chess.Black {
		occupancyOut = occupancy ^ 0xFF_FF_FF_FF_FF_FF_FF_FF
	}

	// Piece placement.
	bf.WriteInt(occupancyOut, 64)
	bf.WriteInt(uint64(color), len(pieces))

	// Side to move.
	if len(pieces) == 32 {
		bf.WriteBool(bool(state.SideToMove))
	}

	// Arithmetic coding.
	bits := arithmetic.EncodeAdaptive(pieces, bitfield.IntToBoolSlice(occupancy, 64), c.frequencies, bitfield.IntToBoolSlice(uint64(color), len(pieces)), state)

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

	return bf
}

func (c *ArithmeticAdaptiveCompactCoder) ToFen(data *bitfield.BitField) string {
	state := chess.NewState()

	// Halfmove clock & fullMove number.
	moveNum := int(data.ReadInt(21))
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Castling availability.
	hasCastling := data.ReadBool()
	if hasCastling {
		for _, option := range chess.CastlingOptions {
			state.Castling[option] = data.ReadBool()
		}
	}
	// En passant square.
	hasEnPassant := data.ReadBool()
	enPassantFile := -1
	if hasEnPassant {
		enPassantFile = int(data.ReadInt(3))
	}

	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := 0
	for _, occupied := range occupancy {
		if occupied {
			pieceCount++
		}
	}

	state.SideToMove = chess.White
	if pieceCount > 32 {
		for i, x := range occupancy {
			occupancy[i] = !x
		}
		pieceCount = 64 - pieceCount
		state.SideToMove = chess.Black
	}

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Side to move.
	if pieceCount == 32 {
		state.SideToMove = chess.Color(data.ReadBool())
	}

	// Finish decoding en passant square.
	if hasEnPassant {
		rank := 2
		if state.SideToMove == chess.White {
			rank = 5
		}
		state.EnPassant = chess.NewSquareFromRankAndFile(rank, enPassantFile)
	}

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(bits, pieceCount, occupancy, c.frequencies, colors, state)

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
