package qDB

import (
	"go.mills.io/bitcask/v2"

	"log"
)

func openDB() {
	opts := []Option{
		WithMaxKeySize(32),
		WithMaxValueSize(32)
	}
	db, _ := bitcask.Open("./db", opts)
	defer db.Close()
}
