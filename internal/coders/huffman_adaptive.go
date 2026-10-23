package coders

import (
	"fmt"

	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/entropy"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanAdaptiveCoder struct {
	frequencies map[rune]int
}

func NewHuffmanAdaptiveCoder() *HuffmanAdaptiveCoder {
	frequencies := map[rune]int{
		'p': 1600,
		'k': 200,
		'q': 200,
		'r': 400,
		'b': 400,
		'n': 400,
	}

	return &HuffmanAdaptiveCoder{
		frequencies: frequencies,
	}
}

func (c *HuffmanAdaptiveCoder) Name() string {
	return "Huffman (adaptive)"
}

func (c *HuffmanAdaptiveCoder) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	// Huffman coding.
	n := len(pieces)
	freqs := entropy.PrepareFrequencies(n, c.frequencies)
	colorSlice := bitfield.IntToBoolSlice(uint64(color), len(pieces))
	fullOccupancy := bitfield.IntToBoolSlice(occupancy, 64)
	runes := piecesToRunes(pieces)
	idx := 0
	for i := range 64 {
		if fullOccupancy[i] {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(i), n, fullOccupancy, colorSlice, state, runes[:idx])
			r := runes[idx]
			if freqs2[r] == 0 {
				panic(fmt.Sprintf("didn't expect '%s' at square %s", string(r), chess.Square(i).String()))
			}
			tree := huffman.BuildHuffmanTree(freqs2)
			dict := tree.BuildDictionary()
			bits.WriteBoolSlice(dict[r].Bits)
			idx++

			freqs = entropy.UpdateFrequencies(r, n-idx, freqs)
		}
	}

	return bits
}

func (c *HuffmanAdaptiveCoder) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	freqs := entropy.PrepareFrequencies(pieceCount, c.frequencies)
	idx := 0
	for square, occupied := range occupancy {
		if occupied {
			freqs2 := entropy.AdjustFrequencies(freqs, chess.Square(square), pieceCount, occupancy, colors, state, pieces)
			tree := huffman.BuildHuffmanTree(freqs2)

			node := tree
			for !node.IsLeaf() {
				bit := bits.ReadBool()
				node = node.Children[bit]
			}
			piece := node.Value
			pieces = append(pieces, piece)

			idx++

			freqs = entropy.UpdateFrequencies(piece, pieceCount-idx, freqs)
		}
	}

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
