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
		node, ok := myDriveTree[currentId]
		if !ok { // Broken parent chain, stop here
			break
		}
		newNode := MyPath{Id: currentId, Name: node.Name}
		currentPath = append([]MyPath{newNode}, currentPath...)
		if node.IsRoot {
			break
		}
		currentId = node.Parent
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
		node, ok := myDriveTree[currentId]
		if !ok { // Broken parent chain, stop here
			break
		}
		newNode := []string{node.Name}
		currentPath = append(newNode, currentPath...)
		if node.IsRoot {
			break
		}
		currentId = node.Parent
	}

	return "/" + strings.Join(currentPath, "/")
}
