package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanCoder struct {
	huffmanTree *huffman.Node
	huffmanDict map[rune]*huffman.Node
}

func NewHuffmanCoder() *HuffmanCoder {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	huffmanTree := huffman.BuildHuffmanTree(frequencies)

	return &HuffmanCoder{
		huffmanTree: huffmanTree,
		huffmanDict: huffmanTree.BuildDictionary(),
	}
}

func (c *HuffmanCoder) Name() string {
	return "Huffman"
}

func (c *HuffmanCoder) Encode(state *chess.State) *bitfield.BitField {
	bf := &bitfield.BitField{}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(state.Board)

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteBoolSlice(c.huffmanDict[rune(piece.Type())].Bits)
	}

	bf.WriteBool(bool(state.SideToMove))
	bf.WriteBoolSlice(state.Castling.BoolSlice())
	bf.WriteInt(uint64(state.EnPassant), 7)
	bf.WriteInt(uint64(state.Halfmove), 8)
	bf.WriteInt(uint64(state.Fullmove), 14)

	return bf
}

func (c *HuffmanCoder) Decode(data *bitfield.BitField) *chess.State {
	state := chess.NewState()

	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := tools.CountTrue(occupancy)

	// Piece placement: colors.
	colors := data.ReadBoolSlice(pieceCount)

	// Piece placement: pieces.
	pieces := make([]rune, 0, pieceCount)
	for range pieceCount {
		// Read Huffman code.
		node := c.huffmanTree
		for !node.IsLeaf() {
			bit := data.ReadBool()
			node = node.Children[bit]
		}
		pieces = append(pieces, node.Value)
	}

	state.SideToMove = chess.Color(data.ReadBool())
	state.Castling = chess.NewCastling(data.ReadBoolSlice(4))
	state.EnPassant = chess.Square(data.ReadInt(7))
	state.Halfmove = int(data.ReadInt(8))
	state.Fullmove = int(data.ReadInt(14))

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	return state
}
