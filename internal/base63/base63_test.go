package base63

import (
	"errors"
	"math"
	"testing"
)

func TestEncode_Positive(t *testing.T) {
	tests := []struct {
		name string
		num  uint64
		want string
	}{
		{
			name: "zero value",
			num:  0,
			want: "aaaaaaaaaa",
		},
		{
			name: "single digit 1",
			num:  1,
			want: "aaaaaaaaab",
		},
		{
			name: "last single digit (62)",
			num:  62,
			want: "aaaaaaaaa_",
		},
		{
			name: "first two-digit number (63)",
			num:  63,
			want: "aaaaaaaaba",
		},
		{
			name: "max 10-char value",
			num:  MaxVal,
			want: "__________",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.num)
			if err != nil {
				t.Fatalf("Encode(%d) unexpected error: %v", tt.num, err)
			}
			if got != tt.want {
				t.Errorf("Encode(%d) = %q, want %q", tt.num, got, tt.want)
			}
		})
	}
}

func TestEncode_Negative(t *testing.T) {
	tests := []struct {
		name    string
		num     uint64
		wantErr error
	}{
		{
			name:    "overflow MaxVal + 1",
			num:     MaxVal + 1,
			wantErr: ErrOverflow,
		},
		{
			name:    "overflow MaxUint64",
			num:     math.MaxUint64,
			wantErr: ErrOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.num)
			if err == nil {
				t.Fatalf("Encode(%d) expected error %v, got result %q", tt.num, tt.wantErr, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Encode(%d) error = %v, want %v", tt.num, err, tt.wantErr)
			}
		})
	}
}

func TestDecode_Positive(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want uint64
	}{
		{
			name: "zero value",
			key:  "aaaaaaaaaa",
			want: 0,
		},
		{
			name: "single digit: 1",
			key:  "aaaaaaaaab",
			want: 1,
		},
		{
			name: "last single digit: 62",
			key:  "aaaaaaaaa_",
			want: 62,
		},
		{
			name: "first two-digit number: 63",
			key:  "aaaaaaaaba",
			want: 63,
		},
		{
			name: "max 10-char value",
			key:  "__________",
			want: MaxVal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.key)
			if err != nil {
				t.Fatalf("Decode(%q) unexpected error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("Decode(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestDecode_Negative(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "empty",
			key:     "",
			wantErr: ErrInvalidLength,
		},
		{
			name:    "too short: 9 chars",
			key:     "aaaaaaaaa",
			wantErr: ErrInvalidLength,
		},
		{
			name:    "too long: 11 chars",
			key:     "aaaaaaaaaaa",
			wantErr: ErrInvalidLength,
		},
		{
			name:    "invalid char: -",
			key:     "aaaaaaaa-a",
			wantErr: ErrInvalidChar,
		},
		{
			name:    "invalid char: @",
			key:     "aaaaaaaa@a",
			wantErr: ErrInvalidChar,
		},
		{
			name:    "invalid char: !",
			key:     "aaaaaaaa!a",
			wantErr: ErrInvalidChar,
		},
		{
			name:    "invalid char: space",
			key:     "aaaaaaaa a",
			wantErr: ErrInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.key)
			if err == nil {
				t.Errorf("Decode(%q) expected error %v, got result %d", tt.key, tt.wantErr, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Decode(%q) error = %v, want %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	numbers := []uint64{0, 1, 62, 63, 100, 12345678, 999999999, MaxVal}

	for _, num := range numbers {
		encoded, err := Encode(num)
		if err != nil {
			t.Fatalf("Encode(%d) failed: %v", num, err)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode(%q) failed: %v", encoded, err)
		}

		if decoded != num {
			t.Errorf("Invalid Decode: original %d, got %d", num, decoded)
		}
	}
}
