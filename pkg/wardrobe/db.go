package wardrobe

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB
var ctx context.Context
var conn *sql.Conn

func InitDB(psqlInfo string) {
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Panic(err)
	}
	db.SetMaxOpenConns(50)
	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
}

func CloseDB() error {
	return db.Close()
}
