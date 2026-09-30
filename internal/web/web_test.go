package web

import (
	"drive-tree/internal/tree"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"google.golang.org/api/drive/v2"
)

func fakeTree() map[string]*tree.MyDrive {
	return map[string]*tree.MyDrive{
		"root":   {Id: "root", Name: "My Drive", IsRoot: true, IsDir: true, Size: 3000, HumanSize: "2.9 KiB", NumberOfFiles: 3, Child: []string{"photos", "notes"}},
		"photos": {Id: "photos", Name: "Photos <2021>", Parent: "root", IsDir: true, Size: 2500, HumanSize: "2.4 KiB", NumberOfFiles: 2, Child: []string{"a", "b"}, MimeType: "application/vnd.google-apps.folder", Link: "https://drive.google.com/photos"},
		"a":      {Id: "a", Name: "beach.jpg", Parent: "photos", Size: 2000, HumanSize: "2.0 KiB", MimeType: "image/jpeg", Link: "https://drive.google.com/a"},
		"b":      {Id: "b", Name: "clip.mp4", Parent: "photos", Size: 500, HumanSize: "500 B", MimeType: "video/mp4", Link: "https://drive.google.com/b"},
		"notes":  {Id: "notes", Name: "notes.txt", Parent: "root", Size: 500, HumanSize: "500 B", MimeType: "text/plain", Link: "https://drive.google.com/notes"},
	}
}

func get(t *testing.T, a *app, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if dir := os.Getenv("DUMP_DIR"); dir != "" && w.Code == http.StatusOK {
		name := strings.NewReplacer("/", "_", "?", "_").Replace(path)
		os.WriteFile(dir+"/"+name+".html", w.Body.Bytes(), 0644)
	}
	return w
}

func TestSignedOut(t *testing.T) {
	a := newApp()

	if w := get(t, a, "/"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `href="/login"`) {
		t.Fatalf("welcome page: %d %s", w.Code, w.Body)
	}
	if w := get(t, a, "/stats"); w.Code != http.StatusFound {
		t.Fatalf("stats should redirect when not ready, got %d", w.Code)
	}

	w := get(t, a, "/login")
	loc, err := url.Parse(w.Header().Get("Location"))
	if w.Code != http.StatusFound || err != nil || loc.Host != "accounts.google.com" {
		t.Fatalf("login should redirect to Google, got %d %q", w.Code, w.Header().Get("Location"))
	}
	q := loc.Query()
	if q.Get("state") != a.flow.State || q.Get("code_challenge_method") != "S256" || q.Get("redirect_uri") == "" {
		t.Fatalf("bad auth URL: %s", loc)
	}

	// A callback with the wrong state is rejected
	get(t, a, "/oauth2callback?state=wrong&code=x")
	if a.phase != phaseError || a.srv != nil {
		t.Fatalf("callback with bad state was accepted")
	}
	if w := get(t, a, "/"); !strings.Contains(w.Body.String(), "expired") {
		t.Fatalf("error not shown on welcome page")
	}
}

func TestScanning(t *testing.T) {
	a := newApp()
	a.phase, a.fetched = phaseScanning, 1234

	if w := get(t, a, "/"); !strings.Contains(w.Body.String(), "Scanning your Drive") {
		t.Fatalf("scanning page not shown")
	}
	if w := get(t, a, "/api/status"); !strings.Contains(w.Body.String(), `"fetched":1234`) {
		t.Fatalf("bad status: %s", w.Body)
	}
}

func TestReady(t *testing.T) {
	a := newApp()
	a.phase, a.myDriveTree, a.startNodeId = phaseReady, fakeTree(), "root"
	a.statistics = getStats(a.myDriveTree)
	a.srv = &drive.Service{} // Signed in
	a.account = account{Email: "me@example.com", QuotaUsed: 5 << 30, QuotaTotal: 15 << 30}

	if w := get(t, a, "/"); w.Code != http.StatusFound || w.Header().Get("Location") != "/node/root" {
		t.Fatalf("home should redirect to root, got %d", w.Code)
	}

	body := get(t, a, "/node/root").Body.String()
	for _, want := range []string{"Photos &lt;2021&gt;", "83.3%", "me@example.com", "33.3%", `action="/rescan"`} {
		if !strings.Contains(body, want) {
			t.Errorf("root page missing %q", want)
		}
	}
	if strings.Index(body, "Photos") > strings.Index(body, "notes.txt") {
		t.Errorf("children not sorted by size")
	}

	body = get(t, a, "/node/photos").Body.String()
	if !strings.Contains(body, `href="/node/root">My Drive`) || !strings.Contains(body, "beach.jpg") {
		t.Errorf("photos page missing breadcrumb or children")
	}

	get(t, a, "/node/b")

	if w := get(t, a, "/node/missing"); w.Code != http.StatusNotFound {
		t.Errorf("unknown node should 404, got %d", w.Code)
	}

	body = get(t, a, "/stats").Body.String()
	for _, want := range []string{"image/jpeg", "beach.jpg", "/My Drive/Photos &lt;2021&gt;"} {
		if !strings.Contains(body, want) {
			t.Errorf("stats page missing %q", want)
		}
	}

	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, httptest.NewRequest("POST", "/logout", nil))
	if a.phase != phaseSignedOut || a.myDriveTree != nil {
		t.Errorf("logout did not clear data")
	}
}
