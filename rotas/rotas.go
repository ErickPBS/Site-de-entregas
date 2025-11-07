package rotas

import (
	"modulo/controle"
	"net/http" 
)

func Rotas() {

    http.HandleFunc("/", controle.Index)
    http.HandleFunc("/login", controle.Login)
    http.HandleFunc("/painel", controle.Painel)

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Serve assets referenced by templates (Templates/assets/...)
    http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("Templates/assets"))))

}
