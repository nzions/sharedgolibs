package benchmarks

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// BenchmarkStringConcatenation benchmarks different string concatenation methods
func BenchmarkStringConcatenation(b *testing.B) {
	parts := []string{"hello", "world", "from", "golang", "benchmarks"}

	b.Run("Plus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := ""
			for _, part := range parts {
				result += part
			}
			_ = result
		}
	})

	b.Run("Builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			for _, part := range parts {
				builder.WriteString(part)
			}
			_ = builder.String()
		}
	})

	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := strings.Join(parts, "")
			_ = result
		}
	})
}

// BenchmarkIntToString benchmarks integer to string conversion
func BenchmarkIntToString(b *testing.B) {
	num := 12345

	b.Run("Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := fmt.Sprintf("%d", num)
			_ = result
		}
	})

	b.Run("Itoa", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := strconv.Itoa(num)
			_ = result
		}
	})

	b.Run("FormatInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := strconv.FormatInt(int64(num), 10)
			_ = result
		}
	})
}

// BenchmarkSliceOperations benchmarks various slice operations
func BenchmarkSliceOperations(b *testing.B) {
	b.Run("Append", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var slice []int
			for j := 0; j < 1000; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("PreAlloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 1000)
			for j := 0; j < 1000; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("Index", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				slice[j] = j
			}
		}
	})
}

// BenchmarkMapOperations benchmarks map operations
func BenchmarkMapOperations(b *testing.B) {
	b.Run("StringKeys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]int)
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key_%d", j)
				m[key] = j
			}
		}
	})

	b.Run("IntKeys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[int]int)
			for j := 0; j < 1000; j++ {
				m[j] = j
			}
		}
	})

	b.Run("PreAlloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]int, 1000)
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key_%d", j)
				m[key] = j
			}
		}
	})
}

// BenchmarkSorting benchmarks different sorting approaches
func BenchmarkSorting(b *testing.B) {
	// Generate random data
	generateData := func(size int) []int {
		data := make([]int, size)
		for i := range data {
			data[i] = rand.Intn(10000)
		}
		return data
	}

	b.Run("Small_100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := generateData(100)
			sort.Ints(data)
		}
	})

	b.Run("Medium_1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := generateData(1000)
			sort.Ints(data)
		}
	})

	b.Run("Large_10000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := generateData(10000)
			sort.Ints(data)
		}
	})
}

// BenchmarkHashing benchmarks different hashing algorithms
func BenchmarkHashing(b *testing.B) {
	data := []byte("Hello, World! This is a test string for hashing benchmarks.")

	b.Run("MD5", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hash := md5.Sum(data)
			_ = hash
		}
	})

	b.Run("SHA256", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hash := sha256.Sum256(data)
			_ = hash
		}
	})
}

// BenchmarkChannelOperations benchmarks channel operations
func BenchmarkChannelOperations(b *testing.B) {
	b.Run("Unbuffered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch := make(chan int)
			go func() {
				ch <- 42
			}()
			<-ch
		}
	})

	b.Run("Buffered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch := make(chan int, 1)
			ch <- 42
			<-ch
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("SmallStructs", func(b *testing.B) {
		type Small struct {
			A int
			B string
		}

		for i := 0; i < b.N; i++ {
			s := Small{A: i, B: "test"}
			_ = s
		}
	})

	b.Run("LargeStructs", func(b *testing.B) {
		type Large struct {
			Data [1000]int
			Name string
		}

		for i := 0; i < b.N; i++ {
			l := Large{Name: "test"}
			_ = l
		}
	})

	b.Run("SliceAllocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]int, 1000)
			_ = slice
		}
	})
}

// BenchmarkRecursiveVsIterative compares recursive vs iterative implementations
func BenchmarkRecursiveVsIterative(b *testing.B) {
	// Fibonacci implementations
	var fibRecursive func(int) int
	fibRecursive = func(n int) int {
		if n <= 1 {
			return n
		}
		return fibRecursive(n-1) + fibRecursive(n-2)
	}

	fibIterative := func(n int) int {
		if n <= 1 {
			return n
		}
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}

	b.Run("Recursive_Fib_20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := fibRecursive(20)
			_ = result
		}
	})

	b.Run("Iterative_Fib_20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := fibIterative(20)
			_ = result
		}
	})
}
