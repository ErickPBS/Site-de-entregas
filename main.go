package main

import (
    "log"
    "net/http"

    "modulo/rotas"
)

func main() {
    // Registra rotas de aplicação (inclui guarda para /Templates/)
    rotas.Rotas()

    log.Println("Servidor rodando em http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
