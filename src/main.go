package main

import (
	"fmt"
	"log"

	"github.com/Carsen/Qube/Login"
	"github.com/Carsen/Qube/QCom"
	"github.com/marcusolsson/tui-go"
)

func main() {
	switch Login.Login(true) {
	case true:
		box := tui.NewHBox(
			tui.NewSpacer(),
			tui.NewPadder(1, 0)
			tui.NewLabel("Qube"),
		)

		ui, err := tui.New(box)
		if err != nil {
			log.Fatal(err)
		}

		ui.SetKeybinding("Esc", func() { ui.Quit() })

		if err := ui.Run(); err != nil {
			log.Fatal(err)
			}
	case false:
		fmt.Println("Goodbye!")
	}
}
