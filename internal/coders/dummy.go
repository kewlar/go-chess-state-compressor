package coders

import "example.com/go-chess/internal/bitfield"

type DummyCoder struct{}

func NewDummyCoder() *DummyCoder {
	return &DummyCoder{}
}

func (c *DummyCoder) Name() string {
	return "Dummy"
}

func (c *DummyCoder) FromFen(fen string) *bitfield.BitField {
	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *DummyCoder) ToFen(data *bitfield.BitField) string {
	l := data.Length/8 - 1
	return string(data.Data[:l])
}
