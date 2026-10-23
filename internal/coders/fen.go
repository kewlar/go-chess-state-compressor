package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type FenCoder struct{}

func NewFenCoder() *FenCoder {
	return &FenCoder{}
}

func (c *FenCoder) Name() string {
	return "FEN"
}

func (c *FenCoder) Encode(state *chess.State) *bitfield.BitField {
	fen := state.ToFen()
	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *FenCoder) Decode(bits *bitfield.BitField) *chess.State {
	l := bits.Length/8 - 1
	state, err := chess.ParseFen(string(bits.Data[:l]))
	if err != nil {
		panic(err)
	}
	return state
}
