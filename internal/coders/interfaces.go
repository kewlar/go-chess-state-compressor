package coders

import "example.com/go-chess/internal/bitfield"

type Coder interface {
	Name() string
	FromFen(fen string) *bitfield.BitField
	ToFen(data *bitfield.BitField) string
}
