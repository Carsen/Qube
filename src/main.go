package main

import (
	"Qube/Intro"
	"fmt"
)

func main() {

	switch Intro.intro() {
	case true:
		fmt.Println("Congrats!")
	case false:
		break
	}
}

