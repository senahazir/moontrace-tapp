// internal/ui/views.go
package ui

import (
	"log"
	"moontrace/ascii"
	"os"

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
		CurrDir:       "/Users/senagulhazir/Desktop/counter/counter",
	}

	f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	v.Logger = log.New(f, "", log.LstdFlags)

	v.createViews()
	v.setupPanels()
	v.setupKeyBindings()
	v.UpdateFileList(v.List, v.CurrDir)

	return v
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
	instructionView.SetText(" CtrlP : Switch Response/File pages\n Tab   : Switch between panels\n CtrlC : Exit \n ->    : View file content \n <-    : Parent dictionary ")
	instructionView.SetTextColor(tcell.ColorDarkGreen)
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
	instructionView.SetText(" CtrlP : Switch Response/File pages\n Tab   : Switch between panels\n CtrlC : Exit \n ->    : View file content \n <-    : Parent dictionary ")
	instructionView.SetTextColor(tcell.ColorDarkGreen)
	v.TopPanel.AddItem(placeholder, 1, 1, false)
	v.TopPanel.AddItem(v.TraceView, 2, 1, false) // You'll need to store traceView in Views struct
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
