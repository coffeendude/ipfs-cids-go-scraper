package main

import (
	"context"
	"os"
	"sync"
	"testing"
)

func BenchmarkWorkerPool(b *testing.B) {
	cids, err := readCIDsFromFile("ipfs_cids.csv")
	if err != nil {
		b.Fatal(err)
	}

	// Retrieve database connection details from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Connect to the database
	db, err := connectToDB(host, port, user, password, database, sslmode)
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
	// Retrieve database connection details from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Connect to the database
	db, err := connectToDB(host, port, user, password, database, sslmode)
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
	for i := 0; i < b.N; i++ {
		if err := fetchAndStoreMetadata(ctx, db, cids); err != nil {
			b.Fatalf("Error fetching and storing metadata: %v", err)
		}
	}

}
