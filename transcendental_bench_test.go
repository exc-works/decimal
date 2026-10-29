package decimal

import (
	"fmt"
	"math/big"
	"testing"
)

var (
	benchIntegerResult *big.Int
	benchBoundsResult  fixedInterval
)

func BenchmarkFixedPointBounds(b *testing.B) {
	for _, bits := range []uint{64, 512} {
		num := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(17))
		den := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits/2), big.NewInt(3))
		scale := new(big.Int).Lsh(big.NewInt(1), bits)
		b.Run(fmt.Sprintf("ceil/%d", bits), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				benchIntegerResult = ceilQuotient(num, den)
			}
		})
		b.Run(fmt.Sprintf("fraction/%d", bits), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				benchBoundsResult = fractionBounds(num, den, scale)
			}
		})
	}
}
