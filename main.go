package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Khởi tạo ứng dụng Fyne
	fmt.Println("Đang khởi tạo ứng dụng...")
	myApp := app.New()
	myWindow := myApp.NewWindow("Codebase Exporter")
	myWindow.Resize(fyne.NewSize(600, 400))

	// --- UI Components ---

	// 1. Source Directory
	sourceLabel := widget.NewLabel("Project Directory:")
	sourceEntry := widget.NewEntry()
	sourceEntry.PlaceHolder = "Select project path..."

	// 2r Output Directory
	outputLabel := widget.NewLabel("Output Directory:")
	outputEntry := widget.NewEntry()
	outputEntry.PlaceHolder = "Select output path..."

	// 3. Extensions
	extLabel := widget.NewLabel("File Extensions (comma separated, empty for all):")
	extEntry := widget.NewEntry()
	extEntry.SetText("go,cpp,h,txt,ipynb,py,js,ts,html,css,java") // Default

	// Status Label
	statusLabel := widget.NewLabel("Ready.")
	statusLabel.Wrapping = fyne.TextWrapWord

	// --- Helper Functions for Buttons ---

	// Hàm chọn thư mục
	browseFunc := func(target *widget.Entry, title string) {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			target.SetText(uri.Path())
		}, myWindow)
	}

	// Hàm set đường dẫn nhanh (Desktop/Downloads/Documents)
	setPathFunc := func(target *widget.Entry, folderName string) {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		target.SetText(filepath.Join(home, folderName))
	}

	// --- Layout Construction ---

	// Source Group
	sourceBrowseBtn := widget.NewButton("Browse...", func() { browseFunc(sourceEntry, "Select Project Directory") })
	sourceQuickBtns := container.NewHBox(
		widget.NewButton("Desktop", func() { setPathFunc(sourceEntry, "Desktop") }),
		widget.NewButton("Downloads", func() { setPathFunc(sourceEntry, "Downloads") }),
		widget.NewButton("Documents", func() { setPathFunc(sourceEntry, "Documents") }),
	)

	sourceGroup := container.NewVBox(
		sourceLabel,
		container.NewBorder(nil, nil, nil, sourceBrowseBtn, sourceEntry),
		sourceQuickBtns,
	)

	// Output Group
	outputBrowseBtn := widget.NewButton("Browse...", func() { browseFunc(outputEntry, "Select Output Directory") })
	outputQuickBtns := container.NewHBox(
		widget.NewButton("Desktop", func() { setPathFunc(outputEntry, "Desktop") }),
		widget.NewButton("Downloads", func() { setPathFunc(outputEntry, "Downloads") }),
		widget.NewButton("Documents", func() { setPathFunc(outputEntry, "Documents") }),
	)

	outputGroup := container.NewVBox(
		outputLabel,
		container.NewBorder(nil, nil, nil, outputBrowseBtn, outputEntry),
		outputQuickBtns,
	)

	// Extension Group
	extGroup := container.NewVBox(
		extLabel,
		extEntry,
	)

	// Process Button logic
	processBtn := widget.NewButton("Process Directory", nil) // Set action below to capture vars
	processBtn.Importance = widget.HighImportance

	processBtn.OnTapped = func() {
		sourceDir := sourceEntry.Text
		outputDir := outputEntry.Text
		extStr := extEntry.Text

		if sourceDir == "" {
			dialog.ShowError(fmt.Errorf("please select a project directory"), myWindow)
			return
		}
		if outputDir == "" {
			dialog.ShowError(fmt.Errorf("please select an output directory"), myWindow)
			return
		}

		processBtn.Disable()
		statusLabel.SetText("Processing... Please wait.")

		// Parse extensions
		extensions := make(map[string]struct{})
		allFiles := false
		trimmedExtStr := strings.TrimSpace(extStr)
		if trimmedExtStr == "" {
			allFiles = true
		} else {
			parts := strings.Split(trimmedExtStr, ",")
			for _, part := range parts {
				ext := strings.TrimSpace(part)
				ext = strings.TrimPrefix(ext, ".") // remove dot if user typed it
				if ext != "" {
					extensions[ext] = struct{}{}
				}
			}
		}

		config := Config{
			SourceDir:  sourceDir,
			OutputDir:  outputDir,
			Extensions: extensions,
			AllFiles:   allFiles,
			UpdateStatus: func(msg string) {
				// Fyne UI update must be thread-safe, usually handled auto but good practice
				statusLabel.SetText(msg)
			},
		}

		// Run in Goroutine
		go func() {
			defer processBtn.Enable()
			err := ProcessProject(config)
			if err != nil {
				statusLabel.SetText("Error: " + err.Error())
				dialog.ShowError(err, myWindow)
			}
			// Success message is handled inside ProcessProject via callback,
			// but we can ensure the final state here if needed.
		}()
	}

	// --- Main Layout Assembly ---

	// GroupBox style using Cards
	sourceCard := widget.NewCard("Input", "", sourceGroup)
	outputCard := widget.NewCard("Output", "", outputGroup)
	extCard := widget.NewCard("Configuration", "", extGroup)

	content := container.NewVBox(
		sourceCard,
		outputCard,
		extCard,
		layout.NewSpacer(),
		processBtn,
		statusLabel,
	)

	// Add padding
	paddedContent := container.NewPadded(content)

	myWindow.SetContent(paddedContent)
	myWindow.ShowAndRun()
}
