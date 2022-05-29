package web

import (
	"drive-tree/internal/tree"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/pkg/browser"

	"github.com/gorilla/mux"
)

var myDriveTree = make(map[string]*tree.MyDrive)
var myRouter *mux.Router
var startNodeId string
var statistics stats

func handleRequests() {

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/node/{id}", returnNodeInfo)
	myRouter.HandleFunc("/node/", func(rw http.ResponseWriter, r *http.Request) {
		urlRedirect := fmt.Sprintf("/node/%s", startNodeId)
		http.Redirect(rw, r, urlRedirect, http.StatusMovedPermanently)
	})
	myRouter.HandleFunc("/stats", returnStats)

}

func returnStats(w http.ResponseWriter, r *http.Request) {

	// Parse template
	//t, err := template.ParseFiles("internal/web/templates/stats.html")
	t, err := template.New("stats.html").Parse(statsPageTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Prepare data for template
	templateRender := struct {
		Statistics stats
	}{
		statistics,
	}

	t.Execute(w, templateRender)

}

func homePage(w http.ResponseWriter, r *http.Request) {

	redirectUrl := fmt.Sprintf("/node/%s", startNodeId)
	http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
}

func returnNodeInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse template
	//t, err := template.ParseFiles("internal/web/templates/index.html")
	t, err := template.New("index.html").Parse(indexPageTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Get current path of the choosen node
	currentPath := tree.GetCurrentPath(myDriveTree, id)

	// Get current node data
	currentNode := myDriveTree[id]

	// Get child list of the selected node
	childList := tree.GetChildList(myDriveTree, id)

	// Prepare data for template
	templateRender := struct {
		CurrentPath []tree.MyPath
		CurrentNode tree.MyDrive
		Child       []tree.MyDrive
	}{
		currentPath,
		*currentNode,
		childList,
	}

	t.Execute(w, templateRender)

}

func Run(myTree map[string]*tree.MyDrive, startId string) {
	myDriveTree = myTree
	startNodeId = startId

	// generate statistics
	statistics = getStats()

	fmt.Println("Starting web server on http://localhost:8080...")

	myRouter = mux.NewRouter().StrictSlash(true)
	handleRequests()
	srv := &http.Server{
		Handler: myRouter,
		Addr:    ":8080",
	}

	url := fmt.Sprintf("http://localhost:8080/node/%s", startNodeId)
	errBrowser := browser.OpenURL(url)
	if errBrowser != nil {
		fmt.Println("Browse to " + url)
	}

	log.Fatal(srv.ListenAndServe())
}
