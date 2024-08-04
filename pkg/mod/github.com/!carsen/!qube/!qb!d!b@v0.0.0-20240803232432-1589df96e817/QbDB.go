package QbDB

import (
	"bytes"
	"log"

	"go.mills.io/bitcask/v2"
)

func CheckForKey(usrk []byte) bool {
	db, err := bitcask.Open("./db")
	defer db.Close()
	if err == nil {
		t := db.Has(usrk)
		return t
	} else {
		log.Fatal(err)
		return false
	}
}

func ValueMatchesKey(userk []byte, userp []byte) bool {
	var checker bool = false
	db, err1 := bitcask.Open("./db")
	defer db.Close()
	if err1 == nil {
		get, err2 := db.Get(userk)
		if err2 == nil {
			checker = bytes.Equal(userp, get)
			return checker
		} else {
			log.Fatal(err2)
			checker = false
			return checker
		}
	} else {
		log.Fatal(err1)
		checker = false
		return checker
	}
}
