// internal/ui/keybindings.go
package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (v *Views) setupKeyBindings() {
	// whole app key bindings
	v.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlG:
			v.ShowVerificationDialog()
			return nil
		case tcell.KeyCtrlB:
			v.App.SetRoot(v.MainFlex, true)
		case tcell.KeyCtrlC:
			v.App.Stop()
			return nil
		case tcell.KeyCtrlP:
			if frontPage, _ := v.Pages.GetFrontPage(); frontPage == "file" {
				v.Pages.SwitchToPage("response")
			} else {
				v.Pages.SwitchToPage("file")
			}
			return nil
		}
		return event
	})

	v.UserInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			prompt := v.UserInput.GetText()
			v.InputHistory = append(v.InputHistory, prompt)
			v.HistoryIndex = len(v.InputHistory)
			v.UserInput.SetText("")
			go func() {
				v.StreamPythonScript(prompt, v.App, false, "", "")
			}()
		}
	})

	v.UserInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			v.App.SetFocus(v.List)
			return nil
		case tcell.KeyUp:
			if v.HistoryIndex > 0 {
				v.HistoryIndex--
				v.UserInput.SetText(v.InputHistory[v.HistoryIndex])
			}
			return nil
		case tcell.KeyDown:
			if v.HistoryIndex < len(v.InputHistory)-1 {
				v.HistoryIndex++
				v.UserInput.SetText(v.InputHistory[v.HistoryIndex])
			}
			return nil
		}
		return event
	})

	// Pages key bindings
	v.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			v.App.SetFocus(v.List)
			return nil
		}
		return event
	})

	// List selection function
	v.List.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		fileName := strings.TrimPrefix(mainText, "🗄️ ")
		fileName = strings.TrimPrefix(fileName, "📄 ")
		fileName = strings.TrimPrefix(fileName, "[#4682B4]")
		fileName = strings.TrimSuffix(fileName, "[-]")
		fileName = strings.Split(fileName, " (")[0] // Remove size info

		if !strings.HasSuffix(fileName, "..") {
			fullPath := filepath.Join(v.CurrDir, fileName)
			fileInfo, err := os.Stat(fullPath)
			if err == nil && !fileInfo.IsDir() {
				content, err := os.ReadFile(fullPath)
				if err != nil {
					v.FileContent.SetText("Error reading file: " + err.Error())
					return
				}
				v.FileContent.SetText(string(content))
				v.Pages.SwitchToPage("file")
			}
		}
	})

	// List key bindings
	v.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentIndex := v.List.GetCurrentItem()
		mainText, _ := v.List.GetItemText(currentIndex)

		switch event.Key() {
		case tcell.KeyTab:
			v.App.SetFocus(v.UserInput)
			return nil
		case tcell.KeyLeft:
			if v.CurrDir != "/" {
				newDir := filepath.Dir(v.CurrDir)
				v.CurrDir = newDir
				v.UpdateFileList(v.List, v.CurrDir)
			}
			return event

		case tcell.KeyEnter:
			fileName := strings.TrimPrefix(mainText, "🗄️ ")
			fileName = strings.TrimPrefix(fileName, "📄 ")
			fileName = strings.Split(fileName, " (")[0]

			fullPath := filepath.Join(v.CurrDir, fileName)
			fileInfo, err := os.Stat(fullPath)
			if err == nil && !fileInfo.IsDir() {
				v.UploadedFiles[fullPath] = !v.UploadedFiles[fullPath]
				v.UpdateFileList(v.List, v.CurrDir)
			}
			return event

		case tcell.KeyRight:
			if currentIndex < 0 {
				return event
			}
			v.App.SetFocus(v.Pages)

			fileName := strings.TrimPrefix(mainText, "🗄️ ")
			fileName = strings.TrimPrefix(fileName, "📄 ")
			fileName = strings.Split(fileName, " (")[0]
			fileName = strings.TrimSuffix(fileName, "[-]")
			fileName = strings.TrimPrefix(fileName, "[#4682B4]")

			fullPath := filepath.Join(v.CurrDir, fileName)
			fileInfo, err := os.Stat(fullPath)
			if err == nil {
				if fileInfo.IsDir() {
					v.CurrDir = fullPath
					v.UpdateFileList(v.List, v.CurrDir)
				} else {
					content, err := os.ReadFile(fullPath)
					if err != nil {
						v.FileContent.SetText("Error reading file: " + err.Error())
						return event
					}
					v.FileContent.SetText(string(content))
					v.Pages.SwitchToPage("file")
				}
			}
			return event
		}
		return event
	})
}
