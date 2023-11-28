package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coffeendude/ipfs-cids-go-scraper/metadata"
	_ "github.com/lib/pq"
)

func StartServer(db *sql.DB) {
	http.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {
		handleTokensRequest(db, w, r)
	})
	http.HandleFunc("/tokens/", func(w http.ResponseWriter, r *http.Request) {
		handleTokensRequest(db, w, r)
	})

	http.ListenAndServe(":8080", nil)
}

func handleTokensRequest(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	log.Println("Received request:", r.Method, r.URL.Path)
	path := strings.TrimPrefix(r.URL.Path, "/tokens")
	if path == "" {
		// Handle /tokens
		log.Println("Fetching all metadata")
		metadata, err := getAllMetadata(db)
		if err != nil {
			log.Println("Error fetching all metadata:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(metadata)

	} else {
		// Handle /tokens/<cid>
		cid := strings.TrimPrefix(path, "/")
		metadata, err := getMetadataForCID(db, cid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(cid)
		json.NewEncoder(w).Encode(metadata)
	}
}

func getAllMetadata(db *sql.DB) ([]metadata.Metadata, error) {
	rows, err := db.Query("SELECT cid, image, description, name FROM metadata")
	if err != nil {
		return nil, fmt.Errorf("error querying metadata: %w", err)
	}
	defer rows.Close()

	metadatas := []metadata.Metadata{}
	for rows.Next() {
		var m metadata.Metadata
		if err := rows.Scan(&m.Cid, &m.Image, &m.Description, &m.Name); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		metadatas = append(metadatas, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return metadatas, nil
}

func getMetadataForCID(db *sql.DB, cid string) (*metadata.Metadata, error) {
	row := db.QueryRow("SELECT cid, image, description, name FROM metadata WHERE cid = $1", cid)

	var m metadata.Metadata
	if err := row.Scan(&m.Cid, &m.Image, &m.Description, &m.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // return nil, nil if no rows were found
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &m, nil
}
