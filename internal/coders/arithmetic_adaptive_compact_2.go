package coders

import (
	"log"
	"slices"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
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
	bf := &bitfield.BitField{}

	// Encode starting position.
	if state.Fullmove == 1 && state.SideToMove == chess.White {
		bf.WriteInt(0, 21) // moveNum
		return bf
	}

	// Encode the board after first move by White
	if state.Fullmove == 1 && state.SideToMove == chess.Black {
		bf.WriteInt(1, 21) // moveNum
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		bf.WriteInt(uint64(whiteMove), 5)
		return bf
	}

	// Encode the board after first move by Black
	if state.Fullmove == 2 && state.SideToMove == chess.White {
		bf.WriteInt(2, 21) // moveNum
		whiteMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		blackMove := entropy.FindMatchingMoveIndex(state.Board, chess.FirstBlackMoves())
		if blackMove == -1 { // Not found.
			panic("move not found")
		}

		bf.WriteInt(uint64(whiteMove*20+blackMove), 9)
		return bf
	}

	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	originalColorRanks := state.Fullmove < 3 || state.Fullmove == 3 && state.SideToMove == chess.White

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
			skip[chess.NewSquare("e1")] = true
			skip[chess.NewSquare("h1")] = true
		}
		if state.Castling['Q'] {
			skip[chess.NewSquare("e1")] = true
			skip[chess.NewSquare("a1")] = true
		}
		if state.Castling['k'] {
			skip[chess.NewSquare("e8")] = true
			skip[chess.NewSquare("h8")] = true
		}
		if state.Castling['q'] {
			skip[chess.NewSquare("e8")] = true
			skip[chess.NewSquare("a8")] = true
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
	skip = map[chess.Square]bool{}
	for cell, isOccupied := range occupancy {
		if !isOccupied {
			skip[chess.Square(cell)] = true
		}
	}
	if originalColorRanks {
		for file := range 8 {
			skip[chess.NewSquareFromRankAndFile(0, file)] = true
			skip[chess.NewSquareFromRankAndFile(1, file)] = true
			skip[chess.NewSquareFromRankAndFile(6, file)] = true
			skip[chess.NewSquareFromRankAndFile(7, file)] = true
		}
	}
	skippedSquares = []int{}
	for s := range skip {
		skippedSquares = append(skippedSquares, int(s))
	}
	colorOut := tools.WithoutSquares(color, skippedSquares)
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
	state := chess.NewState()

	// Halfmove clock & fullMove number.
	moveNum := int(data.ReadInt(21))
	switch moveNum {
	case 0:
		state = chess.NewStartingState()
	case 1:
		num := int(data.ReadInt(5))
		state = chess.NewStartingState()
		move := chess.FirstWhiteMoves()[num]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		}
		state.SideToMove = chess.Black
	case 2:
		num := int(data.ReadInt(9))
		whiteMove, blackMove := num/20, num%20

		state = chess.NewStartingState()
		// Apply move by White.
		move := chess.FirstWhiteMoves()[whiteMove]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		}
		// Apply move by Black.
		move = chess.FirstBlackMoves()[blackMove]
		state.Board[move.To] = state.Board[move.From]
		state.Board[move.From] = chess.Empty
		if state.Board[move.To].Type() == chess.Pawn && tools.Abs(move.From.Rank()-move.To.Rank()) == 2 {
			state.EnPassant = chess.NewSquareFromRankAndFile((move.From.Rank()+move.To.Rank())/2, move.To.File())
		} else {
			state.EnPassant = chess.InvalidSquare
		}
		if state.Board[move.To].Type() != chess.Pawn {
			state.Halfmove++
		} else {
			state.Halfmove = 0
		}

		state.Fullmove = 2
		state.SideToMove = chess.White
	default:
		state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

		// Side to move.
		state.SideToMove = chess.Color(data.ReadBool())

		skip := map[chess.Square]bool{} // Squares that are skipped in occupancy data.

		// Castling availability.
		hasCastling := data.ReadBool()
		if hasCastling {
			state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
			if state.Castling['K'] {
				skip[chess.NewSquare("e1")] = true
				skip[chess.NewSquare("h1")] = true
			}
			if state.Castling['Q'] {
				skip[chess.NewSquare("e1")] = true
				skip[chess.NewSquare("a1")] = true
			}
			if state.Castling['k'] {
				skip[chess.NewSquare("e8")] = true
				skip[chess.NewSquare("h8")] = true
			}
			if state.Castling['q'] {
				skip[chess.NewSquare("e8")] = true
				skip[chess.NewSquare("a8")] = true
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
		pieceCount := 0
		for _, occupied := range occupancy {
			if occupied {
				pieceCount++
			}
		}

		// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
		originalColorRanks := state.Fullmove < 3 || state.Fullmove == 3 && state.SideToMove == chess.White

		skip = map[chess.Square]bool{} // Squares that are skipped in color data.
		for cell, isOccupied := range occupancy {
			if !isOccupied {
				skip[chess.Square(cell)] = false
			}
		}
		if originalColorRanks {
			for file := range 8 {
				for _, rank := range []int{0, 1, 6, 7} {
					sq := chess.NewSquareFromRankAndFile(rank, file)
					skip[sq] = occupancy[sq] && rank <= 1
				}
			}
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
		idx := 0
		for i, occupied := range occupancy {
			if occupied {
				state.Board[i] = chess.NewPiece(pieces[idx], chess.Color(colors[idx]))
				idx++
			}
		}
	}

	return state
}
