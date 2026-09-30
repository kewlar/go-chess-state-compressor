package coders

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"example.com/go-chess/internal/bitfield"
)

type FenValidator struct {
	reCastlingOptions *regexp.Regexp
	reEnPassantSquare *regexp.Regexp
}

func NewFenValidator() *FenValidator {
	return &FenValidator{
		reCastlingOptions: regexp.MustCompile(`^(-|[KQkq]{1,4})$`),
		reEnPassantSquare: regexp.MustCompile(`^(-|[a-h][36])$`),
	}
}

func (c *FenValidator) Name() string {
	return "FenValidator"
}

func (c *FenValidator) FromFen(fen string) *bitfield.BitField {
	parts := strings.Fields(fen)

	// Piece placement
	countWhite, countBlack := 0, 0
	row, col := 0, 0
	for _, r := range parts[0] {
		if r == '/' {
			if col != 8 {
				log.Fatalf("expected row to be 8 squares in size, got %d in %s", col, fen)
			}
			row++
			col = 0
			continue
		}

		if '1' <= r && r <= '8' {
			col += int(r - '0')
			continue
		}

		pieceType := rune(int(r) | 32)
		if !strings.ContainsRune("prkqbn", pieceType) {
			log.Fatalf("unexpected piece %q in fen %s", r, fen)
		}

		isWhite := int(r)&32 == 0
		if isWhite {
			countWhite++
		} else {
			countBlack++
		}

		col++
	}
	if col != 8 {
		log.Fatalf("expected row to be 8 squares in size, got %d in %s", col, fen)
	}

	// Sanity check: should have decoded 64 squares.
	cells := row*8 + col
	if cells != 64 {
		log.Fatalf("expected to decode 64 squares in %s", fen)
	}

	countPieces := countWhite + countBlack
	if countPieces < 2 || countPieces > 32 {
		log.Fatalf("expected number of pieces to be between 2 and 32, got %d", countPieces)
	}
	if countWhite < 1 || countWhite > 16 {
		log.Fatalf("expected number of white pieces to be between 1 and 16, got %d", countWhite)
	}
	if countBlack < 1 || countBlack > 16 {
		log.Fatalf("expected number of black pieces to be between 1 and 16, got %d", countBlack)
	}

	// Side to move.
	if parts[1] != "w" && parts[1] != "b" {
		log.Fatalf("side to move can be b or w, got: %s", parts[1])
	}

	// Castling availability.
	if !c.reCastlingOptions.MatchString(parts[2]) {
		log.Fatalf("expected valid castling options, got: %s", parts[2])
	}

	// En passant square.
	if !c.reEnPassantSquare.MatchString(parts[3]) {
		log.Fatalf("en passant square must be - or a valid en passant square, got: %s", parts[3])
	}

	// Halfmove clock.
	halfmove, err := strconv.Atoi(parts[4])
	if err != nil || halfmove < 0 || halfmove > 150 {
		log.Fatalf("error parsing halfmove clock in '%s': %v", fen, err)
	}

	// Fullmove number.
	fullmove, err := strconv.Atoi(parts[5])
	if err != nil || fullmove < 1 || fullmove > 8848 {
		log.Fatalf("error parsing fullmove number in '%s': %v", fen, err)
	}

	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *FenValidator) ToFen(data *bitfield.BitField) string {
	// Piece placement: occupancy.
	l := data.Length/8 - 1
	return string(data.Data[:l])
}
