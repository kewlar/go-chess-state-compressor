package codecs

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanCodec struct {
	huffmanTree *huffman.Node
	huffmanDict map[rune]*huffman.Node
}

func NewHuffmanCodec() *HuffmanCodec {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	huffmanTree := huffman.BuildHuffmanTree(frequencies)

	return &HuffmanCodec{
		huffmanTree: huffmanTree,
		huffmanDict: huffmanTree.BuildDictionary(),
	}
}

func (c *HuffmanCodec) Name() string {
	return "Huffman"
}

func (c *HuffmanCodec) Encode(state *chess.State) *bitfield.BitField {
	bits := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bits.WriteBoolSlice(c.huffmanDict[rune(piece.Type())].Bits)
	}

	bits.WriteBool(bool(state.SideToMove))
	bits.WriteBoolSlice(state.Castling.BoolSlice())
	bits.WriteInt(uint64(state.EnPassant), 7)
	bits.WriteInt(uint64(state.Halfmove), 8)
	bits.WriteInt(uint64(state.Fullmove), 14)

	return bits
}

func (c *HuffmanCodec) Decode(bits *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := bits.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := bits.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		// Read Huffman code.
		node := c.huffmanTree
		for !node.IsLeaf() {
			bit := bits.ReadBool()
			node = node.Children[bit]
		}
		pieces = append(pieces, node.Value)
	}

	state.SideToMove = chess.Color(bits.ReadBool())
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))
	state.EnPassant = chess.Square(bits.ReadInt(7))
	state.Halfmove = int(bits.ReadInt(8))
	state.Fullmove = int(bits.ReadInt(14))

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
