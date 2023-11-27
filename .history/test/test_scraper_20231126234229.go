package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Metadata struct {
    Image       string `json:"image"`
    Description string `json:"description"`
    Name        string `json:"name"`
}

func readCIDsFromFile(filePath string) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    lines, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    var cids []string
    for _, line := range lines {
        cids = append(cids, line[0]) // Assuming CID is in the first column
    }
	fmt.Printf("CIDS: %s",cids)
    return cids, nil
}

func fetchAndParseMetadata(cid string, wg *sync.WaitGroup, db *sql.DB) { //, db *sql.DB) {
    defer wg.Done()

	// Create a custom HTTP client with a timeout
    client := &http.Client{
        Timeout: time.Second * 10, // 10 seconds timeout
    }

    url := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Printf("Error fetching metadata for CID %s: %v", cid, err)
        return
    }
	resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error fetching metadata for CID %s: %v", cid, err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response for CID %s: %v", cid, err)
        return
    }

    var metadata Metadata
    if err := json.Unmarshal(body, &metadata); err != nil {
        log.Printf("Error parsing metadata for CID %s: %v", cid, err)
        return
    }

    if err := insertMetadata(db, &metadata); err != nil {
    // if err := insertMetadata(&metadata); err != nil {
        log.Printf("Error inserting metadata for CID %s: %v", cid, err)
    }
	fmt.Printf("\nName: %s\n Description: %s\n Image: %s\n", metadata.Name, metadata.Description, metadata.Image)
}

func initDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image TEXT,
		description TEXT,
		name TEXT
	);`

	_, err := db.Exec(query)
	return err
}

func insertMetadata(db *sql.DB, metadata *Metadata) error {
    query := `INSERT INTO metadata (image, description, name) VALUES (?, ?, ?)`
    _, err := db.Exec(query, metadata.Image, metadata.Description, metadata.Name)
    return err
}


func main() {
    cids, err := readCIDsFromFile("ipfs_cids.csv")
    if err != nil {
        log.Fatalf("Error reading CSV: %v", err)
    }

	// sqlite DB connection
    db, err := sql.Open("sqlite3", "file:metadata.db?cache=shared&mode=rwc")
    if err != nil {
        log.Fatalf("Error connecting to the database: %v", err)
    }
    defer db.Close()

    if err := initDB(db); err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }

    var wg sync.WaitGroup
    for _, cid := range cids {
        wg.Add(1)
        go fetchAndParseMetadata(cid, &wg, db)
    }

    wg.Wait()
}