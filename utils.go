package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

type wsparsed struct {
	s string
	i int
}

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

func parseInterface(input interface{}, requestedType string) wsparsed {
	switch requestedType {
	case "s":
		return wsparsed{s: fmt.Sprintf("%s", input)}
	case "i":
		number, err := strconv.Atoi(fmt.Sprintf("%s", input))
		if err != nil {
			return wsparsed{i: 0}
		}
		return wsparsed{i: number}
	default:
		return wsparsed{s: "", i: 0}
	}
}
