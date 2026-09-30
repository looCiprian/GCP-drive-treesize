package web

import (
	"bytes"
	"context"
	"drive-tree/internal/authentication"
	"drive-tree/internal/tree"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
)

//go:embed templates/*.html
var templateFiles embed.FS

// Application phases
const (
	phaseSignedOut = "signed-out"
	phaseScanning  = "scanning"
	phaseReady     = "ready"
	phaseError     = "error"
)

type account struct {
	Name       string
	Email      string
	QuotaUsed  int64
	QuotaTotal int64
}

// Everything the application knows lives in memory, nothing is written to disk
type app struct {
	mu sync.RWMutex

	config  *oauth2.Config
	flow    *authentication.Flow
	srv     *drive.Service
	account account

	phase   string
	fetched int
	err     string

	myDriveTree map[string]*tree.MyDrive
	startNodeId string
	statistics  stats

	pages map[string]*template.Template
}

// Data shared by all pages
type base struct {
	Title    string
	Page     string
	SignedIn bool
	Ready    bool
	Account  account
}

type row struct {
	tree.MyDrive
	Percent float64
}

var templateFuncs = template.FuncMap{
	"humanBytes": func(b int64) string { return tree.HumanBytes(b) },
	"percent": func(part, total int64) string {
		if total <= 0 {
			return "0"
		}
		return fmt.Sprintf("%.1f", float64(part)*100/float64(total))
	},
	"kind":  fileKind,
	"lower": strings.ToLower,
}

func newApp() *app {
	a := &app{
		config: authentication.Config(),
		phase:  phaseSignedOut,
		pages:  make(map[string]*template.Template),
	}
	for _, page := range []string{"welcome", "scanning", "browse", "stats"} {
		a.pages[page] = template.Must(template.New("").Funcs(templateFuncs).
			ParseFS(templateFiles, "templates/layout.html", "templates/"+page+".html"))
	}
	return a
}

func (a *app) routes() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", a.homePage).Methods("GET")
	r.HandleFunc("/login", a.login).Methods("GET")
	r.HandleFunc("/oauth2callback", a.oauthCallback).Methods("GET")
	r.HandleFunc("/logout", a.logout).Methods("POST")
	r.HandleFunc("/rescan", a.rescan).Methods("POST")
	r.HandleFunc("/api/status", a.status).Methods("GET")
	r.HandleFunc("/node/{id}", a.returnNodeInfo).Methods("GET")
	r.HandleFunc("/stats", a.returnStats).Methods("GET")
	return r
}

func (a *app) base(title, page string) base {
	return base{
		Title:    title,
		Page:     page,
		SignedIn: a.srv != nil,
		Ready:    a.phase == phaseReady,
		Account:  a.account,
	}
}

// Render a page to a buffer first, so template errors don't produce half pages
func (a *app) render(w http.ResponseWriter, page string, data interface{}) {
	var buf bytes.Buffer
	if err := a.pages[page].ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Println(err)
		http.Error(w, "Unable to render page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (a *app) homePage(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch a.phase {
	case phaseReady:
		http.Redirect(w, r, "/node/"+a.startNodeId, http.StatusFound)
	case phaseScanning:
		a.render(w, "scanning", struct {
			base
			Fetched int
		}{a.base("Scanning", "scanning"), a.fetched})
	default:
		a.render(w, "welcome", struct {
			base
			Error string
		}{a.base("Welcome", "welcome"), a.err})
	}
}

// Redirect the user to the Google consent page
func (a *app) login(w http.ResponseWriter, r *http.Request) {
	flow, err := authentication.NewFlow()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.mu.Lock()
	a.flow = flow
	a.mu.Unlock()

	http.Redirect(w, r, flow.AuthURL(a.config), http.StatusFound)
}

// Google redirects here after consent: exchange the code and start scanning
func (a *app) oauthCallback(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	query := r.URL.Query()
	switch {
	case query.Get("error") != "":
		a.fail("Sign-in was cancelled (" + query.Get("error") + ").")
	case a.flow == nil || query.Get("state") != a.flow.State:
		a.fail("The sign-in link expired, please try again.")
	default:
		tok, err := a.flow.Exchange(r.Context(), a.config, query.Get("code"))
		if err != nil {
			a.fail("Unable to complete sign-in: " + err.Error())
			break
		}
		// The service outlives this request, so it gets its own context
		srv, err := authentication.NewService(context.Background(), a.config, tok)
		if err != nil {
			a.fail("Unable to create the Drive client: " + err.Error())
			break
		}
		a.srv = srv
		a.startScan()
	}
	a.flow = nil

	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	// Forget the token and all the data
	a.mu.Lock()
	a.flow, a.srv, a.account = nil, nil, account{}
	a.phase, a.fetched, a.err = phaseSignedOut, 0, ""
	a.myDriveTree, a.startNodeId, a.statistics = nil, "", stats{}
	a.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *app) rescan(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	if a.srv != nil && a.phase != phaseScanning {
		a.startScan()
	}
	a.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Polled by the scanning page
func (a *app) status(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Phase   string `json:"phase"`
		Fetched int    `json:"fetched"`
		Error   string `json:"error,omitempty"`
	}{a.phase, a.fetched, a.err})
}

// Must be called with a.mu held
func (a *app) fail(msg string) {
	log.Println(msg)
	a.phase = phaseError
	a.err = msg
}

// Scan the drive in background. Must be called with a.mu held.
func (a *app) startScan() {
	a.phase = phaseScanning
	a.fetched = 0
	a.err = ""
	srv := a.srv

	go func() {
		acc := fetchAccount(srv)
		a.mu.Lock()
		a.account = acc
		a.mu.Unlock()

		myTree, rootId, err := tree.Run(srv, func(fetched int) {
			a.mu.Lock()
			a.fetched = fetched
			a.mu.Unlock()
		})

		a.mu.Lock()
		defer a.mu.Unlock()
		if a.srv != srv { // Signed out in the meantime
			return
		}
		if err != nil {
			a.fail("Unable to read your files: " + err.Error())
			return
		}
		a.myDriveTree = myTree
		a.startNodeId = rootId
		a.statistics = getStats(myTree)
		a.phase = phaseReady
		log.Printf("All %d items were successfully parsed", len(myTree))
	}()
}

// Account info is only used for display, errors are not fatal
func fetchAccount(srv *drive.Service) account {
	about, err := srv.About.Get().Fields("user(displayName,emailAddress),quotaBytesTotal,quotaBytesUsed").Do()
	if err != nil {
		log.Printf("Unable to get account info: %v", err)
		return account{}
	}
	acc := account{QuotaUsed: about.QuotaBytesUsed, QuotaTotal: about.QuotaBytesTotal}
	if about.User != nil {
		acc.Name = about.User.DisplayName
		acc.Email = about.User.EmailAddress
	}
	return acc
}

func (a *app) returnNodeInfo(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.phase != phaseReady {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	id := mux.Vars(r)["id"]
	currentNode, ok := a.myDriveTree[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Get child list of the selected node, sorted by size
	var rows []row
	folders := 0
	for _, child := range tree.GetChildList(a.myDriveTree, id) {
		percent := 0.0
		if currentNode.Size > 0 {
			percent = float64(child.Size) * 100 / float64(currentNode.Size)
		}
		rows = append(rows, row{child, percent})
		if child.IsDir {
			folders++
		}
	}

	parentId := ""
	if _, ok := a.myDriveTree[currentNode.Parent]; ok && !currentNode.IsRoot {
		parentId = currentNode.Parent
	}

	a.render(w, "browse", struct {
		base
		CurrentPath []tree.MyPath
		CurrentNode tree.MyDrive
		ParentId    string
		Rows        []row
		Folders     int
		Files       int
	}{
		a.base(currentNode.Name, "browse"),
		tree.GetCurrentPath(a.myDriveTree, id),
		*currentNode,
		parentId,
		rows,
		folders,
		len(rows) - folders,
	})
}

func (a *app) returnStats(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.phase != phaseReady {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	a.render(w, "stats", struct {
		base
		Statistics stats
	}{a.base("Statistics", "stats"), a.statistics})
}

// Run starts the web server. Signing in, scanning and browsing all happen in
// the browser, so there is no need for token or data files.
func Run(addr string, openBrowser bool) {
	a := newApp()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Unable to listen on %s: %v", addr, err)
	}

	url := strings.TrimSuffix(authentication.RedirectURL, "/oauth2callback") + "/login"
	fmt.Println("Drive Tree is running, sign in at " + url)
	if openBrowser {
		if err := browser.OpenURL(url); err != nil {
			fmt.Println("Unable to open the browser, please open the link above manually")
		}
	}
	fmt.Println("Press Ctrl+C to quit")

	log.Fatal(http.Serve(listener, a.routes()))
}
