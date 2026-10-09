// migrate prepares the managed database and optionally imports existing local data.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"socialNetwork/pkg/cloud"
	"socialNetwork/pkg/db"
)

func main() {
	source := flag.String("source", "", "existing migrated SQLite database to import into an empty PostgreSQL schema")
	uploads := flag.String("uploads", "", "existing upload directory to copy to the private bucket")
	flag.Parse()
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if err := cloud.Validate(); err != nil {
		log.Fatal(err)
	}
	if err := db.InitDB(); err != nil {
		log.Fatal(err)
	}
	defer db.CloseDB()
	if cloud.Enabled() {
		if err := cloud.EnsurePrivateStorage(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
	if *uploads != "" {
		if !cloud.Enabled() {
			log.Fatal("Supabase storage must be configured to copy uploads")
		}
		entries, err := os.ReadDir(*uploads)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || entry.Name() == ".gitkeep" {
				continue
			}
			file, err := os.Open(filepath.Join(*uploads, entry.Name()))
			if err != nil {
				log.Fatal(err)
			}
			err = cloud.Upload(context.Background(), entry.Name(), mime.TypeByExtension(filepath.Ext(entry.Name())), file)
			file.Close()
			if err != nil {
				log.Fatalf("Upload %s: %v", entry.Name(), err)
			}
		}
	}
	if *source != "" {
		if err := db.ImportSQLite(*source); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Managed schema is ready; requested imports completed.")
}
