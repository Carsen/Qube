package main

import (
	"fmt"
	"log"

	"github.com/Carsen/Qube/Login"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	switch Login.Login(true) {
	case true:
		app := tview.NewApplication()

		primTextView := func(text string) tview.Primitive {
			return tview.NewTextView().
				SetDynamicColors(true).
				SetTextColor(tcell.ColorLime).
				SetTextAlign(tview.AlignCenter).
				SetText(text)
		}
		//		primTextArea := func(text string) tview.Primitive {
		//			return tview.NewTextArea().
		//
		//		}

		grid := tview.NewGrid().
			SetRows(1, 0, 20).
			SetColumns(30, 0, 30).
			SetBorders(true).
			AddItem(primTextView("Qube Network Tool"), 0, 0, 1, 3, 0, 0, false)
		//			AddItem(primTextView(strconv.Itoa(QCom.IfaceAmt())), 2, 0, 1, 3, 0, 0, false)

		grid.AddItem(primTextView("Side Tool"), 0, 0, 0, 0, 0, 0, false).
			AddItem(primTextView("Main Tool"), 1, 0, 1, 3, 0, 0, false).
			AddItem(primTextView("Extra Tool"), 0, 0, 0, 0, 0, 0, false)

		grid.AddItem(primTextView("Side Tool"), 1, 0, 1, 1, 0, 100, false).
			AddItem(primTextView("Main Tool"), 1, 1, 1, 1, 0, 100, false).
			AddItem(primTextView("Extra Tool"), 1, 2, 1, 1, 0, 100, false)

		grid.AddItem(primTextView("Interfaces:"), 2, 0, 1, 2, 0, 0, false).
			AddItem(primTextView("Carsen"), 2, 2, 1, 1, 0, 0, false)

		if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
			log.Fatal(err)
		}
	case false:
		fmt.Println("Goodbye!")
	}
}
