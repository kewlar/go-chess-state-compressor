package coders

import (
	"fmt"
	"slices"

	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanAdaptiveCompactCoder struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCompactCoder() *HuffmanAdaptiveCompactCoder {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &HuffmanAdaptiveCompactCoder{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCompactCoder) Name() string {
	return "Huffman (adaptive, compact)"
}

func (c *HuffmanAdaptiveCompactCoder) Encode(state *chess.State) *bitfield.BitField {
	if bits, ok := TryEncodeEarlyPosition(state); ok {
		return bits
	}

	bits := &bitfield.BitField{}

	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := maybeOriginalColorRanks && reallyOriginalColorRanks(state.Board)

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
	bits.WriteInt(uint64(moveNum), 21)

	// Castling availability.
	bits.WriteBool(state.Castling.IsAvailable())
	if state.Castling.IsAvailable() {
		bits.WriteBoolSlice(state.Castling.BoolSlice())
	}

	// En passant square.
	if state.EnPassant.IsValid() {
		bits.WriteBool(true)
		bits.WriteInt(uint64(state.EnPassant), 3) // Only the lowest 3 bits -- the file -- are necessary.
	} else if state.Halfmove == 0 { // Only write "no en passant" flag if en passant is at all possible. In this case, the last move might have been a Pawn move.
		bits.WriteBool(false)
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
	bits.WriteInt(occupancyOut, len(occupancy))

	// Side to move.
	if len(pieces) == 32 {
		bits.WriteBool(bool(state.SideToMove))
	}

	// Piece colors.
	if maybeOriginalColorRanks {
		bits.WriteBool(originalColorRanks)
	}
	bits.WriteInt(uint64(color), colorCount)

	// Huffman coding.
	n := len(pieces)
	freqs := entropy.PrepareFrequencies(n, c.frequencies)
	colorSlice := bitfield.IntToBoolSlice(uint64(fullColor), len(pieces))
	idx := 0
	for i := range 64 {
		if occupancy[i] {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(i), n, fullOccupancy, colorSlice, state, pieces[:idx])
			r := pieces[idx]
			if freqs2[r] == 0 {
				panic(fmt.Sprintf("didn't expect '%s' at square %s", string(r), chess.Square(i).String()))
			}
			tree := huffman.BuildHuffmanTree(freqs2)
			dict := tree.BuildDictionary()
			bits.WriteBoolSlice(dict[r].Bits)
			idx++

			freqs = entropy.UpdateFrequencies(r, n-idx, freqs)
		}
	}

	return bits
}

func (c *HuffmanAdaptiveCompactCoder) Decode(bits *bitfield.BitField) *chess.State {
	// Halfmove clock & fullMove number.
	moveNum := int(bits.ReadInt(21))

	if state, ok := TryDecodeEarlyPosition(moveNum, bits); ok {
		return state
	}

	state := chess.NewState()
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Castling availability.
	hasCastling := bits.ReadBool()
	if hasCastling {
		state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	}
	// En passant square.
	hasEnPassant := false
	enPassantFile := -1
	if state.Halfmove == 0 { // The last move might have been a Pawn move.
		hasEnPassant = bits.ReadBool()
		if hasEnPassant {
			enPassantFile = int(bits.ReadInt(3))
		}
	}

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

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
		state.SideToMove = chess.Color(bits.ReadBool())
	}

	// Up to possibility for Black to make their 3rd move, all pieces on the home ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := false
	if maybeOriginalColorRanks {
		originalColorRanks = bits.ReadBool()
	}

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
	colors := bits.ReadBoolSlice(colorCount)
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
	freqs := entropy.PrepareFrequencies(pieceCount, c.frequencies)
	pieces := make([]rune, 0, pieceCount)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(square), pieceCount, occupancy, colors, state, pieces)
			tree := huffman.BuildHuffmanTree(freqs2)

			node := tree
			for !node.IsLeaf() {
				bit := bits.ReadBool()
				node = node.Children[bit]
			}
			piece := node.Value
			pieces = append(pieces, piece)

			idx++

			freqs = entropy.UpdateFrequencies(piece, pieceCount-idx, freqs)
		}
	}

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
