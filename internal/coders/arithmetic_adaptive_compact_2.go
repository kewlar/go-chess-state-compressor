package coders

import (
	"log"
	"slices"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticAdaptiveCompact2Coder struct {
	frequencies map[rune]int
}

func NewArithmeticAdaptiveCompact2Coder() *ArithmeticAdaptiveCompact2Coder {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &ArithmeticAdaptiveCompact2Coder{
		frequencies: frequencies,
	}
}

func (c *ArithmeticAdaptiveCompact2Coder) Name() string {
	return "Arithmetic (adaptive, compact, v2)"
}

func (c *ArithmeticAdaptiveCompact2Coder) Encode(state *chess.State) *bitfield.BitField {
	if bf, ok := TryEncodeEarlyPosition(state); ok {
		return bf
	}

	bf := &bitfield.BitField{}

	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := maybeOriginalColorRanks && reallyOriginalColorRanks(state.Board)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy := make([]bool, 64)
	color := make([]bool, 64)
	fullColor := []bool{}
	pieces := []rune{}
	for c, piece := range state.Board {
		cell := chess.Square(c)
		if piece.IsEmpty() {
			continue
		}

		occupancy[cell] = true
		color[cell] = piece.Color() == chess.White
		pieces = append(pieces, rune(piece.Type()))
		fullColor = append(fullColor, piece.Color() == chess.White)
	}
	fullOccupancy := slices.Clone(occupancy)

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bf.WriteInt(uint64(moveNum), 21)

	skip := map[chess.Square]bool{} // Map of squares to skip in occupancy data.

	// Side to move.
	bf.WriteBool(bool(state.SideToMove))

	// Castling availability.
	bf.WriteBool(state.Castling.IsAvailable())
	if state.Castling.IsAvailable() {
		bf.WriteBoolSlice(state.Castling.BoolSlice())
		if state.Castling['K'] {
			skip[chess.E1] = true
			skip[chess.H1] = true
		}
		if state.Castling['Q'] {
			skip[chess.E1] = true
			skip[chess.A1] = true
		}
		if state.Castling['k'] {
			skip[chess.E8] = true
			skip[chess.H8] = true
		}
		if state.Castling['q'] {
			skip[chess.E8] = true
			skip[chess.A8] = true
		}
	}

	// En passant square.
	if state.EnPassant.IsValid() {
		bf.WriteBool(true)
		bf.WriteInt(uint64(state.EnPassant), 3) // Only the lowest 3 bits -- the file -- are necessary.
		enPassantPawn := chess.Square(int(state.EnPassant) ^ 0b1000)
		skip[enPassantPawn] = true
		skip[state.EnPassant] = false
		enPassantHome := chess.Square(int(state.EnPassant) ^ 0b11000)
		skip[enPassantHome] = false
	} else {
		bf.WriteBool(false)
	}

	// Piece placement.
	skippedSquares := []int{}
	for s := range skip {
		skippedSquares = append(skippedSquares, int(s))
	}
	occupancyOut := tools.WithoutSquares(occupancy, skippedSquares)
	bf.WriteBoolSlice(occupancyOut)

	// Piece colors.
	for cell, isOccupied := range occupancy {
		if !isOccupied {
			skip[chess.Square(cell)] = true
		}
	}
	if originalColorRanks {
		for file := range 8 {
			for _, rank := range []int{0, 1, 6, 7} {
				skip[chess.NewSquareFromRankAndFile(rank, file)] = true
			}
		}
	}
	skippedSquares = []int{}
	for s := range skip {
		skippedSquares = append(skippedSquares, int(s))
	}
	colorOut := tools.WithoutSquares(color, skippedSquares)
	if maybeOriginalColorRanks {
		bf.WriteBool(originalColorRanks)
	}
	bf.WriteBoolSlice(colorOut)

	// Arithmetic coding.
	bits := arithmetic.EncodeAdaptive(pieces, fullOccupancy, c.frequencies, fullColor, state)

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

	return bf
}

func (c *ArithmeticAdaptiveCompact2Coder) Decode(data *bitfield.BitField) *chess.State {
	// Halfmove clock & fullMove number.
	moveNum := int(data.ReadInt(21))

	if state, ok := TryDecodeEarlyPosition(moveNum, data); ok {
		return state
	}

	state := chess.NewState()
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Side to move.
	state.SideToMove = chess.Color(data.ReadBool())

	skip := map[chess.Square]bool{} // Squares that are skipped in occupancy data.

	// Castling availability.
	hasCastling := data.ReadBool()
	if hasCastling {
		state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
		if state.Castling['K'] {
			skip[chess.E1] = true
			skip[chess.H1] = true
		}
		if state.Castling['Q'] {
			skip[chess.E1] = true
			skip[chess.A1] = true
		}
		if state.Castling['k'] {
			skip[chess.E8] = true
			skip[chess.H8] = true
		}
		if state.Castling['q'] {
			skip[chess.E8] = true
			skip[chess.A8] = true
		}
	}
	// En passant square.
	hasEnPassant := data.ReadBool()
	if hasEnPassant {
		enPassantFile := int(data.ReadInt(3))
		rank := 2
		if state.SideToMove == chess.White {
			rank = 5
		}
		state.EnPassant = chess.NewSquareFromRankAndFile(rank, enPassantFile)
		enPassantPawn := chess.Square(int(state.EnPassant) ^ 0b1000)
		skip[enPassantPawn] = true
		skip[state.EnPassant] = false
		enPassantHome := chess.Square(int(state.EnPassant) ^ 0b11000)
		skip[enPassantHome] = false
	}

	// Piece placement: occupancy.
	skippedSquares := map[int]bool{}
	for s, x := range skip {
		skippedSquares[int(s)] = x
	}
	occupancy := data.ReadBoolSlice(64 - len(skippedSquares))
	occupancy = tools.WithSquares(occupancy, skippedSquares)
	pieceCount := tools.CountTrue(occupancy)

	// Restore full-board piece color map. True = White, False = Black or empty.
	for cell, isOccupied := range occupancy {
		if !isOccupied {
			skip[chess.Square(cell)] = false // Empty squares to be added.
		}
	}
	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := false
	if maybeOriginalColorRanks {
		originalColorRanks = data.ReadBool()
	}
	if originalColorRanks {
		for file := range 8 {
			for _, rank := range []int{0, 1, 6, 7} {
				sq := chess.NewSquareFromRankAndFile(rank, file)
				skip[sq] = occupancy[sq] && rank <= 1 // White or Black.
			}
		}
	}
	if state.Castling['k'] {
		skip[chess.E8] = false // Black squares to be added.
		skip[chess.H8] = false
	}
	if state.Castling['q'] {
		skip[chess.E8] = false
		skip[chess.A8] = false
	}
	if state.EnPassant.IsValid() {
		enPassantPawn := chess.Square(int(state.EnPassant) ^ 0b1000)
		skip[enPassantPawn] = state.SideToMove == chess.Black
	}
	skippedSquares = map[int]bool{}
	for s, x := range skip {
		skippedSquares[int(s)] = x
	}

	// Piece placement: colors.
	colorIn := data.ReadBoolSlice(64 - len(skippedSquares))
	fullColor := tools.WithSquares(colorIn, skippedSquares)
	colors := []bool{}
	for i, isOccupied := range occupancy {
		if isOccupied {
			colors = append(colors, fullColor[i])
		}
	}

	// Piece placement: pieces.
	l := int(data.ReadInt(bitsForArithmeticCodeLength))
	bits := data.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(bits, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
