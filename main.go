package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/l10n-center/api/src/router"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"

	"encoding/base64"

	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/source/file"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	DBURL := os.Getenv("DB")
	if len(DBURL) == 0 {
		log.Print("INFO: db is not set, use default")
		DBURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	}
	SECRET := os.Getenv("SECRET")
	if len(SECRET) == 0 {
		log.Print("INFO: secret is not set, generate random")
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		buf := make([]byte, 20)
		rnd.Read(buf)
		SECRET = base64.URLEncoding.EncodeToString(buf)
	}
	log.Print("INFO: connecting to db")
	db, err := sql.Open("postgres", DBURL)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	defer db.Close()

	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)

	log.Print("INFO: migrating")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migration",
		"postgres", driver)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Printf("INFO: %s", err)
		} else {
			log.Fatalf("ERROR: %s", err)
		}
	}
	log.Print("INFO: ok")
	r := router.New(db, []byte(SECRET))
	http.ListenAndServe(":3000", r)
}
