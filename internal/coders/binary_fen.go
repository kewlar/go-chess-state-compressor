package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type BinaryFenCoder struct {
	pieceCodes [256]byte
	codePieces []rune
}

func NewBinaryFenCoder() *BinaryFenCoder {
	codePieces := []rune{'p', 'k', 'q', 'r', 'b', 'n'}

	pieceCodes := [256]byte{}
	for code, piece := range codePieces {
		pieceCodes[byte(piece)] = byte(code)
	}

	return &BinaryFenCoder{
		pieceCodes: pieceCodes,
		codePieces: codePieces,
	}
}

func (c *BinaryFenCoder) Name() string {
	return "BinaryFen"
}

func (c *BinaryFenCoder) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bits.WriteInt(uint64(c.pieceCodes[piece.Type()]), 3)
	}

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	return bits
}

func (c *BinaryFenCoder) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		piece := byte(bits.ReadInt(3))
		pieces = append(pieces, c.codePieces[piece])
	}

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
