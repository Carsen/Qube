package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/Carsen/Qube/qDB"
)

func main() {
	switch login(true) {
	case true:
		fmt.Println("Welcome!")

	case false:
		fmt.Println("Goodbye")
	}
}

func login(running bool) bool {
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
			var hashUsern []byte
			fmt.Scanln(&inUsern)
			hashUsern = hashInput(inUsern)

			switch qDB.checkForKey(hashUsern) {
			case true:
				fmt.Print("Please enter password: ")
				var inPassw string
				var hashPassw []byte
				fmt.Scanln(&inPassw)
				hashPassw = hashInput(inPassw)

				switch qDB.valueMatchesKey(hashUsern, hashPassw) {
				case true:
					cls()
					checker = true
					return checker
				case false:
					cls()
					fmt.Println("Try again!")
					i--
				}

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

func hashInput(u string) []byte {
	hash := sha256.New()
	defer hash.Reset()
	hash.Write([]byte(u))
	b := hash.Sum(nil)
	return b
}

func cls() {
	fmt.Print("\033[2J")
}
