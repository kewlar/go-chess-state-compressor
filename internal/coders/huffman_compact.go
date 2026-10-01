package coders

import (
	"strconv"
	"strings"

	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/huffman"
)

type HuffmanCompactCoder struct {
	castlingOptions []rune
	huffmanTree     *huffman.Node
	huffmanDict     map[rune]*huffman.Node
}

func NewHuffmanCompactCoder() *HuffmanCompactCoder {
	var (
		castlingOptions = []rune{'K', 'Q', 'k', 'q'}
		frequencies     = map[rune]int{
			'p': 16,
			'k': 2,
			'q': 2,
			'r': 4,
			'b': 4,
			'n': 4,
		}
	)

	huffmanTree := huffman.BuildHuffmanTree(frequencies)

	return &HuffmanCompactCoder{
		castlingOptions: castlingOptions,
		huffmanTree:     huffmanTree,
		huffmanDict:     huffmanTree.BuildDictionary(),
	}
}

func (c *HuffmanCompactCoder) Name() string {
	return "HuffmanCompact"
}

func (c *HuffmanCompactCoder) FromFen(fen string) *bitfield.BitField {
	// PARSE FEN

	parts := strings.Fields(fen)

	// En passant square.
	enPassant := byte(64) // 64 == no en passant
	if parts[3] != "-" {
		file := parts[3][0] - 'a'
		rank := parts[3][1] - '1'
		enPassant = (7-rank)*8 + file
	}

	// Piece placement
	board := [64]rune{}
	cell := 0
	for _, r := range parts[0] {
		if r == '/' {
			continue
		}

		if '1' <= r && r <= '8' {
			cell += int(r - '0')
			continue
		}

		board[cell] = r
		cell++
	}

	// Encode en passant in board configuration
	if enPassant != 64 {
		enPassantPawn := enPassant ^ 8
		homeRankSquare := enPassant ^ 16
		board[enPassantPawn], board[homeRankSquare] = board[homeRankSquare], board[enPassantPawn]
	}

	occupancy := uint64(0)
	color := uint32(0)
	pieces := []rune{}
	for cell, r := range board {
		if r == 0 {
			continue
		}

		pieceType := rune(int(r) | 32)

		occupancy |= 1 << cell

		isWhite := int(r)&32 == 0
		if isWhite {
			color |= 1 << len(pieces)
		}
		pieces = append(pieces, pieceType)
	}

	// Side to move.
	whiteToMove := parts[1] == "w"

	// Castling availability.
	castling := []bool{}
	for _, option := range c.castlingOptions {
		castling = append(castling, strings.ContainsRune(parts[2], option))
	}

	// Halfmove clock.
	halfMoves, _ := strconv.Atoi(parts[4])

	// Fullmove number.
	fullMoves, _ := strconv.Atoi(parts[5])

	// GENERATE BITS

	bf := &bitfield.BitField{}

	// Piece placement.
	bf.WriteInt(occupancy, 64)
	bf.WriteInt(uint64(color), len(pieces))
	for _, piece := range pieces {
		bf.WriteBoolSlice(c.huffmanDict[piece].Bits)
	}

	// Side to move.
	bf.WriteBool(whiteToMove)

	// Castling availability.
	for idx := range c.castlingOptions {
		bf.WriteBool(castling[idx])
	}

	// En passant square.
	// Encoded in piece placement data.

	// Halfmove clock & fullmove number.
	moveNum := fullMoves*151 + halfMoves
	bf.WriteInt(uint64(moveNum), 21)

	return bf
}

func (c *HuffmanCompactCoder) ToFen(data *bitfield.BitField) string {
	// Piece placement: occupancy.
	occupancy := data.ReadBoolSlice(64)
	pieceCount := 0
	for _, occupied := range occupancy {
		if occupied {
			pieceCount++
		}
	}

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

	// Side to move.
	isWhite := data.ReadBool()

	// Castling availability.
	castling := map[rune]bool{}
	for _, option := range c.castlingOptions {
		castling[option] = data.ReadBool()
	}

	// En passant square.
	// Decoded from piece placement data.

	// Halfmove clock & fullMove number.
	moveNum := data.ReadInt(21)
	fullMoves, halfMoves := moveNum/151, moveNum%151

	// GENERATE FEN

	fen := strings.Builder{}

	// Piece placement.
	board := [64]rune{}
	idx := 0
	for i, occupied := range occupancy {
		if occupied {
			piece, color := pieces[idx], colors[idx]
			if color { // White
				piece &= 0b1101_1111
			}
			board[i] = piece
			idx++
		}
	}

	// Restore en passant pawn.
	enPassant := 64
	from, to := 0, 7
	if !isWhite {
		from, to = 56, 63
	}
	for i := from; i <= to; i++ {
		if board[i]|32 == 'p' {
			enPassant = i ^ 16
			enPassantPawn := enPassant ^ 8
			board[enPassantPawn], board[i] = board[i], board[enPassantPawn]
			break
		}
	}

	empty := 0
	idx = 0
	for i, piece := range board {
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
			fen.WriteRune(piece)
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
	if isWhite {
		fen.WriteRune('w')
	} else {
		fen.WriteRune('b')
	}
	fen.WriteRune(' ')

	// Castling availability.
	none := true
	for _, option := range c.castlingOptions {
		if castling[option] {
			fen.WriteRune(option)
			none = false
		}
	}
	if none {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// En passant square.
	if enPassant <= 63 {
		rank := 7 - (enPassant >> 3)
		file := enPassant & 7
		fen.WriteRune('a' + rune(file))
		fen.WriteRune('1' + rune(rank))
	} else {
		fen.WriteRune('-')
	}
	fen.WriteRune(' ')

	// Halfmove clock.
	fen.WriteString(strconv.Itoa(int(halfMoves)))
	fen.WriteRune(' ')

	// Fullmove number.
	fen.WriteString(strconv.Itoa(int(fullMoves)))

	return fen.String()
}
