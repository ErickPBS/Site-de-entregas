package rotas

import (
    "modulo/controle"
    "net/http"
)

func Rotas() {
    // Raiz: direciona via controle.Root (login se não autenticado)
    http.HandleFunc("/", controle.Root)

    // Autenticação
    http.HandleFunc("/login", controle.Login)
    http.HandleFunc("/register", controle.Register)
    http.HandleFunc("/logout", controle.Logout)
    http.HandleFunc("/perfil", controle.Perfil)
    http.HandleFunc("/entregador", controle.PainelEntregador)

    // Exemplo de painel existente
    http.HandleFunc("/painel", controle.Painel)

    // Estáticos
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("Templates/assets"))))
    http.HandleFunc("/api/estab/metrics", controle.MetricsEstabelecimento)
    // API entregas (entregador)
    http.HandleFunc("/api/entregas", controle.EntregasList)
    http.HandleFunc("/api/entregas/status", controle.EntregasUpdate)
    // API pedidos do cliente
    http.HandleFunc("/api/cliente/pedidos", controle.PedidosClienteList)
    // API produtos
    http.HandleFunc("/api/produtos", controle.Produtos)
    http.HandleFunc("/api/produtos/update", controle.ProdutoUpdate)
    http.HandleFunc("/api/produtos/delete", controle.ProdutoDelete)
    // API de perfil (dados do usuário logado)
    http.HandleFunc("/api/me", controle.Me)
    // Checkout: criar pedido
    http.HandleFunc("/api/checkout", controle.CheckoutCriarPedido)
    // Servir templates com guarda de sessão/role
    http.Handle("/Templates/", authTemplates(http.Dir("Templates")))
}

func authTemplates(dir http.FileSystem) http.Handler {
    // Necessário retirar o prefixo "/Templates/" antes de servir os arquivos
    // senão o FileServer tenta abrir Templates/Templates/arquivo.html
    fs := http.StripPrefix("/Templates/", http.FileServer(dir))
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Permitir apenas base.html sem sessão (login/registro via handlers)
        p := r.URL.Path
        if p == "/Templates/base.html" {
            fs.ServeHTTP(w, r)
            return
        }
        // Exigir sessão para demais
        s, ok := controle.GetSession(r)
        if !ok {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        // Restrições simples por papel
    if (containsAny(p, []string{"home_cliente", "perfil_cliente", "pedidos", "products", "cart", "checkout", "confirmacao", "carrinho"}) && s.Role != "cliente") ||
            (containsAny(p, []string{"restaurante", "dashboard_estabelecimento", "cadastro"}) && s.Role != "fornecedor") ||
            (containsAny(p, []string{"delivery_panel", "dashboard_entregador"}) && s.Role != "entregador") {
            http.Redirect(w, r, "/", http.StatusFound)
            return
        }
        fs.ServeHTTP(w, r)
    })
}

func containsAny(s string, subs []string) bool {
    for _, sub := range subs {
        if len(sub) > 0 && contains(s, sub) {
            return true
        }
    }
    return false
}

func contains(s, sub string) bool { return len(sub) <= len(s) && (func() bool { return stringIndex(s, sub) >= 0 })() }

// stringIndex: pequena função para evitar importar strings, mantendo leve.
func stringIndex(haystack, needle string) int {
    // implementação simples O(n*m)
    n, m := len(haystack), len(needle)
    if m == 0 { return 0 }
    for i := 0; i+m <= n; i++ {
        if haystack[i:i+m] == needle {
            return i
        }
    }
    return -1
}



