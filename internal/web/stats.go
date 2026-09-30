package web

import (
	"drive-tree/internal/tree"
	"sort"
	"strings"
)

// Number of entries shown in each statistics table
const statsLimit = 25

type typeStat struct {
	Label    string
	MimeType string
	Count    int
	Size     int64
}

type fileStat struct {
	tree.MyDrive
	Path string
}

type stats struct {
	TotalSize    int64
	TotalFiles   int
	TotalFolders int
	ByType       []typeStat // Sorted by size, then count
	OtherTypes   int        // Types not shown in ByType
	Largest      []fileStat // Largest files, biggest first
}

func getStats(myDriveTree map[string]*tree.MyDrive) stats {

	var s stats
	byType := make(map[string]*typeStat)
	var files []*tree.MyDrive

	for _, v := range myDriveTree {
		if v.IsRoot {
			s.TotalSize = v.Size
			continue
		}
		if v.IsDir {
			s.TotalFolders++
			continue
		}
		s.TotalFiles++
		files = append(files, v)

		t, ok := byType[v.MimeType]
		if !ok {
			t = &typeStat{Label: mimeLabel(v.MimeType), MimeType: v.MimeType}
			byType[v.MimeType] = t
		}
		t.Count++
		t.Size += v.Size
	}

	for _, t := range byType {
		s.ByType = append(s.ByType, *t)
	}
	sort.Slice(s.ByType, func(i, j int) bool {
		if s.ByType[i].Size != s.ByType[j].Size {
			return s.ByType[i].Size > s.ByType[j].Size
		}
		return s.ByType[i].Count > s.ByType[j].Count
	})
	if len(s.ByType) > statsLimit {
		s.OtherTypes = len(s.ByType) - statsLimit
		s.ByType = s.ByType[:statsLimit]
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Size > files[j].Size })
	for _, f := range files {
		if len(s.Largest) == statsLimit || f.Size == 0 {
			break
		}
		s.Largest = append(s.Largest, fileStat{*f, tree.GetCurrentPathSting(myDriveTree, f.Parent)})
	}

	return s
}

var mimeLabels = map[string]string{
	"application/vnd.google-apps.document":     "Google Docs",
	"application/vnd.google-apps.spreadsheet":  "Google Sheets",
	"application/vnd.google-apps.presentation": "Google Slides",
	"application/vnd.google-apps.form":         "Google Forms",
	"application/vnd.google-apps.drawing":      "Google Drawings",
	"application/vnd.google-apps.shortcut":     "Shortcuts",
	"application/pdf":                          "PDF",
	"application/zip":                          "ZIP archive",
}

// Human friendly name of a mime type
func mimeLabel(mimeType string) string {
	if label, ok := mimeLabels[mimeType]; ok {
		return label
	}
	if mimeType == "" {
		return "Unknown"
	}
	return mimeType
}

// Broad category of a file, used to pick its icon
func fileKind(mimeType string, isDir bool) string {
	switch {
	case isDir:
		return "folder"
	case strings.HasPrefix(mimeType, "image/"), mimeType == "application/vnd.google-apps.drawing":
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case mimeType == "application/pdf":
		return "pdf"
	case strings.Contains(mimeType, "spreadsheet"), strings.Contains(mimeType, "excel"), mimeType == "text/csv":
		return "sheet"
	case strings.Contains(mimeType, "presentation"), strings.Contains(mimeType, "powerpoint"):
		return "slides"
	case strings.Contains(mimeType, "document"), strings.Contains(mimeType, "word"), strings.HasPrefix(mimeType, "text/"):
		return "doc"
	case strings.Contains(mimeType, "zip"), strings.Contains(mimeType, "compressed"), strings.Contains(mimeType, "tar"):
		return "archive"
	}
	return "file"
}
