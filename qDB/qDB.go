package qDB

import (
	"go.mills.io/bitcask/v2"

	"log"
)

func openDB() {
	opts := []Option{
		MaxKeySize(32),
		MaxValueSize(32),
	}
	db, _ := bitcask.Open("./db", opts)
	defer db.Close()
}
