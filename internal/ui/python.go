// internal/ui/python.go
package ui

import (
	"bufio"
	"os/exec"

	"github.com/rivo/tview"
)

func (v *Views) StreamPythonScript(prompt string, app *tview.Application) {
	var selectedFiles []string
	for filePath, isSelected := range v.UploadedFiles {
		if isSelected {
			selectedFiles = append(selectedFiles, filePath)
		}
	}
	args := append([]string{"/Users/senagulhazir/Desktop/counter/app.py", prompt}, selectedFiles...)
	cmd := exec.Command("python3", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		app.QueueUpdateDraw(func() {
			v.Response.SetText("[red]Error:[white] Failed to start process.")
		})
		return
	}
	if err := cmd.Start(); err != nil {
		app.QueueUpdateDraw(func() {
			v.Response.SetText("[red]Error:[white] Could not start Python script.")
		})
		return
	}

	scanner := bufio.NewScanner(stdout)
	var responseBuffer string

	for scanner.Scan() {
		line := scanner.Text()
		responseBuffer += line + "\n"

		app.QueueUpdateDraw(func() {
			v.Response.SetText(responseBuffer)
		})
	}
	cmd.Wait()

	app.QueueUpdateDraw(func() {
		app.SetFocus(v.Pages)
	})
}
