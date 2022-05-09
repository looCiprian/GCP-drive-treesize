package main

import (
	"drive-tree/internal/authentication"
	"drive-tree/internal/tree"
	"drive-tree/internal/web"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	// Save tree in json file
	marshalTree, err := json.Marshal(myDriveStructure)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Clean file
	_ = os.Remove("tree.json")
	// Create file
	f, errf := os.OpenFile("tree.json", os.O_RDWR|os.O_CREATE, 0664)

	if errf != nil {
		fmt.Println(errf)
		return
	}
	// Write to file
	f.Write(marshalTree)
	f.Close()

	log.Println("\n\n\nAll files were successfully parsed...")
	log.Println("Use 'drive-tree web' to view your folder size on web")

}

func viewer() {

	// Open file
	f, err := ioutil.ReadFile("tree.json")

	if err != nil {
		log.Fatalln(err)
		return
	}

	// Unmarshal json and load tree
	var myDriveStructure map[string]*tree.MyDrive
	errU := json.Unmarshal(f, &myDriveStructure)
	if errU != nil {
		log.Fatalln(errU)
		return
	}

	// Get root id node
	startNodeId := ""
	for _, v := range myDriveStructure {
		if v.IsRoot {
			startNodeId = v.Id
		}
	}

	// Create web server
	web.Run(myDriveStructure, startNodeId)

}
