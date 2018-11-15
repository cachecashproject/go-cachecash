package util

import "encoding/binary"

func Uint32ToLE(x uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, x)
	return buf
}

func Uint64ToLE(x uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, x)
	return buf
}
