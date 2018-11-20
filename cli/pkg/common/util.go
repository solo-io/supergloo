package common

import (
	"math/rand"
	"time"
)

const lcAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lcAlphaNumeric[rand.Intn(len(lcAlphaNumeric))]
	}
	return string(b)
}

func Contains(a []string, s string) bool {
	for _, n := range a {
		if s == n {
			return true
		}
	}
	return false
}
