package main

import (
	"fmt"
	"testing"
)

func TestCreateRandomString(t *testing.T) {
	rand1 := createRandomString()
	rand2 := createRandomString()
	if rand1 == rand2 {
		fmt.Printf("First string: %s\n. Second string: %s \n", rand1, rand2)
		t.Error("results are exactly the same.")
	}
}
