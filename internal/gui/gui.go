package gui

import (
	"drive-tree/internal/tree"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Run(myTree map[string]*tree.MyDrive, startId string) {

	driveTreeApp := app.New()
	driveTreeWindow := driveTreeApp.NewWindow("Drive Tree")

	currentNodeId := startId
	//currentNode := myTree[startId]
	childList := tree.GetChildList(myTree, currentNodeId)

	// Create file list from root child (left panel)
	startList := widget.NewList(
		func() int {
			return len(childList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("childList")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {

			itemLabelText := "[" + childList[i].HumanSize + "] " + childList[i].Name
			o.(*widget.Label).SetText(itemLabelText)
		})

	// Create file information widget from selected item (right panel)
	widgetInfo := widget.NewLabel("Select a file to get info or a directory to explore it")
	// Set up containers for list of file and file/folder info
	containerData := container.NewHSplit(
		startList,
		widgetInfo,
	)

	// Set up top container for path
	pathInfo := widget.NewLabel(tree.GetCurrentPathSting(myTree, currentNodeId))

	// Once clicked got back to the parent node and update list, path and information
	backButton := widget.NewButton("Back", func() {
		if myTree[currentNodeId].IsRoot {
			return
		}
		currentNodeId = myTree[currentNodeId].Parent
		currentNode := myTree[currentNodeId]
		widgetUpdate(*currentNode, widgetInfo)
		childList = tree.GetChildList(myTree, currentNodeId)
		startList.Refresh()
		pathInfo.SetText(tree.GetCurrentPathSting(myTree, currentNodeId))
		pathInfo.Refresh()
		startList.UnselectAll()
	})

	containerTop := container.NewHBox(pathInfo, backButton)

	containerFinal := container.NewVSplit(containerTop, containerData)
	// Once a file/folder is selected list, information and path are updated
	startList.OnSelected = func(id widget.ListItemID) {
		child := childList[id]
		if child.IsDir {
			currentNodeId = child.Id
			widgetUpdate(child, widgetInfo)
			childList = tree.GetChildList(myTree, child.Id)
			startList.Refresh()
			pathInfo.SetText(tree.GetCurrentPathSting(myTree, currentNodeId))
			pathInfo.Refresh()
			startList.UnselectAll()
		} else {
			widgetUpdate(child, widgetInfo)
			startList.UnselectAll()
		}
	}

	// Draw the window
	driveTreeWindow.SetContent(containerFinal)
	// Resize the window to the size of the content
	driveTreeWindow.Resize(fyne.NewSize(1000, 1000))
	driveTreeWindow.ShowAndRun()
}

func widgetUpdate(node tree.MyDrive, widgetObject *widget.Label) {

	if node.IsDir {
		widgetInfoLabel := "File name: " + node.Name + "\n" + "File size: " + node.HumanSize + "\n" + "File id: " + node.Link + "\n" + "File inside: " + strconv.Itoa(node.NumberOfFiles)
		widgetObject.SetText(widgetInfoLabel)
	} else {
		widgetInfoLabel := "File name: " + node.Name + "\n" + "File size: " + node.HumanSize + "\n" + "File id: " + node.Link
		widgetObject.SetText(widgetInfoLabel)
	}

}
