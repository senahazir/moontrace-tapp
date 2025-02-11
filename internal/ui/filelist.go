// internal/ui/filelist.go
package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rivo/tview"
)

func (v *Views) UpdateFileList(list *tview.List, dir string) {
	list.Clear()
	dir = filepath.Clean(dir)
	v.Logger.Printf("Updating file list for directory: %s", dir)

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
			displayName = fmt.Sprintf("%s%s (%s) *", prefix, fileName, sizeStr)
		}

		list.AddItem(displayName, "", 0, nil)
	}
}
