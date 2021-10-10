package main

import (
	"fmt"
	"math"
)

func main() {
	var a, b, c int64
	var x float64
	fmt.Scanln(&a)
	fmt.Scanln(&b)
	fmt.Scanln(&c)
	x = float64((3*a + b - c) / 3)
	fmt.Println(x)
	fmt.Println(math.Ceil(x))
}
