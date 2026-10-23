package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type BinaryFenCompactCoder struct {
	pieceCodes      [256]byte
	codePieces      []rune
	castlingOptions []rune
	castlingSquares []int
}

func NewBinaryFenCompactCoder() *BinaryFenCompactCoder {
	var (
		codePieces      = []rune{'p', 'k', 'q', 'r', 'b', 'n', 'P', 'R'}
		castlingOptions = []rune{'K', 'Q', 'k', 'q'}
		castlingSquares = []int{63, 56, 7, 0} // Cell coordinates for castling rooks, respectively KQkq.
	)
	pieceCodes := [256]byte{}
	for code, piece := range codePieces {
		pieceCodes[byte(piece)] = byte(code)
	}

	return &BinaryFenCompactCoder{
		pieceCodes:      pieceCodes,
		codePieces:      codePieces,
		castlingOptions: castlingOptions,
		castlingSquares: castlingSquares,
	}
}

func (c *BinaryFenCompactCoder) Name() string {
	return "BinaryFenCompact"
}

func (c *BinaryFenCompactCoder) Encode(state *chess.State) *bitfield.BitField {
	// En passant square.
	enPassantPawn := chess.InvalidSquare
	if state.EnPassant.IsValid() {
		enPassantPawn = state.EnPassant ^ 8
	}

	// Piece placement
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
		pieceType := piece.Type()

		if enPassantPawn == chess.Square(cell) && pieceType == 'p' {
			pieceType = 'P' // En passant pawn.
		} else if pieceType == 'r' {
			pieceType = 'R' // Rook not available for castling by default.
			// Castling availability.
			for i, option := range c.castlingOptions {
				if cell == c.castlingSquares[i] && state.Castling[option] {
					pieceType = 'r'
				}
			}
		}
		pieces = append(pieces, rune(pieceType))
	}

	// GENERATE BITS

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteInt(uint64(c.pieceCodes[piece]), 3)
	}

	// Side to move.
	bf.WriteBool(bool(state.SideToMove))

	// Castling availability.
	// Encoded via special rook code.

	// En passant square.
	// Encoded via special pawn code.

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bf.WriteInt(uint64(moveNum), 21)

	return bf
}

func (c *BinaryFenCompactCoder) Decode(data *bitfield.BitField) *chess.State {
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

	// En passant square.
	state.EnPassant = 64 // To be parsed from piece codes.

	// Halfmove clock & fullMove number.
	moveNum := int(data.ReadInt(21))
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Piece placement.
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			piece, color := pieces[idx], colors[idx]
			if piece == 'P' { // En passant pawn
				state.EnPassant = chess.Square(i ^ 8)
				piece = 'p'
			} else if piece == 'R' { // Tainted rook
				piece = 'r'
			} else if piece == 'r' { // Castling-available rook
				for j, cell := range c.castlingSquares {
					if i == cell {
						state.Castling[c.castlingOptions[j]] = true
					}
				}
			}
			state.Board[i] = chess.NewPiece(piece, chess.Color(color))
			idx++
		}
	}

	return state
}
