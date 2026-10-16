package bitfield_test

import (
	"slices"
	"testing"

	"example.com/go-chess/internal/bitfield"
)

func TestIntToBoolSlice(t *testing.T) {
	tests := []struct {
		name     string
		x        uint64
		n        int
		expected []bool
	}{
		{name: "An empty slice", x: 1, n: 0, expected: []bool{}},
		{name: "A single zero bit", x: 0, n: 1, expected: []bool{false}},
		{name: "A single one bit", x: 1, n: 1, expected: []bool{true}},
		{name: "A sequence of bits", x: 0b111001101, n: 12, expected: []bool{true, false, true, true, false, false, true, true, true, false, false, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := bitfield.IntToBoolSlice(tt.x, tt.n)
			if !slices.Equal(actual, tt.expected) {
				t.Errorf("IntToBoolSlice(%d, %d) = %#v; expected %#v", tt.x, tt.n, actual, tt.expected)
			}
		})
	}
}

func TestBitFieldSingleBit(t *testing.T) {
	tests := []struct {
		name string
		bit  bool
	}{
		{name: "zero", bit: false},
		{name: "one", bit: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bitfield.BitField{}
			b.WriteBool(tt.bit)
			actual := b.ReadBool()
			if actual != tt.bit {
				t.Errorf("BitField.WriteBool(%v) -> BitField.ReadBool() = %v; expected %v", tt.bit, actual, tt.bit)
			}
		})
	}
}

func TestBitField(t *testing.T) {
	tests := []struct {
		name string
		x    uint64
		n    int
	}{
		{name: "Sequence 1", x: 0b01001100011100001111, n: 20},
		{name: "Sequence 2", x: 0b00001111000111010011, n: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name+": int -> int", func(t *testing.T) {
			b := bitfield.BitField{}
			b.WriteInt(tt.x, tt.n)
			actual := b.ReadInt(tt.n)
			if actual != tt.x {
				t.Errorf("BitField.WriteInt(%d, %d) -> BitField.ReadInt(%d) = %d; expected %d", tt.x, tt.n, tt.n, actual, tt.x)
			}
		})
		t.Run(tt.name+": int -> bool slice", func(t *testing.T) {
			b := bitfield.BitField{}
			slice := bitfield.IntToBoolSlice(tt.x, tt.n)
			b.WriteInt(tt.x, tt.n)
			actual := b.ReadBoolSlice(tt.n)
			if !slices.Equal(actual, slice) {
				t.Errorf("BitField.WriteInt(%d, %d) -> BitField.ReadBoolSlice(%d) = %#v; expected %#v", tt.x, tt.n, tt.n, actual, slice)
			}
		})
		t.Run(tt.name+": bool slice -> int", func(t *testing.T) {
			b := bitfield.BitField{}
			slice := bitfield.IntToBoolSlice(tt.x, tt.n)
			b.WriteBoolSlice(slice)
			actual := b.ReadInt(tt.n)
			if actual != tt.x {
				t.Errorf("BitField.WriteBoolSlice(%#v) -> BitField.ReadInt(%d) = %d; expected %d", slice, tt.n, actual, tt.x)
			}
		})
		t.Run(tt.name+": bool slice -> bool slice", func(t *testing.T) {
			b := bitfield.BitField{}
			slice := bitfield.IntToBoolSlice(tt.x, tt.n)
			b.WriteBoolSlice(slice)
			actual := b.ReadBoolSlice(tt.n)
			if !slices.Equal(actual, slice) {
				t.Errorf("BitField.WriteBoolSlice(%#v) -> BitField.ReadBoolSlice(%d) = %#v; expected %#v", slice, tt.n, actual, slice)
			}
		})
	}
}
