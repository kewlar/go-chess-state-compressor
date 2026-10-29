package squareskipper_test

import (
	"slices"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/bitfield"
	"github.com/kewlar/go-chess-state-compressor/internal/chess"
	"github.com/kewlar/go-chess-state-compressor/internal/tools/squareskipper"
)

func TestSkipCastling(t *testing.T) {
	tests := []struct {
		name           string
		initial        uint64
		castling       string
		mode           squareskipper.Mode
		expected       uint64
		expectedLength int
	}{
		{
			name:           "No castling (occupancy)",
			initial:        0b11111111_11111111_00000000_00000000_00000000_00000000_11111111_11111111,
			castling:       "",
			mode:           squareskipper.Occupancy,
			expected:       0b11111111_11111111_00000000_00000000_00000000_00000000_11111111_11111111,
			expectedLength: 64,
		},
		{
			name:           "Full castling (occupancy)",
			initial:        0b11011011_11111111_00000000_00000000_00000000_00000000_11111111_10111101,
			castling:       "KQkq",
			mode:           squareskipper.Occupancy,
			expected:       0b10101_11111111_00000000_00000000_00000000_00000000_11111111_01110,
			expectedLength: 58,
		},
		{
			name:           "Partial castling (occupancy)",
			initial:        0b11011011_11111111_00000000_00000000_00000000_00000000_11111111_10111101,
			castling:       "Kq",
			mode:           squareskipper.Occupancy,
			expected:       0b101011_11111111_00000000_00000000_00000000_00000000_11111111_101110,
			expectedLength: 60,
		},
		{
			name:           "White castling kingside (occupancy)",
			initial:        0b11011011_11111111_00000000_00000000_00000000_00000000_11111111_10111101,
			castling:       "K",
			mode:           squareskipper.Occupancy,
			expected:       0b101011_11111111_00000000_00000000_00000000_00000000_11111111_10111101,
			expectedLength: 62,
		},
		{
			name:           "No castling (color)",
			initial:        0b11111111_11111111_00000000_00000000_00000000_00000000_00000000_00000000,
			castling:       "",
			mode:           squareskipper.Color,
			expected:       0b11111111_11111111_00000000_00000000_00000000_00000000_00000000_00000000,
			expectedLength: 64,
		},
		{
			name:           "Full castling (color)",
			initial:        0b10111101_11111111_00000000_00000000_00000000_00000000_00000000_01000010,
			castling:       "KQkq",
			mode:           squareskipper.Color,
			expected:       0b01110_11111111_00000000_00000000_00000000_00000000_00000000_10001,
			expectedLength: 58,
		},
		{
			name:           "Partial castling (color)",
			initial:        0b10111101_11111111_00000000_00000000_00000000_00000000_00000000_01000010,
			castling:       "Kq",
			mode:           squareskipper.Color,
			expected:       0b011101_11111111_00000000_00000000_00000000_00000000_00000000_010001,
			expectedLength: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Deflate", func(t *testing.T) {
				castling := chess.NewCastlingFromString(tt.castling)
				initialBoard := bitfield.IntToBoolSlice(tt.initial, 64)
				expectedBoard := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipCastling(castling, tt.mode)
				result := skipper.Deflate(initialBoard)
				if !slices.Equal(result, expectedBoard) {
					t.Errorf("SquareSkipper.SkipCastling(%q).Deflate(%#v) = %#v; expected %#v", tt.castling, initialBoard, result, expectedBoard)
				}
			})
			t.Run("Inflate", func(t *testing.T) {
				castling := chess.NewCastlingFromString(tt.castling)
				initialBoard := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				expectedBoard := bitfield.IntToBoolSlice(tt.initial, 64)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipCastling(castling, tt.mode)
				result := skipper.Inflate(initialBoard)
				if !slices.Equal(result, expectedBoard) {
					t.Errorf("SquareSkipper.SkipCastling(%q).Inflate(%#v) = %#v; expected %#v", tt.castling, initialBoard, result, expectedBoard)
				}
			})
		})
	}
}

func TestSkipEnPassant(t *testing.T) {
	tests := []struct {
		name           string
		initial        uint64
		enPassant      chess.Square
		mode           squareskipper.Mode
		expected       uint64
		expectedLength int
	}{
		{
			name:           "e4 (occupancy)",
			initial:        0b11111111_11101111_00000000_00010000_00000000_00000000_11111111_11111111,
			enPassant:      chess.E3,
			mode:           squareskipper.Occupancy,
			expected:       0b11111111_1111111_0000000_0000000_00000000_00000000_11111111_11111111,
			expectedLength: 61,
		},
		{
			name:           "e5 (occupancy)",
			initial:        0b11111111_11111111_00000000_00000000_00010000_00000000_11101111_11111111,
			enPassant:      chess.E6,
			mode:           squareskipper.Occupancy,
			expected:       0b11111111_11111111_00000000_00000000_0000000_0000000_1111111_11111111,
			expectedLength: 61,
		},
		{
			name:           "e4 (color)",
			initial:        0b11111111_11101111_00000000_00010000_00000000_00000000_00000000_00000000,
			enPassant:      chess.E3,
			mode:           squareskipper.Color,
			expected:       0b11111111_1111111_0000000_0000000_00000000_00000000_00000000_00000000,
			expectedLength: 61,
		},
		{
			name:           "e5 (color)",
			initial:        0b11111111_11111111_00000000_00000000_00000000_00000000_00000000_00000000,
			enPassant:      chess.E6,
			mode:           squareskipper.Color,
			expected:       0b11111111_11111111_00000000_00000000_0000000_0000000_0000000_00000000,
			expectedLength: 61,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Deflate", func(t *testing.T) {
				initialBoard := bitfield.IntToBoolSlice(tt.initial, 64)
				expectedBoard := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipEnPassant(tt.enPassant, tt.mode)
				result := skipper.Deflate(initialBoard)
				if !slices.Equal(result, expectedBoard) {
					t.Errorf("SquareSkipper.SkipEnPassant(%q).Deflate(%#v) = %#v; expected %#v", tt.enPassant.String(), initialBoard, result, expectedBoard)
				}
			})
			t.Run("Inflate", func(t *testing.T) {
				initialBoard := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				expectedBoard := bitfield.IntToBoolSlice(tt.initial, 64)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipEnPassant(tt.enPassant, tt.mode)
				result := skipper.Inflate(initialBoard)
				if !slices.Equal(result, expectedBoard) {
					t.Errorf("SquareSkipper.SkipEnPassant(%q).Inflate(%#v) = %#v; expected %#v", tt.enPassant.String(), initialBoard, result, expectedBoard)
				}
			})
		})
	}
}

func TestSkipOriginalRanks(t *testing.T) {
	tests := []struct {
		name           string
		color          uint64
		occupancy      uint64
		expected       uint64
		expectedLength int
	}{
		{
			name:           "e4",
			color:          0b11111111_11101111_00000000_00010000_00000000_00000000_00000000_00000000,
			occupancy:      0b11111111_11101111_00000000_00010000_00000000_00000000_11111111_11111111,
			expected:       0b0_00000000_00010000_00000000_00000000,
			expectedLength: 33,
		},
		{
			name:           "Last move while the pieces on the home ranks still are only in their original colors",
			color:          0b11011111_11101111_00000000_00010000_00000000_00000001_00000000_00000000,
			occupancy:      0b11011111_11101111_00000000_00010000_00001000_00000011_11110101_11111111,
			expected:       0b0_0_00000000_00010000_00000000_00000001_00,
			expectedLength: 36,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Deflate", func(t *testing.T) {
				color := bitfield.IntToBoolSlice(tt.color, 64)
				occupancy := bitfield.IntToBoolSlice(tt.occupancy, 64)
				expectedColor := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipOriginalRanks(occupancy)
				result := skipper.Deflate(color)
				if !slices.Equal(result, expectedColor) {
					t.Errorf("SquareSkipper.SkipOriginalRanks().Deflate(%#v) = %#v; expected %#v", color, result, expectedColor)
				}
			})
			t.Run("Inflate", func(t *testing.T) {
				color := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				occupancy := bitfield.IntToBoolSlice(tt.occupancy, 64)
				expectedColor := bitfield.IntToBoolSlice(tt.color, 64)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipOriginalRanks(occupancy)
				result := skipper.Inflate(color)
				if !slices.Equal(result, expectedColor) {
					t.Errorf("SquareSkipper.SkipOriginalRanks().Inflate(%#v) = %#v; expected %#v", color, result, expectedColor)
				}
			})
		})
	}
}

func TestSkipEmpty(t *testing.T) {
	tests := []struct {
		name           string
		occupancy      uint64
		color          uint64
		expected       uint64
		expectedLength int
	}{
		{
			name:           "Starting board",
			occupancy:      0b11111111_11111111_00000000_00000000_00000000_00000000_11111111_11111111,
			color:          0b11111111_11111111_00000000_00000000_00000000_00000000_00000000_00000000,
			expected:       0b11111111_11111111_00000000_00000000,
			expectedLength: 32,
		},
		{
			name:           "Arbitrary position",
			occupancy:      0b00000000_11011111_01100001_00000001_10000110_10110001_10111110_00000000,
			color:          0b00000000_11001011_01100001_00000001_00000000_00010001_00100000_00000000,
			expected:       0b11101011_111_1_000_0011_010000,
			expectedLength: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Deflate", func(t *testing.T) {
				occupancy := bitfield.IntToBoolSlice(tt.occupancy, 64)
				color := bitfield.IntToBoolSlice(tt.color, 64)
				expectedColor := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipEmpty(occupancy)
				result := skipper.Deflate(color)
				if !slices.Equal(result, expectedColor) {
					t.Errorf("SquareSkipper.SkipEmpty(%#v).Deflate(%#v) = %#v; expected %#v", occupancy, color, result, expectedColor)
				}
			})
			t.Run("Inflate", func(t *testing.T) {
				occupancy := bitfield.IntToBoolSlice(tt.occupancy, 64)
				color := bitfield.IntToBoolSlice(tt.expected, tt.expectedLength)
				expectedColor := bitfield.IntToBoolSlice(tt.color, 64)
				skipper := squareskipper.NewSquareSkipper()
				skipper.SkipEmpty(occupancy)
				result := skipper.Inflate(color)
				if !slices.Equal(result, expectedColor) {
					t.Errorf("SquareSkipper.SkipEmpty(%#v).Inflate(%#v) = %#v; expected %#v", occupancy, color, result, expectedColor)
				}
			})
		})
	}
}
