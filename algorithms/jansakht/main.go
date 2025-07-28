package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	maxDivisor := int(math.Sqrt(float64(n))) + 1
	for d := 3; d < maxDivisor; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func findRightTruncatablePrimes(n int) []string {
	if n == 1 {
		return []string{"2", "3", "5", "7"}
	}

	//starting with single digit primes
	primes := []string{"2", "3", "5", "7"}
	for length := 2; length <= n; length++ {
		var newPrimes []string
		for _, prime := range primes {
			for digit := '1'; digit <= '9'; digit++ {
				candidate := prime + string(digit)
				num, _ := strconv.Atoi(candidate)
				if isPrime(num) {
					newPrimes = append(newPrimes, candidate)
				}
			}
		}
		primes = newPrimes
	}
	return primes
}

func main() {
	var n int
	fmt.Scan(&n)

	if n == 1 {
		fmt.Println(2)
		fmt.Println(3)
		fmt.Println(5)
		fmt.Println(7)
		return
	}

	primes := findRightTruncatablePrimes(n)
	sort.Strings(primes)
	for _, prime := range primes {
		fmt.Println(prime)
	}
}
