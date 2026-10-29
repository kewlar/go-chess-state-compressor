package codecs

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/huffman"
	"example.com/go-chess/internal/tools"
)

type HuffmanCompactCodec struct {
	huffmanTree *huffman.Node
	huffmanDict map[rune]*huffman.Node
}

func NewHuffmanCompactCodec() *HuffmanCompactCodec {
	frequencies := map[rune]int{
		'p': 16,
		'k': 2,
		'q': 2,
		'r': 4,
		'b': 4,
		'n': 4,
	}

	huffmanTree := huffman.BuildHuffmanTree(frequencies)

	return &HuffmanCompactCodec{
		huffmanTree: huffmanTree,
		huffmanDict: huffmanTree.BuildDictionary(),
	}
}

func (c *HuffmanCompactCodec) Name() string {
	return "Huffman (compact)"
}

func (c *HuffmanCompactCodec) Encode(state *chess.State) *bitfield.BitField {
	// Encode en passant in board configuration
	board := state.Board
	if state.EnPassant.IsValid() {
		enPassantPawn := state.EnPassant ^ 8
		backRankSquare := state.EnPassant ^ 16
		board[enPassantPawn], board[backRankSquare] = board[backRankSquare], board[enPassantPawn]
	}

	// Encode board in separate bitboards for occupancy, color, and a list of pieces
	occupancy, color, pieces := separateBoardInformation(board)

	bits := &bitfield.BitField{}

	// Piece placement.
	bits.WriteInt(occupancy, 64)
	bits.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bits.WriteBoolSlice(c.huffmanDict[rune(piece.Type())].Bits)
	}

	// Side to move.
	bits.WriteBool(bool(state.SideToMove))

	// Castling availability.
	bits.WriteBoolSlice(state.Castling.BoolSlice())

	// Halfmove clock & fullmove number.
	moveNum := state.Fullmove*(chess.MaxHalfmove+1) + state.Halfmove
	bits.WriteInt(uint64(moveNum), 21)

	return bits
}

func (c *HuffmanCompactCodec) Decode(bits *bitfield.BitField) *chess.State {
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

	// Side to move.
	state.SideToMove = chess.Color(bits.ReadBool())

	// Castling availability.
	state.Castling = chess.NewCastling(bits.ReadBoolSlice(4))

	// Halfmove clock & fullmove number.
	moveNum := int(bits.ReadInt(21))
	state.Fullmove, state.Halfmove = moveNum/(chess.MaxHalfmove+1), moveNum%(chess.MaxHalfmove+1)

	// Piece placement.
	placePieces(&state.Board, occupancy, pieces, colors)

	// Restore en passant pawn.
	state.EnPassant = chess.InvalidSquare
	from, to := 0, 7
	if state.SideToMove == chess.Black {
		from, to = 56, 63
	}
	for i := from; i <= to; i++ {
		if state.Board[i].Type() == chess.Pawn {
			state.EnPassant = chess.Square(i ^ 16)
			enPassantPawn := state.EnPassant ^ 8
			state.Board[enPassantPawn], state.Board[i] = state.Board[i], state.Board[enPassantPawn]
			break
		}
	}

	return state
}
