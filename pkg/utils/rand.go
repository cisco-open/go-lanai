package utils

import (
	cryptorand "crypto/rand"
	"math/big"
	"math/rand"
	"time"
)

const (
	CharsetAlphanumeric RandomCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphabetic   RandomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomCharset is a string containing all acceptable UTF-8 characters for random string generation
type RandomCharset string

// RandomString returns a random Alphanumeric string of given "length"
// this function uses "crypto/rand" and fallback to "math/rand"
// It panics if len(charset) > 255, and returns empty string if length is non-positive
func RandomString(length int) string {
	return RandomStringWithCharset(length, CharsetAlphanumeric)
}

// RandomStringWithCharset returns a random string of given "length" containing only characters from given "charset"
// this function uses "crypto/rand" and fallback to "math/rand"
// It returns empty string if length is non-positive, and only the first 256 chars in "charset" are used
func RandomStringWithCharset(length int, charset RandomCharset) string {
	if length <= 0 {
		return ""
	}

	data := make([]byte, length)
	b := make([]byte, 1)
	for i := range data {
		if n, e := cryptorand.Reader.Read(b); e != nil || n < 1 {
			data[i] = charset[rand.Intn(len(charset))] //nolint:gosec // this is fallback method, better than not working
		} else {
			data[i] = charset[int(b[0]) % len(charset)]
		}
	}
	return string(data)
}

// RandomInt64N returns, as an int64, a non-negative uniform number in the half-open interval [0,n).
// This function uses "crypto/rand" and fallback to "math/rand".
// It panics if n <= 0.
func RandomInt64N(n int64) int64 {
	bigInt, e := cryptorand.Int(cryptorand.Reader, big.NewInt(n))
	if e != nil {
		return rand.Int63n(n) //nolint:gosec // this is fallback method, better than not working
	}
	return bigInt.Int64()
}

// RandomIntN returns, as an int64, a non-negative uniform number in the half-open interval [0,n).
// This function uses "crypto/rand" and fallback to "math/rand".
// It panics if n <= 0.
func RandomIntN(n int) int {
	return int(RandomInt64N(int64(n)))
}
