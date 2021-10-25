package vips

import (
	"math"
	"unsafe"
)

func ptrToBytes(ptr unsafe.Pointer, size int) []byte {
	return (*[math.MaxInt32]byte)(ptr)[:int(size):int(size)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
