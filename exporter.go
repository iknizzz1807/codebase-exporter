package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Config chứa các cấu hình cho việc xuất codebase
type Config struct {
	SourceDir    string
	OutputDir    string
	Extensions   map[string]struct{} // Sử dụng map rỗng làm set để tra cứu nhanh hơn
	AllFiles     bool
	UpdateStatus func(string) // Callback để cập nhật trạng thái trên UI
}

// Các thư mục cần bỏ qua
var skipDirs = map[string]struct{}{
	".git": {}, ".svn": {}, ".hg": {}, ".bzr": {},
	"node_modules": {}, "__pycache__": {}, ".pytest_cache": {},
	".mypy_cache": {}, ".tox": {}, ".coverage": {}, ".nyc_output": {},
	"coverage": {}, ".idea": {}, ".vscode": {}, ".vs": {},
	"bin": {}, "obj": {}, "build": {}, "dist": {},
	".gradle": {}, "target": {}, ".next": {}, ".nuxt": {},
	"out": {}, ".cache": {}, ".tmp": {}, "tmp": {},
	"temp": {}, "venv": {}, "env": {},
}

// Các file liên quan đến Docker luôn được bao gồm
var importantDockerFiles = map[string]struct{}{
	"dockerfile": {}, "dockerfile.dev": {}, "dockerfile.prod": {},
	"dockerfile.test": {}, "docker-compose.yml": {}, "docker-compose.yaml": {},
	"docker-compose.dev.yml": {}, "docker-compose.prod.yml": {}, "docker-compose.test.yml": {},
	"docker-compose.override.yml": {},
}

// Cấu trúc JSON của một cell trong Jupyter Notebook
type JupyterCell struct {
	CellType string   `json:"cell_type"`
	Source   []string `json:"source"`
}

// Cấu trúc JSON của một file Jupyter Notebook
type JupyterNotebook struct {
	Cells []JupyterCell `json:"cells"`
}

// ProcessProject là hàm chính thực thi việc xuất mã nguồn
func ProcessProject(cfg Config) error {
	outputPath := filepath.Join(cfg.OutputDir, "src.txt")
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("không thể tạo file output: %v", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// Ghi các thông tin đầu file
	writeHeaders(writer, cfg)

	// Ghi cấu trúc thư mục
	writer.WriteString("--- Directory Structure ---\n")
	cfg.UpdateStatus("Building directory tree...")
	buildTree(writer, cfg.SourceDir, "", cfg)
	writer.WriteString("\n")

	// Ghi nội dung chi tiết các file
	writer.WriteString("--- Source Code Details ---\n")
	cfg.UpdateStatus("Exporting file contents...")
	err = filepath.WalkDir(cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Bỏ qua các thư mục không cần thiết
		if d.IsDir() {
			if _, shouldSkip := skipDirs[d.Name()]; shouldSkip {
				return filepath.SkipDir
			}
			return nil
		}

		// Lọc file
		if d.Type().IsRegular() && shouldReadFile(path, cfg) {
			cfg.UpdateStatus(fmt.Sprintf("Processing: %s", path))
			dumpFile(writer, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("lỗi khi duyệt thư mục: %v", err)
	}

	cfg.UpdateStatus(fmt.Sprintf("Hoàn thành! Xuất file thành công tới: %s", outputPath))
	return nil
}

// writeHeaders ghi các thông tin giới thiệu vào đầu file output
func writeHeaders(writer *bufio.Writer, cfg Config) {
	writer.WriteString("--- Project Overview ---\n\n\n")
	writer.WriteString("--- Notes ---\nThis is the complete project code. Please read and understand thoroughly.\n\n")

	writer.WriteString("--- File Types Included ---\n")
	if cfg.AllFiles {
		writer.WriteString("All files\n")
	} else {
		exts := make([]string, 0, len(cfg.Extensions))
		for ext := range cfg.Extensions {
			exts = append(exts, "."+ext)
		}
		sort.Strings(exts)
		for _, ext := range exts {
			writer.WriteString(ext + "\n")
		}
	}
	writer.WriteString("\n--- Important Files Always Included ---\n")
	writer.WriteString("Dockerfile, docker-compose.yml/.yaml files (regardless of extension filter)\n\n")

	writer.WriteString("--- Excluded Directories ---\n")
	excluded := make([]string, 0, len(skipDirs))
	for dir := range skipDirs {
		excluded = append(excluded, dir)
	}
	sort.Strings(excluded)
	writer.WriteString(strings.Join(excluded, ", ") + "\n\n")
}

// buildTree tạo và ghi cấu trúc cây thư mục
func buildTree(writer *bufio.Writer, root, prefix string, cfg Config) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	filteredEntries := []fs.DirEntry{}
	for _, entry := range entries {
		if _, shouldSkip := skipDirs[entry.Name()]; !shouldSkip {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Logic hiển thị ... nếu có quá nhiều file
	tooManyFiles := len(filteredEntries) > 50
	displayCount := len(filteredEntries)
	if tooManyFiles {
		displayCount = 5
	}

	for i, entry := range filteredEntries[:displayCount] {
		connector := "├── "
		newPrefix := prefix + "│   "
		if i == len(filteredEntries)-1 || (tooManyFiles && i == displayCount-1) {
			connector = "└── "
			newPrefix = prefix + "    "
		}

		writer.WriteString(prefix + connector + entry.Name() + "\n")

		if entry.IsDir() {
			buildTree(writer, filepath.Join(root, entry.Name()), newPrefix, cfg)
		}
	}

	if tooManyFiles {
		writer.WriteString(prefix + "└── ... (and " + fmt.Sprintf("%d", len(filteredEntries)-5) + " more items)\n")
	}
}

// shouldReadFile kiểm tra xem file có nên được đọc và đưa vào output không
func shouldReadFile(path string, cfg Config) bool {
	// Luôn bao gồm các file Docker quan trọng
	if _, ok := importantDockerFiles[strings.ToLower(filepath.Base(path))]; ok {
		return true
	}

	if cfg.AllFiles {
		return true
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	_, found := cfg.Extensions[ext]
	return found
}

// dumpFile ghi nội dung của một file vào writer
func dumpFile(writer *bufio.Writer, path string) {
	// Đếm số dòng để bỏ qua file quá lớn
	lineCount, err := countLines(path)
	if err != nil || lineCount > 10000 {
		writer.WriteString(fmt.Sprintf("\n// File: %s (skipped - too large: %d lines or read error)\n", path, lineCount))
		return
	}

	// Xử lý đặc biệt cho Jupyter Notebook
	if filepath.Ext(path) == ".ipynb" {
		processJupyterNotebook(writer, path)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		writer.WriteString(fmt.Sprintf("\n// File: %s (could not be opened)\n", path))
		return
	}
	defer file.Close()

	writer.WriteString(fmt.Sprintf("\n// File: %s (%d lines)\n", path, lineCount))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		writer.WriteString(scanner.Text() + "\n")
	}
}

// processJupyterNotebook đọc file .ipynb, trích xuất và ghi các cell code/markdown
func processJupyterNotebook(writer *bufio.Writer, path string) {
	file, err := os.Open(path)
	if err != nil {
		writer.WriteString(fmt.Sprintf("\n// File: %s (Jupyter Notebook - could not be opened)\n", path))
		return
	}
	defer file.Close()

	writer.WriteString(fmt.Sprintf("\n// File: %s (Jupyter Notebook)\n", path))

	var notebook JupyterNotebook
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&notebook); err != nil {
		writer.WriteString(fmt.Sprintf("// Error parsing notebook: %v\n", err))
		return
	}

	writer.WriteString("\n// Notebook Content:\n")
	for i, cell := range notebook.Cells {
		writer.WriteString(fmt.Sprintf("\n// --- Cell %d (%s) ---\n", i+1, cell.CellType))
		for _, line := range cell.Source {
			writer.WriteString(line)
		}
		writer.WriteString("\n")
	}
}

// countLines đếm số dòng trong một file
func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	count := 0
	for {
		_, err := reader.ReadString('\n')
		count++
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
