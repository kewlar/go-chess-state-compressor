package codecs

import (
	"github.com/kewlar/go-chess-state-compressor/internal/bitfield"
	"github.com/kewlar/go-chess-state-compressor/internal/chess"
	"github.com/kewlar/go-chess-state-compressor/internal/huffman"
	"github.com/kewlar/go-chess-state-compressor/internal/tools"
	"github.com/kewlar/go-chess-state-compressor/internal/tools/squareskipper"
)

type HuffmanAdaptiveCompact2Codec struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCompact2Codec() *HuffmanAdaptiveCompact2Codec {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &HuffmanAdaptiveCompact2Codec{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCompact2Codec) Name() string {
	return "Huffman (adaptive, compact, v2)"
}

func (c *HuffmanAdaptiveCompact2Codec) Encode(state *chess.State) *bitfield.BitField {
	if bits, ok := TryEncodeEarlyPosition(state); ok {
		return bits
	}

	bits := &bitfield.BitField{}

	// Up to possibility for Black to make their 3rd move, all pieces on the starting ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
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

	skipper := squareskipper.NewSquareSkipper()

	// Side to move.
	bits.WriteBool(bool(state.SideToMove))

	// Castling availability.
	bits.WriteBool(state.Castling.IsAvailable())
	if state.Castling.IsAvailable() {
		bits.WriteBoolSlice(state.Castling.BoolSlice())
		skipper.SkipCastling(state.Castling, squareskipper.Occupancy)
	}

	// En passant square.
	if state.EnPassant.IsValid() {
		bits.WriteBool(true)
		bits.WriteInt(uint64(state.EnPassant), 3) // Only the lowest 3 bits -- the file -- are necessary.
		skipper.SkipEnPassant(state.EnPassant, squareskipper.Occupancy)
	} else if state.Halfmove == 0 { // Only write "no en passant" flag if en passant is at all possible. In this case, the last move might have been a Pawn move.
		bits.WriteBool(false)
	}

	// Piece placement.
	occupancyOut := skipper.Deflate(fullOccupancy)
	bits.WriteBoolSlice(occupancyOut)

	// Piece colors.
	skipper.SkipEmpty(fullOccupancy)
	if originalColorRanks {
		skipper.SkipOriginalRanks(fullOccupancy)
	}
	colorOut := skipper.Deflate(color)
	if maybeOriginalColorRanks {
		bits.WriteBool(originalColorRanks)
	}
	bits.WriteBoolSlice(colorOut)

	// Huffman coding.
	colorSlice := bitfield.IntToBoolSlice(uint64(fullColor), len(pieces))
	bits.WriteBoolSlice(huffman.EncodeAdaptive(piecesToRunes(pieces), fullOccupancy, c.frequencies, colorSlice, state))

	return bits
}

func (c *HuffmanAdaptiveCompact2Codec) Decode(bits *bitfield.BitField) *chess.State {
	// Halfmove clock & fullmove number.
	moveNum := int(bits.ReadInt(21))

	if state, ok := TryDecodeEarlyPosition(moveNum, bits); ok {
		return state
	}

	state := chess.NewState()
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Side to move.
	state.SideToMove = chess.Color(bits.ReadBool())

	skipper := squareskipper.NewSquareSkipper()

	// Castling availability.
	hasCastling := bits.ReadBool()
	if hasCastling {
		state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
		skipper.SkipCastling(state.Castling, squareskipper.Occupancy)
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
			skipper.SkipEnPassant(state.EnPassant, squareskipper.Occupancy)
		}
	}

	// Piece placement: occupancy.
	occupancyIn := bits.ReadBoolSlice(64 - skipper.Length())
	occupancy := skipper.Inflate(occupancyIn)
	pieceCount := tools.CountTrue(occupancy)

	// Restore full-board piece color map. True = White, False = Black or empty.
	skipper.SkipEmpty(occupancy)
	// Up to possibility for Black to make their 3rd move, all pieces on the starting ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := false
	if maybeOriginalColorRanks {
		originalColorRanks = bits.ReadBool()
	}
	if originalColorRanks {
		skipper.SkipOriginalRanks(occupancy)
	}
	skipper.SkipCastling(state.Castling, squareskipper.Color)
	if state.EnPassant.IsValid() {
		skipper.SkipEnPassant(state.EnPassant, squareskipper.Color)
	}

	// Piece placement: colors.
	colorIn := bits.ReadBoolSlice(64 - skipper.Length())
	fullColor := skipper.Inflate(colorIn)
	colors := []bool{}
	for i, isOccupied := range occupancy {
		if isOccupied {
			colors = append(colors, fullColor[i])
		}
	}

	// Piece placement: pieces.
	pieces := huffman.DecodeAdaptive(func() bool { return bits.ReadBool() }, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
