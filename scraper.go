package main

import (
	// "database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
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
	fmt.Printf("CIDS: %s", cids)
	return cids, nil
}

func fetchAndParseMetadata(cid string, wg *sync.WaitGroup) { //, db *sql.DB) {
	// defer wg.Done()

	url := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
	resp, err := http.Get(url)
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

	//if err := insertMetadata(db, &metadata); err != nil {
	// if err := insertMetadata(&metadata); err != nil {
	//     log.Printf("Error inserting metadata for CID %s: %v", cid, err)
	// }
	fmt.Printf("\nName: %s\n Description: %s\n Image: %s\n", metadata.Name, metadata.Description, metadata.Image)
}

// func insertMetadata(db *sql.DB, metadata *Metadata) error {
//     query := `INSERT INTO metadata (image, description, name) VALUES ($1, $2, $3)`
//     _, err := db.Exec(query, metadata.Image, metadata.Description, metadata.Name)
//     return err
// }

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}


func main() {
	cids, err := readCIDsFromFile("ipfs_cids.csv")
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	// // Database connection setup (modify with your connection details)
	// // db, err := sql.Open("postgres", "user=yourusername password=yourpassword dbname=yourdbname sslmode=disable")
	// // if err != nil {
	// //     log.Fatalf("Error connecting to the database: %v", err)
	// // }
	// // defer db.Close()

	var wg sync.WaitGroup
	// for _, cid := range cids {
	//     wg.Add(1)
	//     // go fetchAndParseMetadata(cid, &wg, db) // actual db call
	//     go fetchAndParseMetadata(cid, &wg) // so I can print results to console
	// }

	// wg.Wait()

	const maxConcurrency = 1 // adjust as needed
	sem := make(chan struct{}, maxConcurrency)
	s := time.Now().UnixNano()

	for _, cid := range cids {
		// for i := 0; i < 10000; i++ {
		// cid := generateRandomString(i + 1)
		wg.Add(1)
		sem <- struct{}{} // will block if there are already maxConcurrency goroutines
		go func(cid string) {
			defer wg.Done()
			defer func() { <-sem }() // release a spot when this goroutine finishes
			fetchAndParseMetadata(cid, &wg)
			// fmt.Printf("cid: %s\n", cid)
			// time.Sleep(1 * time.Millisecond)
		}(cid)
	}

	wg.Wait()
	fmt.Println(time.Now().UnixNano() - s)
}
