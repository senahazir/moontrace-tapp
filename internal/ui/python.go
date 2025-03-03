// internal/ui/python.go
package ui

import (
	"bufio"
	"os/exec"

	"github.com/rivo/tview"
)

func (v *Views) StreamPythonScript(prompt string, app *tview.Application, verification bool, fileName string, description string) {
	var selectedFiles []string
	for filePath, isSelected := range v.UploadedFiles {
		if isSelected {
			selectedFiles = append(selectedFiles, filePath)
		}
	}
	// args := append([]string{"/Users/senagulhazir/Desktop/demo/moontrace/app/app.py", prompt}, selectedFiles...)
	args := []string{"/Users/senagulhazir/Desktop/demo/moontrace/app/app.py", prompt}
	if verification {
		args = append(args, "--verification")
	}

	if fileName != "" {
		args = append(args, "--fileName", fileName)
	}
	if description != "" {
		args = append(args, "--description", description)
	}
	args = append(args, selectedFiles...)

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
	v.UpdateFileList(v.List, v.CurrDir)
	cmd.Wait()

	app.QueueUpdateDraw(func() {
		app.SetFocus(v.Pages)
	})
}
