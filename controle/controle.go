package controle

import (
    "net/http"
    "text/template"
)

var temp = template.Must(template.ParseGlob("Templates/*.html"))

func Index(w http.ResponseWriter, r *http.Request) {
    temp.ExecuteTemplate(w, "index.html", nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
    temp.ExecuteTemplate(w, "login.html", nil)

}

// Painel renders the "restaurante" template and writes it to the http.ResponseWriter.
// It does not pass any data to the template.
// Parameters:
//
//	w - the http.ResponseWriter to write the response to.
//	r - the http.Request representing the incoming HTTP request.
func Painel(w http.ResponseWriter, r *http.Request) {
    temp.ExecuteTemplate(w, "restaurante.html", nil)
}
