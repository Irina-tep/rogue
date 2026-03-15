package main

import (
	"math/rand/v2"
)

// генерирует число от и до включительно
func GeneratorNum(min int, max int) int {
	return rand.IntN(max-min+1) + min
}
