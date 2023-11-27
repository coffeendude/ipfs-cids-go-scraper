package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkConcurrency(b *testing.B) {
	for _, maxConcurrency := range []int{1, 10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("Concurrency%d", maxConcurrency), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				sem := make(chan struct{}, maxConcurrency)
				var wg sync.WaitGroup

				// s := time.Now().UnixNano()
				for i := 0; i < 10000; i++ {
					cid := generateRandomString(i + 1)
					wg.Add(1)
					sem <- struct{}{} // will block if there are already maxConcurrency goroutines
					go func(cid string) {
						defer wg.Done()
						defer func() { <-sem }() // release a spot when this goroutine finishes
						// fetchAndParseMetadata(cid, &wg)
						// fmt.Printf("cid: %s\n", cid)
						time.Sleep(1 * time.Millisecond)
					}(cid)
				}
				wg.Wait()
			}
		})
	}
}
