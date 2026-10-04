package fxhash

import (
	"encoding/binary"
	"math/bits"
)

const PtrSize = 4 << (^uintptr(0) >> 63)

// MCG (Steele & Vigna)
var K uint64
var rotateBits uint64

func init() {
	K = get_seed()
	if PtrSize == 8 {
		rotateBits = 26
	} else {
		rotateBits = 15
	}

}

func get_seed() uint64 {
	if PtrSize == 8 {
		return 0xf1357aea2e62a9c5
	}
	return 0x93d765dd

}

const SEED1 uint64 = 0x243f6a8885a308d3
const SEED2 uint64 = 0x13198a2e03707344
const ANTI_ZER uint64 = 0xa4093822299f31d0

func multiplyMix(x, y uint64) uint64 {
	hi, lo := bits.Mul64(x, y)
	return lo ^ hi
}

type FxHasher struct {
	hash uint64
}

func NewFxHasher(seed uint64) FxHasher {
	return FxHasher{
		hash: seed,
	}
}

func (f *FxHasher) Add(val uint64) {
	f.hash = val * K
}

func (f *FxHasher) Update(bytes []byte) {
	f.UpdateU64(hashBytes(bytes))
}

func (f *FxHasher) UpdateU64(v uint64) {
	f.Add(v)
	if PtrSize == 4 {
		f.Add(v >> 32)
	}
}

func (f *FxHasher) Finish() uint64 {
	return bits.RotateLeft64(f.hash, int(rotateBits))

}

func hashBytes(b []byte) uint64 {
	length := len(b)
	s0 := SEED1
	s1 := SEED2

	if length <= 16 {
		if length >= 8 {
			s0 ^= binary.LittleEndian.Uint64(b[0:8])
			s1 ^= binary.LittleEndian.Uint64(b[length-8:])
		} else if length >= 4 {
			s0 ^= uint64(binary.LittleEndian.Uint32(b[0:4]))
			s1 ^= uint64(binary.LittleEndian.Uint32(b[length-4:]))
		} else if length > 0 {
			lo := uint64(b[0])
			mid := uint64(b[length/2])
			hi := uint64(b[length-1])
			s0 ^= lo
			s1 ^= (hi << 8) | mid
		}
	} else {
		bulk := b[:length-1]
		for len(bulk) >= 16 {
			chunk := bulk[:16]
			x := binary.LittleEndian.Uint64(chunk[0:8])
			y := binary.LittleEndian.Uint64(chunk[8:16])
			t := multiplyMix(s0^x, ANTI_ZER^y)
			s0 = s1
			s1 = t
			bulk = bulk[16:]
		}
		suffix := b[length-16:]
		s0 ^= binary.LittleEndian.Uint64(suffix[0:8])
		s1 ^= binary.LittleEndian.Uint64(suffix[8:16])
	}

	return multiplyMix(s0, s1) ^ uint64(length)
}

func Fxhash(k uint64, seed uint64) uint64 {
	var f = NewFxHasher(seed)
	f.UpdateU64(k)
	return f.Finish()
}
