package main

import (
	"fmt"
	"log"

	"github.com/Carsen/Qube/Login"
	"github.com/rivo/tview"
)

func main() {
	switch Login.Login(true) {
	case true:
		app := tview.NewApplication()

		if err := app.SetRoot(tview.NewBox(), true).EnableMouse(true).Run(); err != nil {
			log.Fatal(err)
		}
	case false:
		fmt.Println("Goodbye!")
	}
}
