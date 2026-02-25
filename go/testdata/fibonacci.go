// This file calculates fibonacci series
package main

import (
	"fmt"
)

// fibonacci returns a slice containing the fibonacci series up to n terms
func Fibonacci(n int) []int {
	num := make([]int, n)
	num[0] = 0
	num[1] = 1
	for i := 2; i < n; i++ {
		num[i] = num[i-1] + num[i-2]
	}
	return num
}

func main() {
	series := Fibonacci(10)
	fmt.Println("Fibonacci series:", series)
}
