package bitfield

import (
	"fmt"
	"slices"
	"strings"
)

type BitField struct {
	Length   int
	ReadHead int
	Data     []byte
}

// bits2bytes returns how many full bytes do n bits need.
func bits2bytes(n int) int {
	bytes := n >> 3
	if n&7 != 0 {
		bytes++
	}
	return bytes
}

// IntToBoolSlice convers n bits from a number x into a slice of bools, LSB first.
func IntToBoolSlice(x uint64, n int) []bool {
	out := []bool{}
	for i := range n {
		out = append(out, x&(1<<i) != 0)
	}

	return out
}

// String renders BitField to string of 1s and 0s.
func (b *BitField) String() string {
	parts := make([]string, 0, bits2bytes(b.Length)+1)
	parts = append(parts, fmt.Sprintf("%d bits:", b.Length))
	for _, bt := range b.Data {
		part := []rune(fmt.Sprintf("%08b", bt))
		slices.Reverse(part)
		parts = append(parts, string(part))
	}

	return strings.Join(parts, " ")
}

// Byte returns current byte index as indicated by length.
func (b *BitField) Byte() int {
	return b.Length >> 3
}

// Bit returns current bit index in current byte as indicated by length.
func (b *BitField) Bit() int {
	return b.Length & 7
}

// WriteInt writes n bits of number i into BitField, LSB first.
func (b *BitField) WriteInt(i uint64, n int) {
	b.pad(n)
	for range n {
		value := byte(i & 1)
		b.Data[b.Byte()] |= value << b.Bit()
		b.Length++
		i >>= 1
	}
}

// WriteBool writes a single bit into BitField.
func (b *BitField) WriteBool(x bool) {
	b.pad(1)
	value := byte(0)
	if x {
		value = 1
	}
	b.Data[b.Byte()] |= value << b.Bit()
	b.Length++
}

// WriteBoolSlice writes a sequence of bits into Bitfield.
func (b *BitField) WriteBoolSlice(x []bool) {
	for _, y := range x {
		b.WriteBool(y)
	}
}

// ReadBool reads and returns a single bit of the BitField. It adjusts the BitField read head. If it reads out of bounds, it panics.
func (b *BitField) ReadBool() bool {
	if b.ReadHead+1 > b.Length {
		panic("reading out of bounds")
	}
	byte, bit := b.ReadHead>>3, b.ReadHead&7
	b.ReadHead++
	return b.Data[byte]&(1<<bit) != 0
}

// ReadBoolSlice reads and returns a sequence of n bits of the BitField. It adjusts the BitField read head. If it reads out of bounds, it panics.
func (b *BitField) ReadBoolSlice(n int) []bool {
	if b.ReadHead+n > b.Length {
		panic("reading out of bounds")
	}
	out := make([]bool, 0, n)
	for range n {
		byte, bit := b.ReadHead>>3, b.ReadHead&7
		out = append(out, b.Data[byte]&(1<<bit) != 0)
		b.ReadHead++
	}
	return out
}

// ReadInt reads a sequence of n bits of the BitField and returns it as an unsigned integer. It adjusts the BitField read head. If it reads out of bounds, it panics.
func (b *BitField) ReadInt(n int) uint64 {
	if b.ReadHead+n > b.Length {
		panic("reading out of bounds")
	}
	out := uint64(0)
	for i := range n {
		byte, bit := b.ReadHead>>3, b.ReadHead&7
		if b.Data[byte]&(1<<bit) != 0 {
			out |= 1 << i
		}
		b.ReadHead++
	}
	return out
}

// pad pads the BitField with zero bytes needed to write additional n bits into it.
func (b *BitField) pad(n int) {
	lenInBytes := bits2bytes(b.Length)
	newLenInBytes := bits2bytes(b.Length + n)
	for range newLenInBytes - lenInBytes {
		b.Data = append(b.Data, 0)
	}
}
