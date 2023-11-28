package main

import (
	"context"
	"sync"
	"testing"
)

func BenchmarkWorkerPool(b *testing.B) {
	cids, err := readCIDsFromFile("ipfs_cids.csv")
	if err != nil {
		b.Fatal(err)
	}

	db, err := connectToDB()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	const numWorkers = 10 // adjust as needed
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cidChan := make(chan string)
		// Start workers
		for i := 0; i < numWorkers; i++ {
			go worker(ctx, db, cidChan, &wg)
		}

		// Send CIDs to workers
		for _, cid := range cids {
			wg.Add(1)
			cidChan <- cid
		}
		close(cidChan) // close the channel when all CIDs have been sent

		// Wait for all metadata to be fetched
		wg.Wait()
	}
}

// go test -bench=BenchmarkFetchAndStoreMetadata
func BenchmarkFetchAndStoreMetadata(b *testing.B) {
	// Connect to the database
	db, err := connectToDB()
	if err != nil {
		b.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Create the metadata table
	if err := createMetadataTable(db); err != nil {
		b.Fatalf("Error creating table: %v", err)
	}

	// Read CIDs from the file
	cids, err := readCIDsFromFile(CIDFilePath)
	if err != nil {
		b.Fatalf("Error reading CSV: %v", err)
	}

	// Create a context for the operation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Reset the timer before the operation we want to benchmark
	b.ResetTimer()

	// Run the operation b.N times
	var errorCount int
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := fetchAndStoreMetadata(ctx, db, cids); err != nil {
				errorCount++
			}
		}
	})

	// Report errors after the benchmark
	if errorCount > 0 {
		b.Logf("fetchAndStoreMetadata failed %d times", errorCount)
	}
}
