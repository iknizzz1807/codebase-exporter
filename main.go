package main

import (
	"log"
	"strings"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

// MainWindow là cấu trúc của cửa sổ chính
type MainWindow struct {
	*walk.MainWindow
	sourceDirEdit  *walk.LineEdit
	outputDirEdit  *walk.LineEdit
	extensionsEdit *walk.LineEdit
	processButton  *walk.PushButton
	statusBar      *walk.StatusBarItem
}

func main() {
	mw := new(MainWindow)

	err := declarative.MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Codebase Exporter",
		Size:     declarative.Size{Width: 570, Height: 280},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			// GroupBox cho thư mục dự án
			declarative.GroupBox{
				Title:  "Project Directory",
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.LineEdit{AssignTo: &mw.sourceDirEdit, ReadOnly: true},
					declarative.PushButton{
						Text: "Browse...",
						OnClicked: func() {
							mw.browseForFolder(mw.sourceDirEdit, "Select Project Directory")
						},
					},
				},
			},
			// GroupBox cho thư mục output
			declarative.GroupBox{
				Title:  "Output Directory",
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.LineEdit{AssignTo: &mw.outputDirEdit, ReadOnly: true},
					declarative.PushButton{
						Text: "Browse...",
						OnClicked: func() {
							mw.browseForFolder(mw.outputDirEdit, "Select Output Directory")
						},
					},
				},
			},
			// GroupBox cho các phần mở rộng
			declarative.GroupBox{
				Title:  "File Extensions (comma separated, leave empty for all)",
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.LineEdit{
						AssignTo: &mw.extensionsEdit,
						Text:     "go,cpp,h,txt,ipynb,py", // Giá trị mặc định
					},
				},
			},
			// Không gian đệm và nút Process
			declarative.VSpacer{},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{
						AssignTo:  &mw.processButton,
						Text:      "Process Directory",
						OnClicked: mw.processDirectory,
					},
					declarative.HSpacer{},
				},
			},
		},
		// Thanh trạng thái
		StatusBarItems: []declarative.StatusBarItem{
			{
				AssignTo: &mw.statusBar,
				Text:     "Ready. Select a directory and output path.",
			},
		},
	}.Create()

	if err != nil {
		log.Fatal(err)
	}

	mw.Run()
}

// browseForFolder mở hộp thoại chọn thư mục và gán đường dẫn vào LineEdit
func (mw *MainWindow) browseForFolder(target *walk.LineEdit, title string) {
	dlg := new(walk.FileDialog)
	dlg.Title = title
	dlg.Filter = "Folders|"

	if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
		mw.updateStatus("Lỗi mở hộp thoại: " + err.Error())
		return
	} else if !ok {
		return // Người dùng hủy
	}
	target.SetText(dlg.FilePath)
}

// updateStatus cập nhật text trên thanh trạng thái một cách an toàn
func (mw *MainWindow) updateStatus(text string) {
	// Sử dụng Synchronize để cập nhật UI từ một goroutine khác một cách an toàn
	mw.Synchronize(func() {
		mw.statusBar.SetText(text)
	})
}

// processDirectory là hàm xử lý khi nút "Process" được nhấn
func (mw *MainWindow) processDirectory() {
	sourceDir := mw.sourceDirEdit.Text()
	outputDir := mw.outputDirEdit.Text()
	extStr := mw.extensionsEdit.Text()

	if sourceDir == "" {
		walk.MsgBox(mw, "Error", "Please select a project directory.", walk.MsgBoxIconError)
		return
	}
	if outputDir == "" {
		walk.MsgBox(mw, "Error", "Please select an output directory.", walk.MsgBoxIconError)
		return
	}

	// Vô hiệu hóa nút để tránh nhấn nhiều lần
	mw.processButton.SetEnabled(false)
	defer mw.processButton.SetEnabled(true)

	// Chuẩn bị config
	extensions := make(map[string]struct{})
	allFiles := false
	trimmedExtStr := strings.TrimSpace(extStr)
	if trimmedExtStr == "" {
		allFiles = true
	} else {
		parts := strings.Split(trimmedExtStr, ",")
		for _, part := range parts {
			ext := strings.TrimSpace(part)
			ext = strings.TrimPrefix(ext, ".")
			if ext != "" {
				extensions[ext] = struct{}{}
			}
		}
	}

	config := Config{
		SourceDir:    sourceDir,
		OutputDir:    outputDir,
		Extensions:   extensions,
		AllFiles:     allFiles,
		UpdateStatus: mw.updateStatus, // Truyền hàm callback
	}

	// Chạy tác vụ nặng trong một goroutine riêng để không làm treo UI
	go func() {
		mw.updateStatus("Processing... Please wait.")
		err := ProcessProject(config)
		if err != nil {
			// Hiển thị lỗi trên cả status bar và message box
			errMsg := "Error: " + err.Error()
			mw.updateStatus(errMsg)
			walk.MsgBox(mw, "Processing Error", errMsg, walk.MsgBoxIconError)
		}
		// Hàm ProcessProject sẽ tự cập nhật trạng thái thành công
	}()
}
