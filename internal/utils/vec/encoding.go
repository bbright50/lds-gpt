package vec

import (
	"encoding/binary"
	"math"
)

// Float64sToFloat32Bytes converts a slice of float64 (Bedrock API output)
// to packed little-endian float32 bytes for storage.
func Float64sToFloat32Bytes(fs []float64) []byte {
	buf := make([]byte, len(fs)*4)
	for i, f := range fs {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(float32(f)))
	}
	return buf
}

// Float32sToBytes converts a slice of float32 to packed little-endian bytes.
func Float32sToBytes(fs []float32) []byte {
	buf := make([]byte, len(fs)*4)
	for i, f := range fs {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// BytesToFloat32s converts packed little-endian bytes back to a slice of float32.
func BytesToFloat32s(b []byte) []float32 {
	fs := make([]float32, len(b)/4)
	for i := range fs {
		fs[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return fs
}
