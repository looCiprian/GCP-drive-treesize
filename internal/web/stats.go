package web

import (
	"path"
)

type stats struct {
	MimeTypeFreq      map[string]int
	FileExtensionFreq map[string]int
}

func getStats() stats {

	var mimeTypeFreq stats
	mimeTypeFreq.MimeTypeFreq = make(map[string]int)
	mimeTypeFreq.FileExtensionFreq = make(map[string]int)

	for _, v := range myDriveTree {

		mimeTypeFreq.MimeTypeFreq[v.MimeType]++

		ext := getFileExtension(v.Name)
		if ext != "" {
			mimeTypeFreq.FileExtensionFreq[ext]++
		}
	}
	return mimeTypeFreq
}

func getFileExtension(fileName string) string {

	return path.Ext(fileName)

}
