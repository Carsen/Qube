package main

import(
	"fmt"

	"github.com/Carsen/Qube/Login"
)

func main()[
	switch Login.Login(true){
		case true:
			fmt.Println("Welcome!")
		case false: fmt.Prinln("Goodbye!")
	}
]
