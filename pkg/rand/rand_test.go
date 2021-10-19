package rand

import (
	"sync"
	"testing"
)

func TestRand(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				StringK(64, KindAll)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
