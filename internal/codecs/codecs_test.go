package codecs_test

import (
	"fmt"
	"testing"

	"github.com/kewlar/go-chess-state-compressor/internal/chess"
	"github.com/kewlar/go-chess-state-compressor/internal/codecs"
)

type testPosition struct {
	name string
	fen  string
}

func testCodecs() []codecs.Codec {
	return []codecs.Codec{
		codecs.NewFenCodec(),
		codecs.NewBinaryFenCodec(),
		codecs.NewBinaryFenCompactCodec(),
		codecs.NewHuffmanCodec(),
		codecs.NewHuffmanAdaptiveCodec(),
		codecs.NewHuffmanCompactCodec(),
		codecs.NewArithmeticCodec(),
		codecs.NewArithmeticByCountCodec(),
		codecs.NewArithmeticAdaptiveCodec(),
		codecs.NewArithmeticAdaptiveCompactCodec(),
		codecs.NewHuffmanAdaptiveCompactCodec(),
		codecs.NewArithmeticAdaptiveCompact2Codec(),
		codecs.NewHuffmanAdaptiveCompact2Codec(),
	}
}

func testPositions() []testPosition {
	return []testPosition{
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
			name: "18 Queens, no pieces on back ranks",
			fen:  "8/1rkbqQ1q/Q3Qq1q/1qq4q/Q7/Q4QK1/QQqQq1BR/8 b - - 12 86",
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
			name: "Last move while the pieces on the starting ranks still are only in their original colors.",
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
		{
			name: "e4 (en passant)",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		},
		{
			name: "e5 (en passant)",
			fen:  "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
		},
		{
			name: "f4 (en passant not in first two plies)",
			fen:  "rnbqkbnr/pp3ppp/8/2ppp3/2PPPP2/8/PP4PP/RNBQKBNR b KQkq f3 0 4",
		},
		{
			name: "f5 (en passant not in first two plies)",
			fen:  "rnbqkbnr/pp4pp/8/2pppp2/2PPPP2/8/PP4PP/RNBQKBNR w KQkq f6 0 5",
		},
		{
			name: "Last ply where all pieces on home ranks are of their original colors",
			fen:  "rnbqk1nr/pppp1ppp/8/4p2Q/4P3/b7/PPPP1PPP/RNB1KBNR w KQkq - 2 3",
		},
		{
			name: "First ply where a piece on a home rank can be of an opposing color",
			fen:  "rnbqk1nr/pppp1Qpp/8/4p3/4P3/b7/PPPP1PPP/RNB1KBNR b KQkq - 0 3",
		},
	}
}

func TestCodecsRoundTrip(t *testing.T) {
	codecs := testCodecs()
	positions := testPositions()
	for _, codec := range codecs {
		for _, position := range positions {
			t.Run(fmt.Sprintf("%s: %s", codec.Name(), position.name), func(t *testing.T) {
				state, _ := chess.ParseFen(position.fen)
				encoded := codec.Encode(state)
				decoded := codec.Decode(encoded)
				decodedFen := decoded.ToFen()
				if position.fen != decodedFen {
					t.Errorf("Decode(Encode(%q)) = %q; expected %q", position.fen, decodedFen, position.fen)
				}
				if state.ToFen() != position.fen {
					t.Errorf("Encode(%q) changed the state; expected state not to change", position.fen)
				}
			})
		}
	}
}

func BenchmarkCodecs(b *testing.B) {
	codecs := testCodecs()
	positions := testPositions()
	states := make([]*chess.State, len(positions))
	for i, position := range positions {
		states[i], _ = chess.ParseFen(position.fen)
	}
	for _, codec := range codecs {
		b.Run(codec.Name(), func(b *testing.B) {
			for b.Loop() {
				for _, state := range states {
					encoded := codec.Encode(state)
					codec.Decode(encoded)
				}
			}
		})
	}
}
