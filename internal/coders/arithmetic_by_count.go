package coders

import (
	"log"

	"example.com/go-chess/internal/arithmetic"
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/tools"
)

type ArithmeticByCountCoder struct {
	frequencies map[int]map[rune]int
	intervals   map[int]map[rune]*arithmetic.Interval
}

func NewArithmeticByCountCoder() *ArithmeticByCountCoder {
	frequencies := map[int]map[rune]int{}
	frequencies[32] = map[rune]int{
		'p': 8, // 0.500000
		'r': 2, // 0.125000
		'k': 1, // 0.062500
		'q': 1, // 0.062500
		'b': 2, // 0.125000
		'n': 2, // 0.125000
	}
	frequencies[31] = map[rune]int{
		'p': 8, // 0.490208
		'r': 2, // 0.128847
		'k': 1, // 0.064516
		'q': 1, // 0.064294
		'b': 2, // 0.126827
		'n': 2, // 0.125307
	}
	frequencies[30] = map[rune]int{
		'p': 36, // 0.487161
		'r': 10, // 0.133108
		'k': 5,  // 0.066667
		'q': 5,  // 0.066182
		'b': 9,  // 0.122777
		'n': 9,  // 0.124105
	}
	frequencies[29] = map[rune]int{
		'p': 28, // 0.486476
		'r': 8,  // 0.136694
		'k': 4,  // 0.068966
		'q': 4,  // 0.067146
		'b': 7,  // 0.121757
		'n': 7,  // 0.118961
	}
	frequencies[28] = map[rune]int{
		'p': 65, // 0.498144
		'r': 18, // 0.141425
		'k': 9,  // 0.071429
		'q': 9,  // 0.067279
		'b': 15, // 0.113567
		'n': 14, // 0.108156
	}
	frequencies[27] = map[rune]int{
		'p': 52, // 0.495764
		'r': 15, // 0.144911
		'k': 8,  // 0.074074
		'q': 7,  // 0.067409
		'b': 12, // 0.112016
		'n': 11, // 0.105826
	}
	frequencies[26] = map[rune]int{
		'p': 54, // 0.509796
		'r': 16, // 0.149505
		'k': 8,  // 0.076923
		'q': 7,  // 0.066613
		'b': 11, // 0.102396
		'n': 10, // 0.094768
	}
	frequencies[25] = map[rune]int{
		'p': 39, // 0.508391
		'r': 12, // 0.152676
		'k': 6,  // 0.080000
		'q': 5,  // 0.065960
		'b': 8,  // 0.100955
		'n': 7,  // 0.092018
	}
	frequencies[24] = map[rune]int{
		'p': 50, // 0.521655
		'r': 15, // 0.156145
		'k': 8,  // 0.083333
		'q': 6,  // 0.064018
		'b': 9,  // 0.092365
		'n': 8,  // 0.082484
	}
	frequencies[23] = map[rune]int{
		'p': 59, // 0.522622
		'r': 18, // 0.158365
		'k': 10, // 0.086957
		'q': 7,  // 0.062896
		'b': 10, // 0.089885
		'n': 9,  // 0.079275
	}
	frequencies[22] = map[rune]int{
		'p': 52, // 0.534161
		'r': 15, // 0.159107
		'k': 9,  // 0.090909
		'q': 6,  // 0.060064
		'b': 8,  // 0.083165
		'n': 7,  // 0.072595
	}
	frequencies[21] = map[rune]int{
		'p': 46, // 0.538263
		'r': 14, // 0.160051
		'k': 8,  // 0.095238
		'q': 5,  // 0.058063
		'b': 7,  // 0.079245
		'n': 6,  // 0.069140
	}
	frequencies[20] = map[rune]int{
		'p': 51, // 0.550513
		'r': 14, // 0.156922
		'k': 9,  // 0.100000
		'q': 5,  // 0.054496
		'b': 7,  // 0.074281
		'n': 6,  // 0.063788
	}
	frequencies[19] = map[rune]int{
		'p': 54, // 0.556333
		'r': 15, // 0.155511
		'k': 10, // 0.105263
		'q': 5,  // 0.052268
		'b': 7,  // 0.070409
		'n': 6,  // 0.060216
	}
	frequencies[18] = map[rune]int{
		'p': 60, // 0.569262
		'r': 16, // 0.149103
		'k': 12, // 0.111111
		'q': 5,  // 0.047943
		'b': 7,  // 0.066278
		'n': 6,  // 0.056303
	}
	frequencies[17] = map[rune]int{
		'p': 53, // 0.576670
		'r': 13, // 0.144803
		'k': 11, // 0.117647
		'q': 4,  // 0.044239
		'b': 6,  // 0.063548
		'n': 5,  // 0.053094
	}
	frequencies[16] = map[rune]int{
		'p': 58, // 0.588016
		'r': 14, // 0.137650
		'k': 12, // 0.125000
		'q': 4,  // 0.040280
		'b': 6,  // 0.059616
		'n': 5,  // 0.049438
	}
	frequencies[15] = map[rune]int{
		'p': 63, // 0.592721
		'r': 14, // 0.133347
		'k': 14, // 0.133333
		'q': 4,  // 0.036793
		'b': 6,  // 0.057263
		'n': 5,  // 0.046542
	}
	frequencies[14] = map[rune]int{
		'p': 53, // 0.597402
		'r': 11, // 0.127296
		'k': 13, // 0.142857
		'q': 3,  // 0.033814
		'b': 5,  // 0.054487
		'n': 4,  // 0.044144
	}
	frequencies[13] = map[rune]int{
		'p': 58, // 0.596914
		'r': 12, // 0.124991
		'k': 15, // 0.153846
		'q': 3,  // 0.030201
		'b': 5,  // 0.052214
		'n': 4,  // 0.041834
	}
	frequencies[12] = map[rune]int{
		'p': 104, // 0.593477
		'r': 21,  // 0.120223
		'k': 29,  // 0.166667
		'q': 5,   // 0.027836
		'b': 9,   // 0.051034
		'n': 7,   // 0.040764
	}
	frequencies[11] = map[rune]int{
		'p': 90, // 0.582118
		'r': 19, // 0.120149
		'k': 28, // 0.181818
		'q': 4,  // 0.026479
		'b': 8,  // 0.050330
		'n': 6,  // 0.039106
	}
	frequencies[10] = map[rune]int{
		'p': 102, // 0.567788
		'r': 21,  // 0.116648
		'k': 36,  // 0.200000
		'q': 5,   // 0.027644
		'b': 9,   // 0.049745
		'n': 7,   // 0.038175
	}
	frequencies[9] = map[rune]int{
		'p': 56, // 0.540809
		'r': 12, // 0.119558
		'k': 23, // 0.222222
		'q': 3,  // 0.029783
		'b': 5,  // 0.049722
		'n': 4,  // 0.037907
	}
	frequencies[8] = map[rune]int{
		'p': 41, // 0.505636
		'r': 10, // 0.120962
		'k': 20, // 0.250000
		'q': 3,  // 0.036243
		'b': 4,  // 0.049861
		'n': 3,  // 0.037299
	}
	frequencies[7] = map[rune]int{
		'p': 62, // 0.460269
		'r': 16, // 0.122327
		'k': 38, // 0.285714
		'q': 6,  // 0.043590
		'b': 7,  // 0.051008
		'n': 5,  // 0.037091
	}
	frequencies[6] = map[rune]int{
		'p': 62, // 0.398466
		'r': 20, // 0.129620
		'k': 52, // 0.333333
		'q': 9,  // 0.056579
		'b': 7,  // 0.044059
		'n': 6,  // 0.037943
	}
	frequencies[5] = map[rune]int{
		'p': 26, // 0.320847
		'r': 10, // 0.119776
		'k': 32, // 0.400000
		'q': 6,  // 0.073631
		'b': 4,  // 0.048815
		'n': 3,  // 0.036931
	}
	frequencies[4] = map[rune]int{
		'p': 32, // 0.223901
		'r': 13, // 0.092914
		'k': 71, // 0.500000
		'q': 15, // 0.106095
		'b': 6,  // 0.041356
		'n': 5,  // 0.035734
	}
	frequencies[3] = map[rune]int{
		'p': 484,  // 0.113094
		'r': 388,  // 0.090575
		'k': 2855, // 0.666667
		'q': 544,  // 0.127123
		'b': 5,    // 0.001182
		'n': 6,    // 0.001359
	}
	frequencies[2] = map[rune]int{
		'p': 0, // 0.000000
		'r': 0, // 0.000000
		'k': 1, // 1.000000
		'q': 0, // 0.000000
		'b': 0, // 0.000000
		'n': 0, // 0.000000
	}

	intervals := map[int]map[rune]*arithmetic.Interval{}
	for count := 32; count >= 2; count-- {
		intervals[count] = arithmetic.FrequenciesToIntervals(frequencies[count])
	}

	return &ArithmeticByCountCoder{
		frequencies: frequencies,
		intervals:   intervals,
	}
}

func (c *ArithmeticByCountCoder) Name() string {
	return "Arithmetic (by piece count)"
}

func (c *ArithmeticByCountCoder) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))

	// Arithmetic coding.
	code := arithmetic.Encode(piecesToRunes(pieces), c.intervals[len(pieces)])
	if len(code) > maxBitsForArithmeticCode {
		log.Fatalf("encoded data too long (%d bits)", len(code))
	}
	bits.WriteInt(uint64(len(code)), bitsForArithmeticCodeLength)
	bits.WriteBoolSlice(code)

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	return bits
}

func (c *ArithmeticByCountCoder) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	l := int(bits.ReadInt(bitsForArithmeticCodeLength))
	code := bits.ReadBoolSlice(l)
	pieces := arithmetic.Decode(code, pieceCount, c.intervals[pieceCount])

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
