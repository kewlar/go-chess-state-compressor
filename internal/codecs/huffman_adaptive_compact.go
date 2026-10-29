package codecs

import (
	"slices"

	"github.com/kewlar/go-chess-state-compressor/internal/bitfield"
	"github.com/kewlar/go-chess-state-compressor/internal/chess"
	"github.com/kewlar/go-chess-state-compressor/internal/huffman"
	"github.com/kewlar/go-chess-state-compressor/internal/tools"
)

type HuffmanAdaptiveCompactCodec struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCompactCodec() *HuffmanAdaptiveCompactCodec {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &HuffmanAdaptiveCompactCodec{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCompactCodec) Name() string {
	return "Huffman (adaptive, compact)"
}

func (c *HuffmanAdaptiveCompactCodec) Encode(state *chess.State) *bitfield.BitField {
	if bits, ok := TryEncodeEarlyPosition(state); ok {
		return bits
	}

	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, fullColor, pieces := separateBoardInformation(state.Board)

	// Up to possibility for Black to make their 3rd move, all pieces on the starting ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := maybeOriginalColorRanks && reallyOriginalColorRanks(state.Board)

	// Don't store color for home ranks, if we can help it.
	color := fullColor
	colorCount := len(pieces)
	if originalColorRanks {
		color = 0
		colorCount = 0
		for square, piece := range state.Board.OccupiedSquares() {
			if 1 < square.Rank() && square.Rank() < 6 {
				if piece.Color() == chess.White {
					color |= 1 << colorCount
				}
				colorCount++
			}
		}
	}

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
	occupancyOut := occupancy
	if len(pieces) < 32 && state.SideToMove == chess.Black {
		occupancyOut = occupancyOut ^ 0xFF_FF_FF_FF_FF_FF_FF_FF
	}

	// Piece placement.
	bits.WriteInt(occupancyOut, 64)

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
	fullOccupancy := bitfield.IntToBoolSlice(occupancy, 64)
	colorSlice := bitfield.IntToBoolSlice(uint64(fullColor), len(pieces))
	bits.WriteBoolSlice(huffman.EncodeAdaptive(piecesToRunes(pieces), fullOccupancy, c.frequencies, colorSlice, state))

	return bits
}

func (c *HuffmanAdaptiveCompactCodec) Decode(bits *bitfield.BitField) *chess.State {
	// Halfmove clock & fullmove number.
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

	// Side to move. May be encoded in the occupancy bitboard.
	state.SideToMove = chess.White
	if pieceCount > 32 {
		for i, x := range occupancy {
			occupancy[i] = !x
		}
		pieceCount = 64 - pieceCount
		state.SideToMove = chess.Black
	} else if pieceCount == 32 {
		state.SideToMove = chess.Color(bits.ReadBool())
	}

	// Up to possibility for Black to make their 3rd move, all pieces on the starting ranks (1, 2, 7, 8) have their original colors, and we don't need to encode that.
	maybeOriginalColorRanks := maybeOriginalColorRanks(state)
	originalColorRanks := false
	if maybeOriginalColorRanks {
		originalColorRanks = bits.ReadBool()
	}

	// Piece placement: colors.
	colorCount := pieceCount
	numBlack, numWhite := 0, 0
	if originalColorRanks {
		for i, occupied := range occupancy {
			if !occupied {
				continue
			}
			switch chess.Square(i).Rank() {
			case 6, 7:
				numBlack++
			case 0, 1:
				numWhite++
			}
		}
		colorCount -= numBlack + numWhite
	}
	colors := bits.ReadBoolSlice(colorCount)
	if originalColorRanks {
		blacks := make([]bool, numBlack)
		whites := slices.Repeat([]bool{true}, numWhite)
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
	pieces := huffman.DecodeAdaptive(func() bool { return bits.ReadBool() }, pieceCount, occupancy, c.frequencies, colors, state)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
