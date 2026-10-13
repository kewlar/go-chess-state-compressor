package coders

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type Coder interface {
	Name() string
	Encode(state *chess.State) *bitfield.BitField
	Decode(data *bitfield.BitField) *chess.State
}
