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

    return cids, nil
}

func fetchAndParseMetadata(cid string) (*Metadata, error) {
    url := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var metadata Metadata
    if err := json.Unmarshal(body, &metadata); err != nil {
        return nil, err
    }
	fmt.Println(&metadata)
    return &metadata, nil
}

func insertMetadata(db *sql.DB, metadata *Metadata) error {
    query := `INSERT INTO metadata (image, description, name) VALUES ($1, $2, $3)`
    _, err := db.Exec(query, metadata.Image, metadata.Description, metadata.Name)
    return err
}

func main() {
    cids, err := readCIDsFromFile("ipfs_cids.csv")
    if err != nil {
        log.Fatalf("Error reading CSV: %v", err)
    }

    // Database connection setup (modify with your connection details)
    db, err := sql.Open("postgres", "user=yourusername password=yourpassword dbname=yourdbname sslmode=disable")
    if err != nil {
        log.Fatalf("Error connecting to the database: %v", err)
    }
    defer db.Close()

    for _, cid := range cids {
        metadata, err := fetchAndParseMetadata(cid)
        if err != nil {
            log.Printf("Error fetching metadata for CID %s: %v", cid, err)
            continue
        }

        if err := insertMetadata(db, metadata); err != nil {
            log.Printf("Error inserting metadata for CID %s: %v", cid, err)
        }
    }
}
