package main

import (
	"drive-tree/internal/authentication"
	"drive-tree/internal/datamng"
	guimng "drive-tree/internal/gui"
	"drive-tree/internal/tree"
	"drive-tree/internal/web"
	"flag"
	"fmt"
	"log"
)

func main() {

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
		return
	}

	if args[0] == "scraper" {
		scraper()
	} else if args[0] == "web" {
		viewer()
	} else if args[0] == "gui" {
		gui()
	} else {
		usage()
	}

}

func usage() {

	fmt.Println("Usage:")
	fmt.Println("\tdrive-tree scraper --> parse all your files")
	fmt.Println("\tdrive-tree web --> view your files on web after scraper")

}

func scraper() {

	srv := authentication.Authentication()

	//startNodeId := "1NesnNegi27dNuYL2E92oWav0ZShCIimo"
	myDriveStructure := tree.Run(srv)

	err := datamng.SaveData(myDriveStructure)

	if err != nil {
		return
	}
	log.Println("\n\n\nAll files were successfully parsed...")
	log.Println("}Use 'drive-tree web' to view your folder size on web")
}

func viewer() {

	myDriveStructure, startNodeId := datamng.LoadData()

	if myDriveStructure != nil {
		// Create web server
		web.Run(myDriveStructure, startNodeId)
	}

}

func gui() {

	myDriveStructure, startNodeId := datamng.LoadData()

	if myDriveStructure != nil {
		// Create web server
		guimng.Run(myDriveStructure, startNodeId)
	}

}
