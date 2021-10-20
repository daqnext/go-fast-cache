package go_fast_cache

import "math/rand"

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func genRandStr(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
