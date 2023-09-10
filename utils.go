package main

import (
	"math/rand"
	"strconv"
)

func createRandomString() string {
	runes := []rune("abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, 6)
	for i := range s {
		s[i] = runes[rand.Intn(len(runes))]
	}

	return string(s)
}

func strToInt(s string) (int, error) {
	intVar, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return intVar, nil
}
