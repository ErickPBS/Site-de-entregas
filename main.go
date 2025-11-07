package main

import (
	"fmt"
	"modulo/rotas"
	"net/http"
)

func main() {
	rotas.Rotas()
	fmt.Println("servidor rodando na porta 8080")
	http.ListenAndServe(":8080", nil)
}
