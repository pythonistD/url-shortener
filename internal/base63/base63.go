package base63

import (
	"errors"
	"fmt"
	"strings"
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	base    = 63
	keyLen  = 10
	MaxVal  = 63*63*63*63*63*63*63*63*63*63 - 1
)

var (
	ErrInvalidLength = errors.New("encoded str must be 10 characters long")
	ErrInvalidChar   = errors.New("contains invalid character")
	ErrOverflow      = errors.New("number exceeds max character capacity")
)

func Encode(num uint64) (string, error) {
	if num > MaxVal {
		return "", ErrOverflow
	}

	var buf [keyLen]byte

	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = symbols[num%base]
		num /= base
	}

	return string(buf[:]), nil
}

func Decode(s string) (uint64, error) {
	if len(s) != keyLen {
		return 0, ErrInvalidLength
	}

	var num uint64

	for i := 0; i < len(s); i++ {
		pos := strings.IndexByte(symbols, s[i])
		if pos < 0 {
			return 0, fmt.Errorf("%w: %q at index %d", ErrInvalidChar, s[i], i)
		}

		num = num*base + uint64(pos)
	}

	return num, nil
}
