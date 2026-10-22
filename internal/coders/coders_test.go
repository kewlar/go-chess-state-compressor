package coders_test

import (
	"fmt"
	"testing"

	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/coders"
)

func TestCodersRoundTrip(t *testing.T) {
	positions := []struct {
		name string
		fen  string
	}{
		{
			name: "Starting position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "Longest FEN",
			fen:  "b1q1k1r1/r1n1n1b1/1p1p1p1p/p1p1p1p1/P1P1P1P1/1P1P1P1P/R1N1N1B1/B1Q1K1R1 w - - 18 18",
		},
		{
			name: "Earliest possible promotion (to Bishop)",
			fen:  "rnbqkbBr/pppppp1p/5n2/8/8/8/PPPPP1PP/RNBQKBNR b KQkq - 0 5",
		},
		{
			name: "Earliest possible promotion (to Rook)",
			fen:  "rnbqkbRr/pppppp1p/5n2/8/8/8/PPPPP1PP/RNBQKBNR b KQkq - 0 5",
		},
		{
			name: "18 Queens",
			fen:  "3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80",
		},
		{
			name: "18 Queens, no pieces on home ranks",
			fen:  "3qQq1q/1rkb1Q1q/Q6q/7q/Q7/Q5K1/Q1q3B1/QqQq1QR1 b - - 0 80",
		},
		{
			name: "Ending for the longest possible game",
			fen:  "8/8/8/5K2/6Qk/8/8/8 b - - 150 8849",
		},
		{
			name: "Black cannot be in check",
			fen:  "4k3/5R2/6Q1/8/1q6/8/3R4/4K3 w - - 0 99",
		},
		{
			name: "White cannot be in check",
			fen:  "4k3/5R2/6Q1/8/1q6/8/3R4/4K3 b - - 0 99",
		},
		{
			name: "Last move while the pieces on the home ranks still are only in their original colors.",
			fen:  "rnbqkbnr/p1p1pppp/Bp6/3p4/4P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 3",
		},
		{
			name: "Fastest capture by White on the back rank",
			fen:  "rnBqkbnr/p1p1pppp/1p6/3p4/4P3/8/PPPP1PPP/RNBQK1NR b KQkq - 0 3",
		},
		{
			name: "Kuwait Immortal - 6 Pawns in a file.",
			fen:  "k7/P7/P7/P7/P7/P7/P7/R3K3 w Q - 1 35",
		},
		{
			name: "Shortest bare kings endgame",
			fen:  "8/4k3/8/8/8/8/5K2/8 b - - 0 17",
		},
		{
			name: "Typical puzzle without movecount",
			fen:  "4R3/8/8/2Pkp3/N7/4rnKB/1nb5/b1r5 w - - 0 1",
		},
		{
			name: "Typical puzzle without movecount (1)",
			fen:  "4R3/8/8/2Pkp3/N5K1/4rn1B/1nb5/b1r5 b - - 1 1",
		},
		{
			name: "Typical puzzle without movecount (2)",
			fen:  "4R3/8/8/2Pkp3/N5K1/4r2B/1nb4n/b1r5 w - - 2 2",
		},
	}
	coders := []coders.Coder{
		coders.NewFenCoder(),
		coders.NewBinaryFenCoder(),
		coders.NewBinaryFenCompactCoder(),
		coders.NewHuffmanCoder(),
		coders.NewHuffmanCompactCoder(),
		coders.NewArithmeticCoder(),
		coders.NewArithmeticByCountCoder(),
		coders.NewArithmeticAdaptiveCoder(),
		coders.NewArithmeticAdaptiveCompactCoder(),
		coders.NewHuffmanAdaptiveCompactCoder(),
		coders.NewArithmeticAdaptiveCompact2Coder(),
	}

	for _, coder := range coders {
		for _, position := range positions {
			t.Run(fmt.Sprintf("%s: %s", coder.Name(), position.name), func(t *testing.T) {
				state, err := chess.ParseFen(position.fen)
				if err != nil {
					panic("error in test data")
				}
				encoded := coder.Encode(state)
				decoded := coder.Decode(encoded)
				decodedFen := decoded.ToFen()
				if position.fen != decodedFen {
					t.Errorf("Decode(Encode(%q)) = %q; expected %q", position.fen, decodedFen, position.fen)
				}
			})
		}
	}
}
