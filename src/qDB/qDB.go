package qDB

import(
	"log"
	"github.com/Carsen/Qube"
	"github.com/tidwall/buntdb"
)

func openDB(){
	db, err := buntdb.Open("auth.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
