package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(min, max int) (string, error) {
	if min <= 0 || max < min {
		return "", fmt.Errorf("invalid range")
	}
	length, err := randomInt(int64(max - min + 1))
	if err != nil {
		return "", err
	}
	length += int64(min)

	var builder strings.Builder
	for i := int64(0); i < length; i++ {
		index, err := randomInt(int64(len(alphabet)))
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[index])
	}
	return builder.String(), nil
}

func randomInt(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}
