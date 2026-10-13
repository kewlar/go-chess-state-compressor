package chess

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type State struct {
	Board      Board
	SideToMove Color // 'w' or 'b'
	Castling   Castling
	EnPassant  Square // 0..63, or InvalidSquare
	Halfmove   int    // 0..150
	Fullmove   int    // 1..8849
}

const (
	MaxHalfmove = 150
	MaxFullmove = 8849
)

var (
	reCastlingOptions = regexp.MustCompile(`^(-|[KQkq]{1,4})$`)
	reEnPassantSquare = regexp.MustCompile(`^(-|[a-h][36])$`)
)

func NewState() *State {
	return &State{
		Castling:  Castling{},
		EnPassant: InvalidSquare,
	}
}

func NewStartingState() *State {
	state := NewState()
	homeRow := []Piece{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for i := range 8 {
		piece := homeRow[i]
		state.Board[i] = NewPiece(rune(piece), Black)
		state.Board[i+8] = NewPiece(rune(Pawn), Black)
		state.Board[i+48] = NewPiece(rune(Pawn), White)
		state.Board[i+56] = NewPiece(rune(piece), White)
	}
	state.SideToMove = White
	state.Castling['K'] = true
	state.Castling['Q'] = true
	state.Castling['k'] = true
	state.Castling['q'] = true
	state.Halfmove = 0
	state.Fullmove = 1
	return state
}

func ParseFen(fen string) (*State, error) {
	parts := strings.Fields(fen)
	state := NewState()

	// Piece placement
	countWhite, countBlack := 0, 0
	row, col := 0, 0
	for _, r := range parts[0] {
		if r == '/' {
			if col != 8 {
				return nil, fmt.Errorf("expected row to be 8 squares in size, got %d in %s", col, fen)
			}
			row++
			col = 0
			continue
		}

		if '1' <= r && r <= '8' {
			col += int(r - '0')
			continue
		}

		piece := Piece(r)
		if !piece.IsValid() {
			return nil, fmt.Errorf("unexpected piece %q in fen %s", r, fen)
		}

		if piece.Color() == White {
			countWhite++
		} else {
			countBlack++
		}

		state.Board[NewSquareFromRankAndFile(7-row, col)] = piece

		col++
	}
	if col != 8 {
		return nil, fmt.Errorf("expected row to be 8 squares in size, got %d in %s", col, fen)
	}

	// Should have decoded 64 squares.
	cells := row*8 + col
	if cells != 64 {
		return nil, fmt.Errorf("expected to decode 64 squares in %s", fen)
	}

	countPieces := countWhite + countBlack
	if countPieces < 2 || countPieces > 32 {
		return nil, fmt.Errorf("expected number of pieces to be between 2 and 32, got %d", countPieces)
	}
	if countWhite < 1 || countWhite > 16 {
		return nil, fmt.Errorf("expected number of white pieces to be between 1 and 16, got %d", countWhite)
	}
	if countBlack < 1 || countBlack > 16 {
		return nil, fmt.Errorf("expected number of black pieces to be between 1 and 16, got %d", countBlack)
	}

	// Side to move.
	if parts[1] != "w" && parts[1] != "b" {
		return nil, fmt.Errorf("side to move can be b or w, got: %s", parts[1])
	}
	state.SideToMove = NewColor(rune(parts[1][0]))

	// Castling availability.
	if !reCastlingOptions.MatchString(parts[2]) {
		return nil, fmt.Errorf("expected valid castling options, got: %s", parts[2])
	}
	for _, option := range castlingOptions {
		state.Castling[option] = strings.ContainsRune(parts[2], option)
	}

	// En passant square.
	if !reEnPassantSquare.MatchString(parts[3]) {
		return nil, fmt.Errorf("en passant square must be - or a valid en passant square, got: %s", parts[3])
	}
	if parts[3] != "-" {
		state.EnPassant = NewSquare(parts[3])
	} else {
		state.EnPassant = InvalidSquare
	}

	// Halfmove clock.
	halfmove, err := strconv.Atoi(parts[4])
	if err != nil || halfmove < 0 || halfmove > MaxHalfmove {
		return nil, fmt.Errorf("error parsing halfmove clock in '%s': %v", fen, err)
	}
	state.Halfmove = halfmove

	// Fullmove number.
	fullmove, err := strconv.Atoi(parts[5])
	if err != nil || fullmove < 1 || fullmove > MaxFullmove {
		return nil, fmt.Errorf("error parsing fullmove number in '%s': %v", fen, err)
	}
	state.Fullmove = fullmove

	return state, nil
}

func (state *State) ToFen() string {
	fen := strings.Builder{}

	// Piece placement.
	empty := 0
	idx := 0
	for i, piece := range state.Board {
		if i&7 == 0 && i != 0 {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			fen.WriteRune('/')
		}
		if piece != 0 {
			if empty > 0 {
				fen.WriteRune('0' + rune(empty))
				empty = 0
			}
			fen.WriteRune(rune(piece))
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
	fen.WriteRune(state.SideToMove.Rune())
	fen.WriteRune(' ')

	// Castling availability.
	fen.WriteString(state.Castling.String())
	fen.WriteRune(' ')

	// En passant square.
	fen.WriteString(Square(state.EnPassant).String())
	fen.WriteRune(' ')

	// Halfmove clock.
	fen.WriteString(strconv.Itoa(state.Halfmove))
	fen.WriteRune(' ')

	// Fullmove number.
	fen.WriteString(strconv.Itoa(state.Fullmove))

	return fen.String()
}
