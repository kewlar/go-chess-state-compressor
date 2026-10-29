package codecs

import (
	"github.com/kewlar/go-chess-state-compressor/internal/bitfield"
	"github.com/kewlar/go-chess-state-compressor/internal/chess"
)

type Codec interface {
	Name() string
	Encode(state *chess.State) *bitfield.BitField
	Decode(data *bitfield.BitField) *chess.State
}
