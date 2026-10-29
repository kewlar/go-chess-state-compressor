package codecs

import (
	"example.com/go-chess/internal/bitfield"
	"example.com/go-chess/internal/chess"
)

type Codec interface {
	Name() string
	Encode(state *chess.State) *bitfield.BitField
	Decode(data *bitfield.BitField) *chess.State
}
