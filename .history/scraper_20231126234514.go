package main

import (
    // "database/sql"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "sync"

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
	fmt.Printf("CIDS: %s",cids)
    return cids, nil
}

func fetchAndParseMetadata(cid string, wg *sync.WaitGroup) { //, db *sql.DB) {
    defer wg.Done()

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

func main() {
    cids, err := readCIDsFromFile("ipfs_cids.csv")
    if err != nil {
        log.Fatalf("Error reading CSV: %v", err)
    }

    // Database connection setup (modify with your connection details)
    // db, err := sql.Open("postgres", "user=yourusername password=yourpassword dbname=yourdbname sslmode=disable")
    // if err != nil {
    //     log.Fatalf("Error connecting to the database: %v", err)
    // }
    // defer db.Close()

    var wg sync.WaitGroup
    for _, cid := range cids {
        wg.Add(1)
        // go fetchAndParseMetadata(cid, &wg, db)
        go fetchAndParseMetadata(cid, &wg) // 
    }

    wg.Wait()
}