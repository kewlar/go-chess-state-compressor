package coders

import (
	"log"

	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type FenValidator struct{}

func NewFenValidator() *FenValidator {
	return &FenValidator{}
}

func (c *FenValidator) Name() string {
	return "FenValidator"
}

func (c *FenValidator) FromFen(fen string) *bitfield.BitField {
	_, err := chess.ParseFen(fen)
	if err != nil {
		log.Fatal(err)
	}

	return &bitfield.BitField{
		Length: (len(fen) + 1) * 8,
		Data:   append([]byte(fen), 0),
	}
}

func (c *FenValidator) ToFen(data *bitfield.BitField) string {
	l := data.Length/8 - 1
	return string(data.Data[:l])
}
