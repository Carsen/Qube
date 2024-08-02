package qDB

import (
	"github.com/Carsen/Qube"
	"github.com/tidwall/buntdb"
	"log"
)

func openDB() {
	db, err := buntdb.Open("auth.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
