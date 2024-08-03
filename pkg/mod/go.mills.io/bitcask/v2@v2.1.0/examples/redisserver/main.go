// Package main implements a simple Redis-like server using Bitcask as the store
package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/redcon"

	"go.mills.io/bitcask/v2"
)

var addr = ":6379"

func main() {
	db, err := bitcask.Open("test.db")
	if err != nil {
		log.Fatalf("error opening database: %s", err)
	}
	defer db.Close()

	go log.Printf("started server at %s", addr)

	if err := redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError(fmt.Sprintf("ERR unknown command %q", string(cmd.Args[0])))
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for %q command", string(cmd.Args[0])))
					return
				}

				if err := db.Put(cmd.Args[1], cmd.Args[2]); err != nil {
					conn.WriteError(fmt.Sprintf("ERR could not write key %q: %s", string(cmd.Args[1]), err.Error()))
					return
				}

				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for %q command", string(cmd.Args[0])))
					return
				}

				val, err := db.Get(cmd.Args[1])
				if err != nil {
					if errors.Is(err, bitcask.ErrKeyNotFound) {
						conn.WriteNull()
					} else {
						conn.WriteError(fmt.Sprintf("ERR could not read key %q: %s", string(cmd.Args[1]), err.Error()))
					}
					return
				}

				conn.WriteBulk(val)
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for %q command", string(cmd.Args[0])))
					return
				}

				if !db.Has(cmd.Args[1]) {
					conn.WriteInt(0)
					return
				}

				if err := db.Delete(cmd.Args[1]); err != nil {
					conn.WriteError(fmt.Sprintf("ERR could not delete key %q: %s", string(cmd.Args[1]), err.Error()))
					return
				}

				conn.WriteInt(1)
			}
		},
		func(conn redcon.Conn) bool { return true },
		func(conn redcon.Conn, err error) {},
	); err != nil {
		log.Fatal(err)
	}
}
