package authentication

import (
	"html/template"
	"net/http"
)

// configuring http server to handle oauth2 callback
func callbackListener() *http.Server {

	http.HandleFunc("/oauth2callback", callbackHandler)

	server := &http.Server{Addr: ":8080", Handler: nil}
	go listenAndServe(server)
	return server
}

// listenAndServe starts a http server to handle oauth2 callback
func listenAndServe(server *http.Server) {
	server.ListenAndServe()
}

// handle oauth2 endpoint
func callbackHandler(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	w.Header().Add("Content Type", "text/html")
	templateaasdf := `<h1>Successfully authenticated. <br>Please copy this value, swith to the application and past it there<br><input type='text' style="width: 500px;" value="{{.Code}}"> </h1>`

	data := struct {
		Code string
	}{
		code,
	}

	mynewtemplate, _ := template.New("oauth").Parse(templateaasdf)
	mynewtemplate.Execute(w, data)

}

// close http server to handle oauth2 callback
func closeCallbackListener(server *http.Server) {
	server.Close()
}
