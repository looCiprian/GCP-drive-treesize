package tree

import (
	"sort"
	"strings"
)

type MyPath struct {
	Id   string
	Name string
}

// Get path from node to root return MyPath struct
func GetCurrentPath(myDriveTree map[string]*MyDrive, startId string) []MyPath {

	currentId := startId
	currentPath := []MyPath{}

	for {
		newNode := MyPath{Id: currentId, Name: myDriveTree[currentId].Name}
		currentPath = append([]MyPath{newNode}, currentPath...)
		if myDriveTree[currentId].IsRoot {
			break
		}
		currentId = myDriveTree[currentId].Parent
	}

	return currentPath
}

func GetChildList(myDriveTree map[string]*MyDrive, id string) []MyDrive {
	// Get child list of the selected node
	currentNode := myDriveTree[id]
	childList := []MyDrive{}

	for _, child := range currentNode.Child {
		childList = append(childList, *myDriveTree[child])
	}
	// Sort by size
	sort.Slice(childList, func(i, j int) bool { return childList[i].Size > childList[j].Size })

	return childList
}

// Get path from node to root return string
func GetCurrentPathSting(myDriveTree map[string]*MyDrive, startId string) string {

	currentId := startId
	currentPath := []string{}

	for {
		newNode := []string{myDriveTree[currentId].Name}
		currentPath = append(newNode, currentPath...)
		if myDriveTree[currentId].IsRoot {
			break
		}
		currentId = myDriveTree[currentId].Parent
	}

	return "/" + strings.Join(currentPath, "/")
}
