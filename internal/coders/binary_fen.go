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
	bf := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteInt(uint64(c.pieceCodes[piece.Type()]), 3)
	}

	bf.WriteBool(bool(state.SideToMove))
	bf.WriteBoolSlice(state.Castling.BoolSlice())
	bf.WriteInt(uint64(state.EnPassant), 7)
	bf.WriteInt(uint64(state.Halfmove), 8)
	bf.WriteInt(uint64(state.Fullmove), 14)

	return bf
}

func (c *BinaryFenCoder) Decode(data *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		piece := byte(data.ReadInt(3))
		pieces = append(pieces, c.codePieces[piece])
	}

	state.SideToMove = chess.Color(data.ReadBool())
	state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
	state.EnPassant = chess.Square(data.ReadInt(7))
	state.Halfmove = int(data.ReadInt(8))
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
