package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	for i := 0; i <= 10000; i++ {
		go counterNext()
	}
}

func TestIsSA(t *testing.T) {
	assert.True(t, IsSA(SAAreaID()))
	assert.False(t, IsSA(CloudAreaID(0, 0)))
}

func BenchmarkSAAreaID(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			SAAreaID()
		}
	})
}
