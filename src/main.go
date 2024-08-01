package main

import (
	"fmt"

	"github.com/Carsen/Qube/Intro"
)

func main() {

	switch Intro.Intro() {
	case true:
		fmt.Println("Congrats!")
	case false:
		break
	}
}
