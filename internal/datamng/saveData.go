package datamng

import (
	"drive-tree/internal/tree"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func SaveData(myDriveStructure map[string]*tree.MyDrive) error {
	// Save tree in json file
	marshalTree, err := json.Marshal(myDriveStructure)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Clean file
	_ = os.Remove("tree.json")
	// Create file
	f, errf := os.OpenFile("tree.json", os.O_RDWR|os.O_CREATE, 0664)

	if errf != nil {
		fmt.Println(errf)
		return errf
	}
	// Write to file
	f.Write(marshalTree)
	f.Close()

	return nil
}

func LoadData() (map[string]*tree.MyDrive, string) {
	// Open file
	f, err := ioutil.ReadFile("tree.json")

	if err != nil {
		log.Fatalln(err)
		return nil, ""
	}

	// Unmarshal json and load tree
	var myDriveStructure map[string]*tree.MyDrive
	errU := json.Unmarshal(f, &myDriveStructure)
	if errU != nil {
		log.Fatalln(errU)
		return nil, ""
	}

	// Get root id node
	startNodeId := ""
	for _, v := range myDriveStructure {
		if v.IsRoot {
			startNodeId = v.Id
		}
	}

	return myDriveStructure, startNodeId
}
