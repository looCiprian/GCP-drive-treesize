package tree

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"google.golang.org/api/drive/v2"
)

// File structure
type MyDrive struct {
	Name          string   `json:"name"`
	Id            string   `json:"id"`
	Size          int64    `json:"size"`
	IsDir         bool     `json:"isdir"`
	Child         []string `json:"child"`
	IsRoot        bool     `json:"isroot"`
	Parent        string   `json:"parent"`
	NumberOfFiles int      `json:"numberoffiles"`
	HumanSize     string   `json:"humansize"`
	MimeType      string   `json:"mimetype"`
	Link          string   `json:"link"`
}

var myDriveTree = make(map[string]*MyDrive) // Drive tree
var filesList []*drive.File                 // Google file list info
var rootNodeId string                       // Root node id

func popFromFileList() *drive.File {
	var file *drive.File
	file, filesList = filesList[0], filesList[1:]
	return file
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime)

}

func Run(d *drive.Service) map[string]*MyDrive {

	log.Println("Starting get all your files...")

	// Get all files that are owned by you using Google Drive API
	allFiles(d)

	//log.SetFlags(log.Ldate | log.Ltime)
	log.Println("Creating the tree...")

	// Migrate filesList to myDriveTree
	for len(filesList) > 0 {
		fileInfo := popFromFileList()
		isdir := false
		if fileInfo.MimeType == "application/vnd.google-apps.folder" {
			isdir = true
		}

		parent := ""
		if len(fileInfo.Parents) > 0 {
			parent = fileInfo.Parents[0].Id
		}

		myDriveTree[fileInfo.Id] = &MyDrive{
			Name:      fileInfo.Title,
			Id:        fileInfo.Id,
			Size:      fileInfo.FileSize,
			IsDir:     isdir,
			IsRoot:    false,
			Parent:    parent,
			MimeType:  fileInfo.MimeType,
			HumanSize: getHumanBytes(uint64(fileInfo.FileSize)),
			Link:      fileInfo.AlternateLink}

	}

	// Get the root node info
	fileInfo := getRootInfo(d)

	// Create root node
	myDriveTree[fileInfo.Id] = &MyDrive{
		Name:      fileInfo.Title,
		Id:        fileInfo.Id,
		Size:      fileInfo.FileSize,
		IsDir:     fileInfo.MimeType == "application/vnd.google-apps.folder",
		Parent:    "",
		IsRoot:    true,
		MimeType:  fileInfo.MimeType,
		HumanSize: getHumanBytes(uint64(fileInfo.FileSize)),
		Link:      fileInfo.AlternateLink}
	rootNodeId = fileInfo.Id

	// Ingesting the files and creates the tree
	fileIngestor(d)

	return myDriveTree
}

// Get all files from Google Drive
func allFiles(d *drive.Service) error {

	pageToken := ""
	for {
		// Get all files in a specifc nodeId
		q := d.Files.List().IncludeItemsFromAllDrives(false).SupportsAllDrives(false).Q("'me' in owners and trashed = false").MaxResults(1000)
		// If we have a pageToken set, apply it to the query
		if pageToken != "" {
			q = q.PageToken(pageToken)
		}

		// Get the files
		var r *drive.FileList
		for { // Loop to manage errors
			var err error
			r, err = q.Do()
			if err != nil {
				fmt.Printf("An error occurred: %v\n", err)
				fmt.Printf("Retrying in 2 seconds...\n")
				time.Sleep(time.Second * 2)
			} else {
				//log.SetFlags(log.Ldate | log.Ltime)
				log.Println("Getting data...")
				break
			}
		}

		// Add the files to the list
		filesList = append(filesList, r.Items...)

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return nil
}

// Create the tree
func fileIngestor(srv *drive.Service) {

	for _, fileInfo := range myDriveTree {
		if fileInfo.MimeType == "application/vnd.google-apps.shortcut" {
			// Avoid loop
			continue
		} else { // File or folder

			// Updating the parent node
			updateParentNode(fileInfo)

		}
	}

	// Attaching all files that are owned by you but in a shared directory to root node (they don't have a parent)
	for _, fileInfo := range myDriveTree {
		if fileInfo.Parent == "" && !fileInfo.IsRoot {
			log.Println("Parent node not found attaching to root file: " + fileInfo.Id + " file name: " + fileInfo.Name)
			rootNode := myDriveTree[rootNodeId]
			fileInfo.Parent = rootNodeId
			rootNode.Size += fileInfo.Size
			rootNode.HumanSize = getHumanBytes(uint64(rootNode.Size))
			if fileInfo.IsDir {
				rootNode.NumberOfFiles += fileInfo.NumberOfFiles
			} else {
				rootNode.NumberOfFiles++
			}

			rootNode.Child = append(rootNode.Child, fileInfo.Id)
		}
	}
}

// Update parent node, by updating child list, the size, and number of files
func updateParentNode(fileInfo *MyDrive) {

	// Updating the parent node
	if parentNode, ok := myDriveTree[fileInfo.Parent]; ok {

		// Updating child list
		parentNode.Child = append(parentNode.Child, fileInfo.Id)
		// Update tree size
		updateParentInfoRecursive(fileInfo.Id, fileInfo.Size, fileInfo.IsDir)
	} else { // In this case we own a file that is inside a shared directory owned by someone else, the file will be attached to root node after all processing
		if !fileInfo.IsRoot {
			fileInfo.Parent = ""
			log.Println("Parent node not found, probably owned by you but in a shared dir: " + fileInfo.Name)
		}
	}
}

func updateParentInfoRecursive(nodeId string, size int64, isDir bool) {
	if node, ok := myDriveTree[nodeId]; ok {
		if node.Parent == "" {
			return
		}
		if parentNode, okParent := myDriveTree[node.Parent]; okParent {

			// Update number of files and size only if the child is a file
			if !isDir {
				parentNode.Size += size
				parentNode.NumberOfFiles++
			}
			parentNode.HumanSize = getHumanBytes(uint64(parentNode.Size))
			updateParentInfoRecursive(parentNode.Id, size, isDir)
		} else { // In this case we own a file that is inside a shared directory owned by someone else, the file will be attached to root node
			return
		}
	}
}

func getRootInfo(srv *drive.Service) *drive.File {
	//1NesnNegi27dNuYL2E92oWav0ZShCIimo

	var fileRootInfo *drive.File

	for {
		fileInfo, err := srv.Files.Get("root").Do()

		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
			fmt.Printf("Retrying in 2 seconds...\n")
			time.Sleep(time.Second * 2)
		} else {
			fileRootInfo = fileInfo
			break
		}
	}

	return fileRootInfo
}

func getHumanBytes(bytes uint64) string {

	return humanize.IBytes(bytes)
}
