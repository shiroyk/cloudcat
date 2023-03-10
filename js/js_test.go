package js

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	goroutineNum := 15
	blockNum := 4
	SetScheduler(NewScheduler(Options{InitialVMs: 2, MaxVMs: 4}))
	wg := new(sync.WaitGroup)

	for i := 1; i <= goroutineNum; i++ {
		wg.Add(1)
		go func(i int) {
			timeout := time.Second
			script := "1"
			if i < blockNum {
				script = `while(true){}`
				timeout *= 2
			}

			ctx, _ := context.WithTimeout(context.Background(), timeout)
			defer func() {
				wg.Done()
			}()

			_, err := RunString(ctx, script)
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("%v: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkScheduler(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	wg := sync.WaitGroup{}
	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func() {
			_, _ = RunString(context.Background(), `1`)
			wg.Done()
		}()
	}
	b.StopTimer()
}
