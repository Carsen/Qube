package qDB

import (
	"go.mills.io/bitcask/v2"
)

func checkForKey(usrk []byte) {
	db, _ := bitcask.Open("./db")
	defer db.Close()
	db.Has(usrk)
}
