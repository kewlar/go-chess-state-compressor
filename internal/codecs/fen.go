package codecs

import (
	"github.com/kewlar/go-chess-state-compressor/internal/bitfield"
	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

type FenCodec struct{}

func NewFenCodec() *FenCodec {
	return &FenCodec{}
}

func (c *FenCodec) Name() string {
	return "FEN"
}

func (c *FenCodec) Encode(state *chess.State) *bitfield.BitField {
	fen := state.ToFen()
	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *FenCodec) Decode(bits *bitfield.BitField) *chess.State {
	l := bits.Length/8 - 1
	state, err := chess.ParseFen(string(bits.Data[:l]))
	if err != nil {
		panic(err)
	}
	return state
}
