// internal/ui/views.go
package ui

import (
	"fmt"
	"log"
	"moontrace/ascii"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Views struct {
	App           *tview.Application
	MainFlex      *tview.Flex
	LeftPanel     *tview.Flex
	RightPanel    *tview.Flex
	List          *tview.List
	UserInput     *tview.InputField
	Pages         *tview.Pages
	FileContent   *tview.TextView
	Response      *tview.TextView
	TopPanel      *tview.Flex
	TraceView     *tview.TextView
	Logger        *log.Logger
	UploadedFiles map[string]bool
	CurrDir       string
	InputHistory  []string
	HistoryIndex  int
}

func InitializeViews(app *tview.Application) *Views {
	v := &Views{
		App:           app,
		UploadedFiles: make(map[string]bool),
		CurrDir:       "/Users/senagulhazir/Desktop/demo/counter",
	}

	f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	v.Logger = log.New(f, "", log.LstdFlags)

	v.createViews()
	v.setupPanels()
	v.setupKeyBindings()
	v.UpdateFileList(v.List, v.CurrDir)

	return v
}

func (v *Views) UpdateVerificationFileList(list *tview.List, dir string) {
	list.Clear()
	dir = filepath.Clean(dir)

	files, err := os.ReadDir(dir)
	if err != nil {
		list.AddItem("❌ Error loading files", "", 0, nil)
		return
	}

	if dir != "/" {
		list.AddItem("🗄️ ..", "", 0, nil)
	}

	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		fileName := file.Name()
		prefix := "📄 "
		if file.IsDir() {
			prefix = "🗄️ "
			list.AddItem(prefix+fileName, "", 0, nil)
			continue
		}

		size := fileInfo.Size()
		sizeStr := ""
		switch {
		case size < 1024:
			sizeStr = fmt.Sprintf("%d B", size)
		case size < 1024*1024:
			sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
		default:
			sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		}

		fullPath := filepath.Join(dir, fileName)
		displayName := fmt.Sprintf("%s%s (%s)", prefix, fileName, sizeStr)

		if v.UploadedFiles[fullPath] {
			displayName = fmt.Sprintf("%s%s (%s) ✓", prefix, fileName, sizeStr)
		}

		list.AddItem(displayName, "", 0, nil)
	}
}

// view for verification dialog
func (v *Views) ShowVerificationDialog() {
	// Create main layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true)
	flex.SetTitle("Generate Verification")
	flex.SetBorderColor(tcell.NewHexColor(0x87CEFA))
	flex.SetBackgroundColor(tcell.ColorBlack)

	form := tview.NewForm()
	form.SetBackgroundColor(tcell.ColorBlack)

	var moduleName string
	form.AddInputField("Verification File Name:", "", 40, nil, func(text string) {
		moduleName = text
	})
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetLabelColor(tcell.NewHexColor(0x87CEFA))
	form.SetBorder(true)
	form.SetFieldTextColor(tcell.ColorAntiqueWhite)

	form.SetLabelColor(tcell.ColorGreen)

	v.App.SetFocus(form)

	var description string

	descriptionInput := tview.NewInputField().SetLabel("Design Description:").SetFieldWidth(100).SetChangedFunc(func(text string) {
		description = text
	})
	form.AddFormItem(descriptionInput)

	verificationList := tview.NewList()
	verificationList.SetMainTextColor(tcell.ColorWhiteSmoke)
	verificationList.SetSecondaryTextColor(tcell.NewHexColor(0x87CEFA))
	verificationList.SetBorder(true)
	verificationList.SetTitle("Select Files for Verification")
	verificationList.SetBackgroundColor(tcell.ColorBlack)

	v.UpdateVerificationFileList(verificationList, v.CurrDir)

	descriptionInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			v.App.SetFocus(verificationList)
			return nil
		}
		return event
	})
	verificationList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		fileName := strings.TrimPrefix(mainText, "🗄️ ")
		fileName = strings.TrimPrefix(fileName, "📄 ")
		fileName = strings.Split(fileName, " (")[0]

		fullPath := filepath.Join(v.CurrDir, fileName)
		fileInfo, err := os.Stat(fullPath)

		if err == nil && !fileInfo.IsDir() {
			// Toggle selection
			v.UploadedFiles[fullPath] = !v.UploadedFiles[fullPath]
			v.UpdateVerificationFileList(verificationList, v.CurrDir)
		} else if err == nil && fileInfo.IsDir() && fileName != ".." {
			// Navigate into directory
			v.CurrDir = fullPath
			v.UpdateVerificationFileList(verificationList, v.CurrDir)
		} else if fileName == ".." {
			// Navigate up one level
			if v.CurrDir != "/" {
				v.CurrDir = filepath.Dir(v.CurrDir)
				v.UpdateVerificationFileList(verificationList, v.CurrDir)
			}
		}
	})

	// todo: makes buttn functional
	buttonsForm := tview.NewForm()
	buttonsForm.SetBackgroundColor(tcell.ColorBlack)
	buttonsForm.SetButtonsAlign(tview.AlignCenter)

	buttonsForm.AddButton("Generate", func() {
		var selectedFiles []string
		for filePath, isSelected := range v.UploadedFiles {
			if isSelected {
				selectedFiles = append(selectedFiles, filePath)
			}
		}
		fileName := moduleName
		if fileName == "" {
			fileName = "verification_tb.cpp"
		} else if !strings.HasSuffix(fileName, "cpp") {
			fileName += ".cpp"
		}

		message := fmt.Sprintf("Generating verification testbench...\n"+
			"Module: %s\n"+
			"Description: %s\n"+
			"Selected Files: %d",
			fileName,
			description,
			len(selectedFiles))

		v.Response.SetText(message)
		// change the message to a json object that can help generate a verification file
		v.Pages.SwitchToPage("response")
		v.App.SetRoot(v.MainFlex, true)
		prompt := "Generate a comprehensive verification testbench for this hardware design."
		go v.StreamPythonScript(prompt, v.App, true, fileName, description)

	})

	buttonsForm.AddButton("Cancel", func() {
		v.App.SetRoot(v.MainFlex, true)
	})
	verificationList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			v.App.SetFocus(buttonsForm)
			return nil
		}
		currentIndex := verificationList.GetCurrentItem()
		if currentIndex < 0 {
			return event
		}
		mainText, _ := verificationList.GetItemText(currentIndex)
		switch event.Key() {
		case tcell.KeyLeft:
			if v.CurrDir != "/" {
				v.CurrDir = filepath.Dir(v.CurrDir)
				v.UpdateVerificationFileList(verificationList, v.CurrDir)
			}
			return nil

		case tcell.KeyRight:
			fileName := strings.TrimPrefix(mainText, "🗄️ ")
			fileName = strings.TrimPrefix(fileName, "📄 ")
			fileName = strings.Split(fileName, " (")[0]

			fullPath := filepath.Join(v.CurrDir, fileName)
			fileInfo, err := os.Stat(fullPath)

			if err == nil && fileInfo.IsDir() {
				v.CurrDir = fullPath
				v.UpdateVerificationFileList(verificationList, v.CurrDir)
			}
			return nil
		}

		return event
	})

	// buttonsForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if event.Key() == tcell.KeyTab {
	// 		v.App.SetFocus(buttonsForm)
	// 		return nil
	// 	}
	// 	return event
	// })

	flex.AddItem(form, 7, 1, true)
	flex.AddItem(verificationList, 0, 3, false)
	flex.AddItem(buttonsForm, 3, 1, false)
	v.App.SetRoot(flex, true)
}
func (v *Views) createViews() {
	// Create ASCII view
	moon_ascii := ascii.ConvertImage("/Users/senagulhazir/Desktop/terminal_app/tviewapp/moon.png")
	asciiTextView := tview.NewTextView()
	asciiTextView.SetText(moon_ascii)
	asciiTextView.SetDynamicColors(true)
	asciiTextView.SetTextColor(tcell.ColorDarkGreen)
	asciiTextView.SetTextAlign(tview.AlignLeft)
	asciiTextView.SetWrap(false)

	// Create trace view
	traceTextView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tcell.NewHexColor(0x87CEFA))
	traceTextView.SetText(" ▀▄▀▄▀▄ MOONTRACE 🌝 ▄▀▄▀▄")
	traceTextView.SetSize(50, 50)
	asciiTextView.SetDynamicColors(true)
	traceTextView.SetTextAlign(tview.AlignCenter)
	traceTextView.SetBackgroundColor(tcell.ColorBlack)
	v.TraceView = traceTextView

	// Create response view
	v.Response = tview.NewTextView()
	v.Response.SetDynamicColors(true)
	v.Response.SetLabel("[#87CEFA]Moontrace:[white] ")
	v.Response.SetText("Moontrace response will appear here...")
	v.Response.SetTextColor(tcell.NewHexColor(0x87CEFA))
	v.Response.SetTextStyle(tcell.StyleDefault.Italic(true))
	v.Response.SetBorder(true)
	v.Response.SetTitle("Program Answer")
	v.Response.SetBackgroundColor(tcell.ColorBlack)
	v.Response.SetScrollable(true)
	v.Response.SetRegions(true)

	// Create file content view
	v.FileContent = tview.NewTextView()
	v.FileContent.SetDynamicColors(true)
	v.FileContent.SetBorder(true)
	v.FileContent.SetTitle("File Content")
	v.FileContent.SetScrollable(true)
	v.FileContent.SetRegions(true)

	// Create pages
	v.Pages = tview.NewPages()
	v.Pages.AddPage("file", v.FileContent, true, true)
	v.Pages.AddPage("response", v.Response, true, false)

	// Create file list
	v.List = tview.NewList()
	v.List.SetMainTextColor(tcell.ColorWhiteSmoke)
	v.List.SetSecondaryTextColor(tcell.NewHexColor(0x87CEFA))
	v.List.SetBorder(true).SetTitle("Files")

	// Create user input
	v.UserInput = tview.NewInputField()
	v.UserInput.SetLabel("You: ")
	v.UserInput.SetFieldBackgroundColor(tcell.ColorBlack)
	v.UserInput.SetLabelColor(tcell.NewHexColor(0x87CEFA))
	v.UserInput.SetPlaceholder("Please Enter Your Question")
	v.UserInput.SetBorder(true)
	v.UserInput.SetTitle(" User Input ")
	v.UserInput.SetFieldBackgroundColor(tcell.ColorAntiqueWhite)
	v.UserInput.SetFieldTextColor(tcell.ColorBlack)
	v.UserInput.SetPlaceholderStyle(tcell.StyleDefault.Italic(true))
	v.UserInput.SetFieldWidth(0)

	// Create top panel
	v.TopPanel = tview.NewFlex()
	v.TopPanel.SetDirection(tview.FlexRow)
	placeholder := tview.NewTextView()
	placeholder.SetBorderColor(tcell.ColorBlack)
	instructionView := tview.NewTextView()
	instructionView.SetText(" CtrlP : Switch Response/File pages\n Tab   : Switch between panels\n CtrlC : Exit \n ->    : View file content \n <-    : Parent dictionary\n CtrlG/CtrlB:  ")
	instructionView.SetTextColor(tcell.ColorDeepPink)
	v.TopPanel.AddItem(placeholder, 1, 1, false)
	v.TopPanel.AddItem(traceTextView, 2, 1, false)
	v.TopPanel.AddItem(instructionView, 7, 1, false)
	v.TopPanel.SetBackgroundColor(tcell.ColorBlack)

	// Setup panels
	v.LeftPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.TopPanel, 0, 1, false).
		AddItem(v.List, 0, 3, false)
	v.LeftPanel.SetBackgroundColor(tcell.ColorBlack)

	v.RightPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.UserInput, 10, 20, true).
		AddItem(v.Pages, 30, 20, false)
	v.RightPanel.SetBackgroundColor(tcell.ColorBlack)

	// Create main flex
	box := tview.NewBox().SetBackgroundColor(tcell.ColorBlack)
	v.MainFlex = tview.NewFlex().
		AddItem(box, 0, 0, false).
		AddItem(v.LeftPanel, 0, 1, false).
		AddItem(v.RightPanel, 0, 2, true)
}

func (v *Views) SetBackgrounds() {
	v.UserInput.SetBackgroundColor(tcell.ColorBlack)
	v.TopPanel.SetBackgroundColor(tcell.ColorBlack)
	v.MainFlex.SetBackgroundColor(tcell.ColorBlack)
	v.LeftPanel.SetBackgroundColor(tcell.ColorBlack)
	v.RightPanel.SetBackgroundColor(tcell.ColorBlack)
	v.List.SetBackgroundColor(tcell.ColorBlack)
	v.FileContent.SetBackgroundColor(tcell.ColorBlack)
	v.Response.SetBackgroundColor(tcell.ColorBlack)
	v.Pages.SetBackgroundColor(tcell.ColorBlack)
}
func (v *Views) setupPanels() {
	// Create top panel
	v.TopPanel = tview.NewFlex()
	v.TopPanel.SetDirection(tview.FlexRow)
	placeholder := tview.NewTextView()
	placeholder.SetBorderColor(tcell.ColorBlack)
	instructionView := tview.NewTextView()
	instructionView.SetText(" CtrlP : Switch Response/File pages\n Tab   : Switch between panels\n CtrlC : Exit \n ->    : View file content \n <-    : Parent dictionary\n CtrlG : Verification Page\n CtrlB : Back Page  ")
	instructionView.SetTextColor(tcell.ColorGreen)
	v.TopPanel.AddItem(placeholder, 1, 1, false)
	v.TopPanel.AddItem(v.TraceView, 2, 1, false)
	v.TopPanel.AddItem(instructionView, 7, 1, false)
	v.TopPanel.SetBackgroundColor(tcell.ColorBlack)

	// Create left panel
	v.LeftPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.TopPanel, 0, 1, false).
		AddItem(v.List, 0, 3, false)
	v.LeftPanel.SetBackgroundColor(tcell.ColorBlack)

	// Create right panel
	v.RightPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.UserInput, 10, 20, true).
		AddItem(v.Pages, 30, 20, false)
	v.RightPanel.SetBackgroundColor(tcell.ColorBlack)

	// Create box and main flex
	box := tview.NewBox().SetBackgroundColor(tcell.ColorBlack)
	v.MainFlex = tview.NewFlex().
		AddItem(box, 0, 0, false).
		AddItem(v.LeftPanel, 0, 1, false).
		AddItem(v.RightPanel, 0, 2, true)
}
