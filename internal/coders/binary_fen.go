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
	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy := uint64(0)
	color := uint32(0)
	pieces := []chess.Piece{}
	for cell, piece := range state.Board {
		if piece.IsEmpty() {
			continue
		}

		occupancy |= 1 << cell
		if piece.Color() == chess.White {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, piece)
	}

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteInt(uint64(c.pieceCodes[piece.Type()]), 3)
	}

	// Side to move.
	bf.WriteBool(bool(state.SideToMove))

	// Castling availability.
	bf.WriteBoolSlice(state.Castling.BoolSlice())

	// En passant square.
	bf.WriteInt(uint64(state.EnPassant), 7)

	// Halfmove clock.
	bf.WriteInt(uint64(state.Halfmove), 8)

	// Fullmove number.
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

	// Side to move.
	state.SideToMove = chess.Color(data.ReadBool())

	// Castling availability.
	state.Castling = chess.NewCastling(data.ReadBoolSlice(4))

	// En passant square.
	state.EnPassant = chess.Square(data.ReadInt(7))

	// Halfmove clock.
	state.Halfmove = int(data.ReadInt(8))

	// Fullmove number.
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement.
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			state.Board[i] = chess.NewPiece(pieces[idx], chess.Color(colors[idx]))
			idx++
		}
	}

	return state
}
