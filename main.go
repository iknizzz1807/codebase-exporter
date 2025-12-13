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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Codebase Exporter Pro")
	myWindow.Resize(fyne.NewSize(900, 700))

	// State variables
	var excludedDirs []string
	var selectedFiles []string

	// --- UI Components ---

	// Source Directory
	sourceEntry := widget.NewEntry()
	sourceEntry.PlaceHolder = "Select project path..."
	sourceBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			sourceEntry.SetText(uri.Path())
		}, myWindow)
	})

	// Output Directory
	outputEntry := widget.NewEntry()
	outputEntry.PlaceHolder = "Select output path..."
	outputBrowseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			outputEntry.SetText(uri.Path())
		}, myWindow)
	})

	// Quick path buttons
	quickPathBtns := container.NewHBox(
		widget.NewButton("Desktop", func() {
			if home, err := os.UserHomeDir(); err == nil {
				outputEntry.SetText(filepath.Join(home, "Desktop"))
			}
		}),
		widget.NewButton("Downloads", func() {
			if home, err := os.UserHomeDir(); err == nil {
				outputEntry.SetText(filepath.Join(home, "Downloads"))
			}
		}),
	)

	// Extensions
	extEntry := widget.NewEntry()
	extEntry.SetText("go,cpp,h,txt,ipynb,py,js,ts,tsx,jsx,html,css,java,svelte,vue")
	allFilesCheck := widget.NewCheck("Include all file types", nil)

	// Export Mode
	exportModeGroup := widget.NewRadioGroup([]string{
		"Structure + Code (Default)",
		"Structure Only",
	}, nil)
	exportModeGroup.SetSelected("Structure + Code (Default)")
	exportModeGroup.Horizontal = false

	// Scan Mode
	scanModeGroup := widget.NewRadioGroup([]string{
		"Scan all subdirectories",
		"Select specific files only",
	}, nil)
	scanModeGroup.SetSelected("Scan all subdirectories")
	scanModeGroup.Horizontal = false

	// Code Output Mode
	codeOutputGroup := widget.NewRadioGroup([]string{
		"Full source code",
		"AST summary (experimental)",
	}, nil)
	codeOutputGroup.SetSelected("Full source code")
	codeOutputGroup.Horizontal = false

	// Excluded directories list - KHAI BÁO TRƯỚC
	var excludeList *widget.List
	excludeList = widget.NewList(
		func() int { return len(excludedDirs) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
				widget.NewLabel(""))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			btn := c.Objects[1].(*widget.Button)
			label.SetText(excludedDirs[id])
			btn.OnTapped = func() {
				excludedDirs = append(excludedDirs[:id], excludedDirs[id+1:]...)
				excludeList.Refresh()
			}
		},
	)

	excludeEntry := widget.NewEntry()
	excludeEntry.PlaceHolder = "Directory name to exclude..."
	addExcludeBtn := widget.NewButton("Add", func() {
		dir := strings.TrimSpace(excludeEntry.Text)
		if dir != "" {
			excludedDirs = append(excludedDirs, dir)
			excludeList.Refresh()
			excludeEntry.SetText("")
		}
	})

	// Selected files list - KHAI BÁO TRƯỚC
	var fileList *widget.List
	fileList = widget.NewList(
		func() int { return len(selectedFiles) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
				widget.NewLabel(""))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			btn := c.Objects[1].(*widget.Button)
			label.SetText(filepath.Base(selectedFiles[id]))
			btn.OnTapped = func() {
				selectedFiles = append(selectedFiles[:id], selectedFiles[id+1:]...)
				fileList.Refresh()
			}
		},
	)

	selectFilesBtn := widget.NewButton("Select Files", func() {
		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()
			selectedFiles = append(selectedFiles, uc.URI().Path())
			fileList.Refresh()
		}, myWindow)
	})
	clearFilesBtn := widget.NewButton("Clear All", func() {
		selectedFiles = []string{}
		fileList.Refresh()
	})

	// Status
	statusLabel := widget.NewLabel("Ready.")
	statusLabel.Wrapping = fyne.TextWrapWord
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	// Process button
	processBtn := widget.NewButton("Export Codebase", nil)
	processBtn.Importance = widget.HighImportance
	processBtn.OnTapped = func() {
		sourceDir := sourceEntry.Text
		outputDir := outputEntry.Text

		if sourceDir == "" {
			dialog.ShowError(fmt.Errorf("please select a project directory"), myWindow)
			return
		}
		if outputDir == "" {
			dialog.ShowError(fmt.Errorf("please select an output directory"), myWindow)
			return
		}

		// Validate scan mode
		if scanModeGroup.Selected == "Select specific files only" && len(selectedFiles) == 0 {
			dialog.ShowError(fmt.Errorf("please select at least one file"), myWindow)
			return
		}

		processBtn.Disable()
		progressBar.Show()
		statusLabel.SetText("Processing... Please wait.")

		// Parse extensions
		extensions := make(map[string]struct{})
		allFiles := allFilesCheck.Checked
		if !allFiles {
			parts := strings.Split(extEntry.Text, ",")
			for _, part := range parts {
				ext := strings.TrimSpace(part)
				ext = strings.TrimPrefix(ext, ".")
				if ext != "" {
					extensions[ext] = struct{}{}
				}
			}
		}

		// Merge excluded dirs
		customExcludeDirs := make(map[string]struct{})
		for dir := range skipDirs {
			customExcludeDirs[dir] = struct{}{}
		}
		for _, dir := range excludedDirs {
			customExcludeDirs[dir] = struct{}{}
		}

		config := Config{
			SourceDir:     sourceDir,
			OutputDir:     outputDir,
			Extensions:    extensions,
			AllFiles:      allFiles,
			ExcludedDirs:  customExcludeDirs,
			StructureOnly: exportModeGroup.Selected == "Structure Only",
			SpecificFiles: selectedFiles,
			ScanAllDirs:   scanModeGroup.Selected == "Scan all subdirectories",
			UseASTMode:    codeOutputGroup.Selected == "AST summary (experimental)",
			UpdateStatus: func(msg string) {
				statusLabel.SetText(msg)
			},
		}

		go func() {
			defer func() {
				processBtn.Enable()
				progressBar.Hide()
			}()
			err := ProcessProject(config)
			if err != nil {
				statusLabel.SetText("Error: " + err.Error())
				dialog.ShowError(err, myWindow)
			}
		}()
	}

	// Layout
	sourceCard := widget.NewCard("Source Directory", "", container.NewVBox(
		container.NewBorder(nil, nil, nil, sourceBrowseBtn, sourceEntry),
	))

	outputCard := widget.NewCard("Output Directory", "", container.NewVBox(
		container.NewBorder(nil, nil, nil, outputBrowseBtn, outputEntry),
		quickPathBtns,
	))

	configCard := widget.NewCard("Configuration", "", container.NewVBox(
		widget.NewLabel("File Extensions (comma separated):"),
		extEntry,
		allFilesCheck,
		widget.NewSeparator(),
		widget.NewLabel("Export Mode:"),
		exportModeGroup,
		widget.NewSeparator(),
		widget.NewLabel("Scan Mode:"),
		scanModeGroup,
		widget.NewSeparator(),
		widget.NewLabel("Code Output Mode:"),
		codeOutputGroup,
	))

	excludeCard := widget.NewCard("Excluded Directories", "Custom directories to skip", container.NewVBox(
		container.NewBorder(nil, nil, nil, addExcludeBtn, excludeEntry),
		container.NewScroll(excludeList),
	))

	filesCard := widget.NewCard("Specific Files", "Only used in 'Select specific files' mode", container.NewVBox(
		container.NewHBox(selectFilesBtn, clearFilesBtn),
		container.NewScroll(fileList),
	))

	leftColumn := container.NewVBox(sourceCard, outputCard, configCard)
	rightColumn := container.NewVBox(excludeCard, filesCard)

	mainContent := container.NewHSplit(leftColumn, rightColumn)
	mainContent.SetOffset(0.5)

	content := container.NewBorder(
		nil,
		container.NewVBox(widget.NewSeparator(), progressBar, statusLabel, processBtn),
		nil, nil,
		mainContent,
	)

	myWindow.SetContent(container.NewPadded(content))
	myWindow.ShowAndRun()
}
