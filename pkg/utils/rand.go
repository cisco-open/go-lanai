package utils

import (
	"math/rand"
	"time"
)

const (
	CharsetAlphanumeric RandomCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphabetic   RandomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumeric      RandomCharset = "0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type RandomCharset string

func RandomString(length int) string {
	return RandomStringWithCharset(length, CharsetAlphanumeric)
}

func RandomStringWithCharset(length int, charset RandomCharset) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

