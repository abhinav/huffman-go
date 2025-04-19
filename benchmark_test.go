package huffman

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkLabel(b *testing.B) {
	bases := []int{2, 8, 32}
	numItems := []int{10, 100, 1000}

	for _, base := range bases {
		for _, numItem := range numItems {
			name := fmt.Sprintf("base=%d/numItems=%d", base, numItem)
			b.Run(name, func(b *testing.B) {
				freqs := make([]int, numItem)
				for i := range freqs {
					freqs[i] = rand.Intn(100)
				}

				b.ResetTimer()

				for range b.N {
					got := Label(base, freqs)
					if len(got) != len(freqs) {
						b.Fatalf("unexpected length: got=%d, want=%d", len(got), len(freqs))
					}
				}
			})
		}
	}
}
