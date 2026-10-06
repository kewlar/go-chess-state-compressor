package coders

import (
	"fmt"
	"log"
	"maps"
	"slices"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
	"example.com/go-chess/internal/tools"
)

type HuffmanAdaptiveCompactCoder struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCompactCoder() *HuffmanAdaptiveCompactCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	return &HuffmanAdaptiveCompactCoder{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCompactCoder) Name() string {
	return "Huffman (adaptive, compact)"
}

func (c *HuffmanAdaptiveCompactCoder) FromFen(fen string) *bitfield.BitField {
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
	color := uint32(0)
	fullColor := color
	pieces := []rune{}
	colorCount := 0
	for c, piece := range state.Board {
		cell := chess.Square(c)
		if piece.IsEmpty() {
			continue
		}

		logColor := !originalColorRanks || 1 < cell.Rank() && cell.Rank() < 6

		occupancy[cell] = true
		if piece.Color() == chess.White {
			fullColor |= 1 << len(pieces)
			if logColor {
				color |= 1 << colorCount
			}
		}
		pieces = append(pieces, rune(piece.Type()))
		if logColor {
			colorCount++
		}
	}
	fullOccupancy := slices.Clone(occupancy)

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
	occupancyOut := uint64(0)
	for i, occupied := range occupancy {
		if occupied {
			occupancyOut |= 1 << i
		}
	}
	if len(pieces) < 32 && state.SideToMove == chess.Black {
		occupancyOut = occupancyOut ^ 0xFF_FF_FF_FF_FF_FF_FF_FF
	}

	// Piece placement.
	bf.WriteInt(occupancyOut, len(occupancy))

	// Side to move.
	if len(pieces) == 32 {
		bf.WriteBool(bool(state.SideToMove))
	}

	// Piece colors.
	bf.WriteInt(uint64(color), colorCount)

	// Huffman coding.
	// TODO: implement
	freqs := maps.Clone(c.frequencies)
	interval := arithmetic.NewInterval(arithmetic.NewRational(0, 1), arithmetic.NewRational(1, 1))
	n := len(pieces)
	colorSlice := bitfield.IntToBoolSlice(uint64(fullColor), len(pieces))
	idx := 0
	for i := range 64 {
		if occupancy[i] {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(i), n, fullOccupancy, colorSlice, state, pieces[:idx])
			intervals := arithmetic.FrequenciesToIntervals(freqs2)
			r := pieces[idx]
			if freqs2[r] == 0 {
				panic(fmt.Sprintf("didn't expect '%s' at square %s", string(r), chess.Square(i).String()))
			}
			interval.Mul(intervals[r])
			idx++

			freqs = entropy.UpdateFrequencies(r, n-idx, freqs)
		}
	}

	bits := interval.ToBits()

	if len(bits) > (1<<bitsForArithmeticCodeLength)-1 {
		log.Fatalf("encoded data too long (%d bits)", len(bits))
	}
	bf.WriteInt(uint64(len(bits)), bitsForArithmeticCodeLength)
	bf.WriteBoolSlice(bits)

	return bf
}

func (c *HuffmanAdaptiveCompactCoder) ToFen(data *bitfield.BitField) string {
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

		// Side to move.
		if pieceCount == 32 {
			state.SideToMove = chess.Color(data.ReadBool())
		}

		// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
		originalColorRanks := state.Fullmove < 3 || state.Fullmove == 3 && state.SideToMove == chess.White

		// Piece placement: colors.
		colorCount := pieceCount
		numBlack, numWhite := 0, 0
		if originalColorRanks {
			colorCount = 0
			for i, x := range occupancy {
				c := chess.Square(i)
				if !x {
					continue
				}
				switch c.Rank() {
				case 6, 7:
					numBlack++
				case 0, 1:
					numWhite++
				default:
					colorCount++
				}
			}
		}
		colors := data.ReadBoolSlice(colorCount)
		if originalColorRanks {
			blacks := []bool{}
			for range numBlack {
				blacks = append(blacks, false)
			}
			whites := []bool{}
			for range numWhite {
				whites = append(whites, true)
			}
			colors = append(blacks, colors...)
			colors = append(colors, whites...)
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
		// TODO: implement
		freqs := maps.Clone(c.frequencies)
		n := pieceCount
		data := make([]rune, 0, n)
		x := arithmetic.NewRationalFromBits(bits)
		idx := 0
		for square, occupied := range occupancy {
			if occupied {
				freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(square), n, occupancy, colors, state, data)
				intervals := arithmetic.FrequenciesToIntervals(freqs2)
				piece := rune(0)
				for r, i := range intervals {
					if i.Cover(x) {
						piece = r
						break
					}
				}
				if piece == 0 {
					panic("covering interval not found")
				}
				data = append(data, piece)
				idx++
				x.ScaleWithInterval(intervals[piece])

				freqs = entropy.UpdateFrequencies(piece, n-idx, freqs)
			}
		}

		pieces := data

		// Piece placement.
		idx = 0
		for i, occupied := range occupancy {
			if occupied {
				state.Board[i] = chess.NewPiece(pieces[idx], chess.Color(colors[idx]))
				idx++
			}
		}
	}

	return state.ToFen()
}
