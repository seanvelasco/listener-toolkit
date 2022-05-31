package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func CreateTable(db *sql.DB) {

	log.Println("Creating SQL table for the first time")

	query := `CREATE TABLE diagnostic_results (
		"barcode" TEXT,
		"result" TEXT,
		"created_at" TEXT
	)`
	statement, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = statement.Exec()

	if err != nil {
		log.Fatal(err.Error())
	}

}

func InitializeDatabase() (database *sql.DB) {
	log.Println("Initializing database")
	os.Remove("sqlite.db")

	file, err := os.Create("sqlite.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()

	log.Println("Successfully initialized database")

	db, _ := sql.Open("sqlite3", "sqlite.db")

	defer db.Close()

	CreateTable(db)

	// insertData(db)

	return db

}

func (database *DB) Insert(data interface{}) {
	query := `INSERT INTO diagnostic_results (barcode, result, created_at) VALUES (?, ?, ?)`

	statement, err := database.Prepare(query)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = statement.Exec("123456789", "positive", time.Now().String())
	if err != nil {
		log.Fatal(err.Error())
	}

}
