package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type DummyCoder struct{}

func NewDummyCoder() *DummyCoder {
	return &DummyCoder{}
}

func (c *DummyCoder) Name() string {
	return "Dummy"
}

func (c *DummyCoder) Encode(state *chess.State) *bitfield.BitField {
	fen := state.ToFen()
	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *DummyCoder) Decode(data *bitfield.BitField) *chess.State {
	l := data.Length/8 - 1
	state, err := chess.ParseFen(string(data.Data[:l]))
	if err != nil {
		panic(err)
	}
	return state
}
