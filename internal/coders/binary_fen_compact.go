package coders

import (
	"strconv"
	"strings"

	"example.com/go-chess/internal/bitfield"
)

type BinaryFenCompactCoder struct {
	pieceCodes      [256]byte
	codePieces      []rune
	castlingOptions []rune
}

func NewBinaryFenCompactCoder() *BinaryFenCompactCoder {
	var (
		codePieces      = []rune{'p', 'k', 'q', 'r', 'b', 'n', 'P', 'R'}
		castlingOptions = []rune{'K', 'Q', 'k', 'q'}
	)
	pieceCodes := [256]byte{}
	for code, piece := range codePieces {
		pieceCodes[byte(piece)] = byte(code)
	}

	return &BinaryFenCompactCoder{
		pieceCodes:      pieceCodes,
		codePieces:      codePieces,
		castlingOptions: castlingOptions,
	}
}

func (c *BinaryFenCompactCoder) Name() string {
	return "BinaryFenCompact"
}

func (c *BinaryFenCompactCoder) FromFen(fen string) *bitfield.BitField {
	// PARSE FEN

	parts := strings.Fields(fen)

	// Side to move.
	whiteToMove := parts[1] == "w"

	// Castling availability.
	castling := map[rune]bool{}
	for _, option := range c.castlingOptions {
		castling[option] = strings.ContainsRune(parts[2], option)
	}

	// En passant square.
	enPassant, enPassantPawn := byte(64), 64 // 64 == no en passant
	if parts[3] != "-" {
		file := parts[3][0] - 'a'
		rank := parts[3][1] - '1'
		enPassant = (7-rank)*8 + file
		enPassantPawn = int(enPassant) ^ 8 // The coordinate of the pawn that moved two squares.
	}

	// Halfmove clock.
	halfMoves, _ := strconv.Atoi(parts[4])

	// Fullmove number.
	fullMoves, _ := strconv.Atoi(parts[5])

	// Piece placement
	occupancy := uint64(0)
	color := uint32(0)
	pieces := []byte{}
	cell := 0
	for _, r := range parts[0] {
		if r == '/' {
			continue
		}

		if '1' <= r && r <= '8' {
			cell += int(r - '0')
			continue
		}

		pieceType := rune(int(r) | 32)

		occupancy |= 1 << cell

		isWhite := int(r)&32 == 0
		if isWhite {
			color |= 1 << len(pieces)
		}

		if enPassantPawn == cell && pieceType == 'p' {
			pieceType = 'P' // En passant pawn.
		} else if pieceType == 'r' {
			pieceType = 'R'
			if cell == 0 && castling['q'] || cell == 7 && castling['k'] || cell == 56 && castling['Q'] || cell == 63 && castling['K'] {
				pieceType = 'r'
			}
		}
		pieces = append(pieces, c.pieceCodes[pieceType])

		cell++
	}

	// GENERATE BITS

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteInt(uint64(piece), 3)
	}

	// Side to move.
	bf.WriteBool(whiteToMove)

	// Castling availability.
	// Encoded via special rook code.

	// En passant square.
	// Encoded via special pawn code.

	// Halfmove clock & fullmove number.
	moveNum := fullMoves*151 + halfMoves
	bf.WriteInt(uint64(moveNum), 21)

	return bf
}

func (c *BinaryFenCompactCoder) ToFen(data *bitfield.BitField) string {
	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := 0
	for _, occupied := range occupancy {
		if occupied {
			pieceCount++
		}
	}

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		piece := byte(data.ReadInt(3))
		pieces = append(pieces, c.codePieces[piece])
	}

	// Side to move.
	isWhite := data.ReadBool()

	// Castling availability.
	castling := map[rune]bool{}

	// En passant square.
	enPassant := 64 // To be parsed from piece codes.

	// Halfmove clock & fullMove number.
	moveNum := data.ReadInt(21)
	fullMoves, halfMoves := moveNum/151, moveNum%151

	// GENERATE FEN

	fen := strings.Builder{}

	// Piece placement.
	empty := 0
	idx := 0
	for i, occupied := range occupancy {
		if i&7 == 0 && i != 0 {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			fen.WriteRune('/')
		}
		if occupied {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			piece, color := pieces[idx], colors[idx]
			if piece == 'P' { // En passant pawn
				enPassant = i ^ 8
				piece = 'p'
			} else if piece == 'R' { // Tainted rook
				piece = 'r'
			} else if piece == 'r' { // Castling-available rook
				switch i {
				case 0:
					castling['q'] = true
				case 7:
					castling['k'] = true
				case 56:
					castling['Q'] = true
				case 63:
					castling['K'] = true
				}
			}
			if color { // White
				piece &= 0b1101_1111
			}
			fen.WriteRune(piece)
			idx++
		} else {
			empty++
		}
	}
	if empty > 0 {
		fen.WriteRune('0' + rune(empty))
	}
	fen.WriteRune(' ')

	// Side to move.
	if isWhite {
		fen.WriteRune('w')
	} else {
		fen.WriteRune('b')
	}
	fen.WriteRune(' ')

	// Castling availability.
	none := true
	for _, option := range c.castlingOptions {
		if castling[option] {
			fen.WriteRune(option)
			none = false
		}
	}
	if none {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// En passant square.
	if enPassant <= 63 {
		rank := 7 - (enPassant >> 3)
		file := enPassant & 7
		fen.WriteRune('a' + rune(file))
		fen.WriteRune('1' + rune(rank))
	} else {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// Halfmove clock.
	fen.WriteString(strconv.Itoa(int(halfMoves)))
	fen.WriteRune(' ')

	// Fullmove number.
	fen.WriteString(strconv.Itoa(int(fullMoves)))

	return fen.String()
}
