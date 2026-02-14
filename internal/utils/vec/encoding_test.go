package vec

import (
	"math"
	"testing"
)

func TestFloat64sToFloat32BytesRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []float64
	}{
		{"positive values", []float64{0.1, 0.2, 0.3, 0.4, 0.5}},
		{"negative values", []float64{-0.1, -0.5, -1.0}},
		{"zeros", []float64{0.0, 0.0, 0.0}},
		{"mixed", []float64{-1.5, 0.0, 1.5, 3.14}},
		{"single", []float64{42.0}},
		{"empty", []float64{}},
		{"large embedding", func() []float64 {
			fs := make([]float64, 1024)
			for i := range fs {
				fs[i] = float64(i) * 0.001
			}
			return fs
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Float64sToFloat32Bytes(tt.input)

			if len(b) != len(tt.input)*4 {
				t.Fatalf("expected %d bytes, got %d", len(tt.input)*4, len(b))
			}

			decoded := BytesToFloat32s(b)
			if len(decoded) != len(tt.input) {
				t.Fatalf("expected %d floats, got %d", len(tt.input), len(decoded))
			}

			for i, want := range tt.input {
				wantF32 := float32(want)
				if math.Abs(float64(decoded[i]-wantF32)) > 1e-7 {
					t.Errorf("[%d] got %f, want %f", i, decoded[i], wantF32)
				}
			}
		})
	}
}

func TestFloat32sToBytesRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []float32
	}{
		{"typical embedding", []float32{0.1, 0.2, 0.3, 0.4, 0.5}},
		{"empty", []float32{}},
		{"single", []float32{1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Float32sToBytes(tt.input)

			if len(b) != len(tt.input)*4 {
				t.Fatalf("expected %d bytes, got %d", len(tt.input)*4, len(b))
			}

			decoded := BytesToFloat32s(b)
			if len(decoded) != len(tt.input) {
				t.Fatalf("expected %d floats, got %d", len(tt.input), len(decoded))
			}

			for i, want := range tt.input {
				if decoded[i] != want {
					t.Errorf("[%d] got %f, want %f", i, decoded[i], want)
				}
			}
		})
	}
}

func TestBytesToFloat32sEmptyInput(t *testing.T) {
	result := BytesToFloat32s(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}

	result = BytesToFloat32s([]byte{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}
