package coders

import (
	"log"

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
	if bits, ok := TryEncodeEarlyPosition(state); ok {
		return bits
	}

	bits := &bitfield.BitField{}

	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := maybeOriginalColorRanks && reallyOriginalColorRanks(state.Board)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, fullColor, pieces := separateBoardInformation(state.Board)
	color := make([]bool, 64)
	for square, piece := range state.Board.OccupiedSquares() {
		color[square] = piece.Color() == chess.White
	}
	fullOccupancy := bitfield.IntToBoolSlice(occupancy, 64)

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bits.WriteInt(uint64(moveNum), 21)

	skip := map[chess.Square]bool{} // Map of squares to skip in occupancy data.

	// Side to move.
	bits.WriteBool(bool(state.SideToMove))

	// Castling availability.
	bits.WriteBool(state.Castling.IsAvailable())
	if state.Castling.IsAvailable() {
		bits.WriteBoolSlice(state.Castling.BoolSlice())
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
		bits.WriteBool(true)
		bits.WriteInt(uint64(state.EnPassant), 3) // Only the lowest 3 bits -- the file -- are necessary.
		enPassantPawn := chess.Square(int(state.EnPassant) ^ 0b1000)
		skip[enPassantPawn] = true
		skip[state.EnPassant] = false
		enPassantHome := chess.Square(int(state.EnPassant) ^ 0b11000)
		skip[enPassantHome] = false
	} else if state.Halfmove == 0 { // Only write "no en passant" flag if en passant is at all possible. In this case, the last move might have been a Pawn move.
		bits.WriteBool(false)
	}

	// Piece placement.
	skippedSquares := []int{}
	for s := range skip {
		skippedSquares = append(skippedSquares, int(s))
	}
	occupancyOut := tools.WithoutSquares(fullOccupancy, skippedSquares)
	bits.WriteBoolSlice(occupancyOut)

	// Piece colors.
	for cell, isOccupied := range fullOccupancy {
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
		bits.WriteBool(originalColorRanks)
	}
	bits.WriteBoolSlice(colorOut)

	// Arithmetic coding.
	colorSlice := bitfield.IntToBoolSlice(uint64(fullColor), len(pieces))
	code := arithmetic.EncodeAdaptive(piecesToRunes(pieces), fullOccupancy, c.frequencies, colorSlice, state)
	if len(code) > maxBitsForArithmeticCode {
		log.Fatalf("encoded data too long (%d bits)", len(code))
	}
	bits.WriteInt(uint64(len(code)), bitsForArithmeticCodeLength)
	bits.WriteBoolSlice(code)

	return bits
}

func (c *ArithmeticAdaptiveCompact2Coder) Decode(bits *bitfield.BitField) *chess.State {
	// Halfmove clock & fullMove number.
	moveNum := int(bits.ReadInt(21))

	if state, ok := TryDecodeEarlyPosition(moveNum, bits); ok {
		return state
	}

	state := chess.NewState()
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Side to move.
	state.SideToMove = chess.Color(bits.ReadBool())

	skip := map[chess.Square]bool{} // Squares that are skipped in occupancy data.

	// Castling availability.
	hasCastling := bits.ReadBool()
	if hasCastling {
		state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
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
	hasEnPassant := false
	if state.Halfmove == 0 { // The last move might have been a Pawn move.
		hasEnPassant = bits.ReadBool()
		if hasEnPassant {
			enPassantFile := int(bits.ReadInt(3))
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
	}

	// Piece placement: occupancy.
	skippedSquares := map[int]bool{}
	for s, x := range skip {
		skippedSquares[int(s)] = x
	}
	occupancy := bits.ReadBoolSlice(64 - len(skippedSquares))
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
		originalColorRanks = bits.ReadBool()
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
	colorIn := bits.ReadBoolSlice(64 - len(skippedSquares))
	fullColor := tools.WithSquares(colorIn, skippedSquares)
	colors := []bool{}
	for i, isOccupied := range occupancy {
		if isOccupied {
			colors = append(colors, fullColor[i])
		}
	}

	// Piece placement: pieces.
	l := int(bits.ReadInt(bitsForArithmeticCodeLength))
	code := bits.ReadBoolSlice(l)
	pieces := arithmetic.DecodeAdaptive(code, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
