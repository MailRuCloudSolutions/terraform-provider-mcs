package mcs

import (
	"fmt"
	"testing"
)

func TestRandomName(t *testing.T) {
	rndName := randomName(5)
	if len(rndName) != 5 {
		t.Fatal(fmt.Sprintf("Got wrong result length: %d, expected: 5", len(rndName)))
	}
}
