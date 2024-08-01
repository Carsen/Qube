package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type auth struct {
	usern string
	passw string
}

func main() {
	switch start(true) {
	case true:
		fmt.Println("Welcome!")
	case false:
		fmt.Println("Goodbye")
	}
}

func start(running bool) bool {
	var checker bool = false
	var i int = -2
	for running == true {
		cls()
		fmt.Println("Hello, and welcome to Qube!")
		fmt.Println("Want to take a ride? y/n + Enter")
		var answer string
		fmt.Scanln(&answer)

		if answer == "y" {
			i = 5
			cls()
		} else if answer == "n" {
			cls()
			fmt.Println("I'm sorry to see you go so soon. We hope to see you back!")
			i = -1
			checker = false
			break
		} else {
			i = -2
		}
		if i == -2 {
		}
		for i > 0 {
			fmt.Print("Please enter username: ")
			var inUsern string
			fmt.Scanln(&inUsern)
			fmt.Print("Please enter password: ")
			var inPassw string
			fmt.Scanln(&inPassw)
			db := hashInput("Carsen", "Ebert")
			usr := hashInput(inUsern, inPassw)
			switch authCompare(db, usr) {
			case true:
				cls()
				checker = true
				return checker

			case false:
				cls()
				fmt.Println("Try again!")
				i--
			}
		}
		if i == 0 {
			cls()
			fmt.Println("Too many tries!")
			checker = false
			break
		}
	}
	for running == false {
		checker = false
		break
	}
	fmt.Println(checker)
	return checker
}

func hashInput(u string, p string) auth {
	var tmpUsern string
	var tmpPassw string

	hash := sha256.New()
	hash.Write([]byte(u))
	tmpUsern = hex.EncodeToString(hash.Sum(nil))
	hash.Reset()
	hash.Write([]byte(p))
	tmpPassw = hex.EncodeToString(hash.Sum(nil))
	hash.Reset()

	s := auth{
		usern: tmpUsern,
		passw: tmpPassw,
	}
	return s
}
func authCompare(a auth, b auth) bool {
	if a == b {
		return true
	} else if a != b {
		return false
	} else {
		return false
	}
}

func cls() {
	fmt.Print("\033[2J")
}
