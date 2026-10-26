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
	castlingSquares []chess.Square
}

func NewBinaryFenCompactCoder() *BinaryFenCompactCoder {
	var (
		codePieces      = []rune{'p', 'k', 'q', 'r', 'b', 'n', 'P', 'R'}
		castlingOptions = []rune{'K', 'Q', 'k', 'q'}
		castlingSquares = []chess.Square{chess.H1, chess.A1, chess.H8, chess.A8} // Cell coordinates for castling rooks, respectively KQkq.
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
	return "BinaryFen (compact)"
}

func (c *BinaryFenCompactCoder) Encode(state *chess.State) *bitfield.BitField {
	// En passant square.
	enPassantPawn := chess.InvalidSquare
	if state.EnPassant.IsValid() {
		enPassantPawn = state.EnPassant ^ 8
	}

	// Piece placement
	occupancy, color, _pieces := separateBoardInformation(state.Board)
	pieces := piecesToRunes(_pieces)
	idx := 0
	for square, piece := range state.Board.OccupiedSquares() {
		pieceType := piece.Type()
		if enPassantPawn == square && pieceType == 'p' {
			pieceType = 'P' // En passant pawn.
		} else if pieceType == 'r' {
			pieceType = 'R' // Rook not available for castling by default.
			// Castling availability.
			for i, option := range c.castlingOptions {
				if square == c.castlingSquares[i] && state.Castling[option] {
					pieceType = 'r'
				}
			}
		}
		pieces[idx] = rune(pieceType)
		idx++
	}

	bits := &bitfield.BitField{}

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bits.WriteInt(uint64(c.pieceCodes[piece]), 3)
	}

	// Side to move.
	bits.WriteBool(bool(state.SideToMove))

	// Castling availability.
	// Encoded via special rook code.

	// En passant square.
	// Encoded via special pawn code.

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bits.WriteInt(uint64(moveNum), 21)

	return bits
}

func (c *BinaryFenCompactCoder) Decode(bits *bitfield.BitField) *chess.State {
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

	// Side to move.
	state.SideToMove = chess.Color(bits.ReadBool())

	// Halfmove clock & fullMove number.
	moveNum := int(bits.ReadInt(21))
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// En passant and castling.
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			switch pieces[idx] {
			case 'P': // En passant pawn
				state.EnPassant = chess.Square(i ^ 8)
				pieces[idx] = 'p'
			case 'R': // Tainted rook
				pieces[idx] = 'r'
			case 'r': // Castling-available rook
				for j, cell := range c.castlingSquares {
					if chess.Square(i) == cell {
						state.Castling[c.castlingOptions[j]] = true
					}
				}
			}
			idx++
		}
	}
	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
