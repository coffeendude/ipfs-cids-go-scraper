package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/coffeendude/ipfs-cids-go-scraper/api"
	"github.com/coffeendude/ipfs-cids-go-scraper/metadata"

	_ "github.com/lib/pq"
)

const (
	NumWorkers  = 10
	CIDFilePath = "ipfs_cids.csv"
)

func main() {
	host := flag.String("host", "localhost", "Database host")
	port := flag.String("port", "5432", "Database port")
	user := flag.String("user", "postgres", "Database user")
	password := flag.String("password", "example", "Database password")
	dbname := flag.String("dbname", "postgres", "Database name")
	sslmode := flag.String("sslmode", "disable", "SSL mode")
	flag.Parse()

	db, err := connectToDB(*host, *port, *user, *password, *dbname, *sslmode)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err := createMetadataTable(db); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	cids, err := readCIDsFromFile(CIDFilePath)
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := fetchAndStoreMetadata(ctx, db, cids); err != nil {
		log.Fatalf("Error fetching and storing metadata: %v", err)
	}

	if err := printMetadata(db); err != nil {
		log.Fatalf("Error printing metadata: %v", err)
	}

	api.StartServer(db)
}

func connectToDB(host, port, user, password, dbname, sslmode string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return db, nil
}

func createMetadataTable(db *sql.DB) error {

	_, err := db.Exec(`DROP TABLE IF EXISTS metadata;`)
	if err != nil {
		return fmt.Errorf("error dropping table: %w", err)
	}

	_, err = db.Exec(`
		
        CREATE TABLE metadata (
            cid TEXT PRIMARY KEY,
            image TEXT,
            description TEXT,
            name TEXT
        )
    `)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}
	return nil
}

func readCIDsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	var cids []string
	for _, line := range lines {
		cids = append(cids, line[0])
	}
	return cids, nil
}

func fetchAndStoreMetadata(ctx context.Context, db *sql.DB, cids []string) error {
	cidChan := make(chan string)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < NumWorkers; i++ {
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

	return nil
}

func worker(ctx context.Context, db *sql.DB, cidChan <-chan string, wg *sync.WaitGroup) {
	for cid := range cidChan {
		fetchAndParseMetadata(db, cid)
		wg.Done()
	}
}

func fetchAndParseMetadata(db *sql.DB, cid string) (*metadata.Metadata, error) {
	url := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata for CID %s: %w", cid, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response for CID %s: %w", cid, err)
	}

	var metadata metadata.Metadata
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid JSON for CID %s: %s", cid, string(body))
	}
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("error parsing metadata for CID %s: %w", cid, err)
	}

	metadata.Cid = cid
	if err := storeMetadata(db, &metadata); err != nil {
		return nil, fmt.Errorf("error storing metadata for CID %s: %w", cid, err)
	}

	return &metadata, nil
}

func storeMetadata(db *sql.DB, metadata *metadata.Metadata) error {
	sqlStatement := `
        INSERT INTO metadata (cid, image, description, name)
        VALUES ($1, $2, $3, $4)
		ON CONFLICT (cid) DO UPDATE SET
        image = EXCLUDED.image,
        description = EXCLUDED.description,
        name = EXCLUDED.name`
	_, err := db.Exec(sqlStatement, metadata.Cid, metadata.Image, metadata.Description, metadata.Name)
	if err != nil {
		return fmt.Errorf("error storing metadata: %w", err)
	}
	return nil
}

func printMetadata(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM metadata")
	if err != nil {
		return fmt.Errorf("error querying metadata: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid, image, description, name string
		if err := rows.Scan(&cid, &image, &description, &name); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("Cid: %s,Image: %s, Description: %s, Name: %s\n", cid, image, description, name)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	return nil
}
