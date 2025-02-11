package main

import (
	"moontrace/internal/ui"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	views := ui.InitializeViews(app)

	if err := app.SetRoot(views.MainFlex, true).
		SetFocus(views.UserInput).
		Run(); err != nil {
		panic(err)
	}
}
