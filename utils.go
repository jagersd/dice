package main

import (
	"math/rand"
)

func createRandomString() string {
	runes := []rune("abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, 6)
	for i := range s {
		s[i] = runes[rand.Intn(len(runes))]
	}

	return string(s)
}
