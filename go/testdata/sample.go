// This is a sample file for computing statistics
package main

import (
	"fmt"
)

// This func computes stats counts
func ComputeStats(data []int) int {
	count := 0
	for _, v := range data {
		count += v
	}
	fmt.Println(count)
	return count / len(data)
}
