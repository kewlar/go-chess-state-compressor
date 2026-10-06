package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
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

func findMatchingMoveIndex(board chess.Board, moves []chess.Move) int {
	for i, m := range moves {
		if board[m.From].IsEmpty() && board[m.To].IsOccupied() {
			return i
		}
	}
	return -1
}

func (c *ArithmeticAdaptiveCompactCoder) FromFen(fen string) *bitfield.BitField {
	state, _ := chess.ParseFen(fen)

	bf := &bitfield.BitField{}

	// Encode starting position.
	if state.Fullmove == 1 && state.SideToMove == chess.White {
		bf.WriteInt(0, 21) // moveNum
		return bf
	}

	// Encode the board after first move by White
	if state.Fullmove == 1 && state.SideToMove == chess.Black {
		bf.WriteInt(1, 21) // moveNum
		whiteMove := findMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		bf.WriteInt(uint64(whiteMove), 5)
		return bf
	}

	// Encode the board after first move by Black
	if state.Fullmove == 2 && state.SideToMove == chess.White {
		bf.WriteInt(2, 21) // moveNum
		whiteMove := findMatchingMoveIndex(state.Board, chess.FirstWhiteMoves())
		if whiteMove == -1 { // Not found.
			panic("move not found")
		}
		blackMove := findMatchingMoveIndex(state.Board, chess.FirstBlackMoves())
		if blackMove == -1 { // Not found.
			panic("move not found")
		}

		bf.WriteInt(uint64(whiteMove*20+blackMove), 9)
		return bf
	}

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
	}

	return state.ToFen()
}
