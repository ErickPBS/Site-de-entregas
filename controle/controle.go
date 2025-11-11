package controle

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "html/template"
    "time"
)

var temp = func() *template.Template {
    // Tenta carregar a partir do diretÃ³rio atual
    if t, err := template.ParseGlob("Templates/*.html"); err == nil {
        return t
    }
    // Fallback: tenta com caminho absoluto do diretÃ³rio de trabalho
    if wd, err := os.Getwd(); err == nil {
        if t, err := template.ParseGlob(filepath.Join(wd, "Templates", "*.html")); err == nil {
            return t
        }
    }
    panic("falha ao carregar templates (Templates/*.html)")
}()

// render centraliza cabeÃ§alho correto e execuÃ§Ã£o do template
func render(w http.ResponseWriter, name string, data any) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _ = temp.ExecuteTemplate(w, name, data)
}

// Modelo simples em memÃ³ria (MVP)
type User struct {
    Name  string
    Email string
    Pass  string
    Role  string // cliente | fornecedor | entregador
    Phone string
    Rua   string
    Bairro string
    Numero string
    Cidade string
}

type Session struct {
    Email string
    Role  string
}

var (
    usersMu    sync.RWMutex
    users      = map[string]User{}    // key: email
    sessions   = map[string]Session{} // key: sessionID
    sessionsMu sync.RWMutex
)

type PageData struct {
    Error   string
    Success string
    User    User
    Aguardando []Pedido
    Preparando []Pedido
    Pronto     []Pedido
}

// Pedidos (MVP em memÃ³ria)
type Item struct {
    Nome string
    Qtd  int
}

type Pedido struct {
    ID     string
    Itens  []Item
    Status string // aguardando | preparando | pronto
    Hora   int    // hora de criaÃ§Ã£o (0-23)
    Loja   string // nome do estabelecimento (opcional)
    Cliente string // email do cliente (opcional)
    Endereco string // endereço de entrega (opcional)
}

var (
    ordersMu sync.RWMutex
    orders   = map[string]*Pedido{}
    seeded   bool
)

func seedOrders() {
    // Se já existem pedidos (carregados do disco), não semear
    ordersMu.RLock()
    if len(orders) > 0 || seeded { ordersMu.RUnlock(); return }
    ordersMu.RUnlock()
    ordersMu.Lock()
    defer ordersMu.Unlock()
    if len(orders) > 0 || seeded { return }
    cliente := "erick.antunes0@gmail.com"
    orders["12351"] = &Pedido{ID: "12351", Status: "aguardando", Hora: 10, Loja: "Restô A", Cliente: cliente, Itens: []Item{{Nome: "Burger Clássico", Qtd: 2}, {Nome: "Batata Média", Qtd: 1}}}
    orders["12352"] = &Pedido{ID: "12352", Status: "aguardando", Hora: 11, Loja: "Restô B", Cliente: cliente, Itens: []Item{{Nome: "Pizza Margherita", Qtd: 1}, {Nome: "Refrigerante Lata", Qtd: 1}}}
    orders["12346"] = &Pedido{ID: "12346", Status: "preparando", Hora: 12, Loja: "Restô C", Cliente: cliente, Itens: []Item{{Nome: "Sushi Combo", Qtd: 3}}}
    orders["12349"] = &Pedido{ID: "12349", Status: "preparando", Hora: 14, Loja: "Restô A", Cliente: cliente, Itens: []Item{{Nome: "Salada Caesar", Qtd: 1}, {Nome: "Suco de Laranja", Qtd: 2}}}
    orders["12347"] = &Pedido{ID: "12347", Status: "pronto", Hora: 13, Loja: "Restô B", Cliente: cliente, Itens: []Item{{Nome: "Burger Clássico", Qtd: 1}, {Nome: "Refrigerante Lata", Qtd: 1}}}
    seeded = true
}

func randID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}

func setSession(w http.ResponseWriter, email, role string) {
    sid := randID()
    sessionsMu.Lock()
    sessions[sid] = Session{Email: email, Role: role}
    sessionsMu.Unlock()
    http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/", HttpOnly: true})
}

func GetSession(r *http.Request) (Session, bool) {
    c, err := r.Cookie("sid")
    if err != nil || c.Value == "" {
        return Session{}, false
    }
    sessionsMu.RLock()
    s, ok := sessions[c.Value]
    sessionsMu.RUnlock()
    return s, ok
}

func clearSession(w http.ResponseWriter, r *http.Request) {
    if c, err := r.Cookie("sid"); err == nil {
        sessionsMu.Lock()
        delete(sessions, c.Value)
        sessionsMu.Unlock()
    }
    http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
}

// Root: se logado, vai para Ã¡rea do papel; senÃ£o, exibe login
func Root(w http.ResponseWriter, r *http.Request) {
    if s, ok := GetSession(r); ok {
        redirectByRole(w, r, s.Role)
        return
    }
    render(w, "login.html", PageData{})
}

// Login: GET mostra tela; POST autentica
func Login(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        render(w, "login.html", PageData{})
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            render(w, "login.html", PageData{Error: "RequisiÃ§Ã£o invÃ¡lida"})
            return
        }
        email := r.FormValue("email")
        pass := r.FormValue("password")
        usersMu.RLock()
        u, ok := users[email]
        usersMu.RUnlock()
        if !ok || u.Pass != pass {
            w.WriteHeader(http.StatusUnauthorized)
            render(w, "login.html", PageData{Error: "UsuÃ¡rio ou senha invÃ¡lidos"})
            return
        }
        setSession(w, u.Email, u.Role)
        redirectByRole(w, r, u.Role)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

// Register: GET mostra tela; POST cria usuÃ¡rio
func Register(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        render(w, "register.html", PageData{})
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            render(w, "register.html", PageData{Error: "RequisiÃ§Ã£o invÃ¡lida"})
            return
        }
        name := r.FormValue("name")
        email := r.FormValue("email")
        pass := r.FormValue("password")
        role := r.FormValue("role")
        phone := normalizePhone(r.FormValue("telefone"))
        rua := r.FormValue("rua")
        bairro := r.FormValue("bairro")
        numero := r.FormValue("numero")
        cidade := r.FormValue("cidade")
        if name == "" || email == "" || pass == "" || role == "" || phone == "" || rua == "" || bairro == "" || numero == "" || cidade == "" {
            w.WriteHeader(http.StatusBadRequest)
            render(w, "register.html", PageData{Error: "Preencha todos os campos"})
            return
        }
        usersMu.Lock()
        users[email] = User{Name: name, Email: email, Pass: pass, Role: role, Phone: phone, Rua: rua, Bairro: bairro, Numero: numero, Cidade: cidade}
        usersMu.Unlock()
        // Persistir usuÃ¡rios e exibir sucesso no login
        _ = saveUsers()
        render(w, "login.html", PageData{Success: "Conta criada! FaÃ§a login."})
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func Logout(w http.ResponseWriter, r *http.Request) {
    clearSession(w, r)
    http.Redirect(w, r, "/login", http.StatusFound)
}

// UtilitÃ¡rio: recupera usuÃ¡rio logado
func currentUser(r *http.Request) (User, bool) {
    if s, ok := GetSession(r); ok {
        usersMu.RLock()
        u, ok2 := users[s.Email]
        usersMu.RUnlock()
        return u, ok2
    }
    return User{}, false
}

// Home do cliente (dinÃ¢mica, com nome)
func HomeCliente(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "cliente" {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    render(w, "home_cliente.html", PageData{User: u})
}

// Painel do fornecedor (dinÃ¢mica, com nome)
func Painel(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "fornecedor" {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    seedOrders()
    switch r.Method {
    case http.MethodPost:
        if err := r.ParseForm(); err == nil {
            id := r.FormValue("id")
            novo := r.FormValue("status")
            ordersMu.Lock()
            if p, ok := orders[id]; ok {
                if novo == "aguardando" || novo == "preparando" || novo == "pronto" {
                    p.Status = novo
                    // Se ficou pronto, disponibiliza para o entregador como corrida "aceita"
                    if novo == "pronto" {
                        deliveriesMu.Lock()
                        if _, exists := deliveries[p.ID]; !exists {
                            // tenta obter telefone do cliente
                            usersMu.RLock()
                            tel := users[p.Cliente].Phone
                            usersMu.RUnlock()
                            deliveries[p.ID] = &Corrida{ID: p.ID, Loja: p.Loja, Endereco: p.Endereco, Distancia: 0, Status: "aceita", Phone: tel}
                        }
                        deliveriesMu.Unlock()
                        _ = saveDeliveries()
                    }
                }
            }
            ordersMu.Unlock()
            _ = saveOrders()
        }
        http.Redirect(w, r, "/painel", http.StatusFound)
        return
    }
    // GET: agrupa por status
    var agu, prep, pro []Pedido
    ordersMu.RLock()
    for _, p := range orders {
        switch p.Status {
        case "aguardando":
            agu = append(agu, *p)
        case "preparando":
            prep = append(prep, *p)
        case "pronto":
            pro = append(pro, *p)
        }
    }
    ordersMu.RUnlock()
    render(w, "restaurante.html", PageData{User: u, Aguardando: agu, Preparando: prep, Pronto: pro})
}

// --- API: mÃ©tricas do estabelecimento ---
type MixItem struct {
    Label string  `json:"label"`
    Value int     `json:"value"`
    Color string  `json:"color"`
    Amount float64 `json:"amount"`
}

type Metrics struct {
    PedidosHoje    int       `json:"pedidosHoje"`
    TempoMedioMin  int       `json:"tempoMedioMin"`
    AvaliacaoMedia float64   `json:"avaliacaoMedia"`
    TicketMedio    float64   `json:"ticketMedio"`
    ReceitaHoje    float64   `json:"receitaHoje"`
    Horas          []string  `json:"horas"`
    PedidosPorHora []int     `json:"pedidosPorHora"`
    Mix            []MixItem `json:"mix"`
    StatusLabels   []string  `json:"statusLabels"`
    StatusValues   []int     `json:"statusValues"`
}

func MetricsEstabelecimento(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "fornecedor" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    seedOrders()
    // AgregaÃ§Ãµes simples
    horasSet := map[int]int{}
    cats := map[string]int{}
    catAmounts := map[string]float64{}
    status := map[string]int{"aguardando":0, "preparando":0, "pronto":0}
    var totalReceita float64
    ordersMu.RLock()
    for _, p := range orders {
        horasSet[p.Hora]++
        status[p.Status]++
        var totalPedido float64
        for _, it := range p.Itens {
            cats[categoriaDe(it.Nome)] += it.Qtd
            totalPedido += float64(it.Qtd) * precoDe(it.Nome)
            catAmounts[categoriaDe(it.Nome)] += float64(it.Qtd) * precoDe(it.Nome)
        }
        totalReceita += totalPedido
    }
    ordersCount := len(orders)
    ordersMu.RUnlock()

    // Ordena horas de 10..17 para ficar prÃ³ximo ao template
    labels := []string{"10h","11h","12h","13h","14h","15h","16h","17h"}
    vals := make([]int, len(labels))
    base := 10
    for i := range labels {
        vals[i] = horasSet[base+i]
    }
    mix := []MixItem{
        {Label: "Burgers", Value: cats["Burgers"], Color: "#ff3d3d", Amount: catAmounts["Burgers"]},
        {Label: "Pizzas", Value: cats["Pizzas"], Color: "#f59e0b", Amount: catAmounts["Pizzas"]},
        {Label: "Sushi", Value: cats["Sushi"], Color: "#10b981", Amount: catAmounts["Sushi"]},
        {Label: "Saladas", Value: cats["Saladas"], Color: "#3b82f6", Amount: catAmounts["Saladas"]},
    }

    m := Metrics{
        PedidosHoje:    ordersCount,
        TempoMedioMin:  28, // placeholder simples
        AvaliacaoMedia: 4.6,
        TicketMedio:    0,
        ReceitaHoje:    totalReceita,
        Horas:          labels,
        PedidosPorHora: vals,
        Mix:            mix,
    }
    if ordersCount > 0 {
        m.TicketMedio = totalReceita / float64(ordersCount)
    }
    // Status em vetores paralelos
    m.StatusLabels = []string{"Aguardando","Preparando","Pronto"}
    m.StatusValues = []int{status["aguardando"], status["preparando"], status["pronto"]}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(m)
}

func categoriaDe(nome string) string {
    // ClassificaÃ§Ã£o simples por substring
    if containsFold(nome, "burger") { return "Burgers" }
    if containsFold(nome, "pizza") { return "Pizzas" }
    if containsFold(nome, "sushi") { return "Sushi" }
    if containsFold(nome, "salada") { return "Saladas" }
    return "Outros"
}

func precoDe(nome string) float64 {
    // preÃ§os aproximados por categoria/termo
    if containsFold(nome, "burger") { return 25.0 }
    if containsFold(nome, "batata") { return 12.0 }
    if containsFold(nome, "pizza") { return 45.0 }
    if containsFold(nome, "sushi") { return 59.0 }
    if containsFold(nome, "salada") { return 22.0 }
    if containsFold(nome, "refrigerante") { return 8.0 }
    if containsFold(nome, "suco") { return 10.0 }
    return 20.0
}

func containsFold(s, sub string) bool {
    // case-insensitive contains sem importar strings
    bS := []rune(s)
    bSub := []rune(sub)
    n, m := len(bS), len(bSub)
    if m == 0 { return true }
    for i := 0; i+m <= n; i++ {
        ok := true
        for j := 0; j < m; j++ {
            r := bS[i+j]
            if r >= 'A' && r <= 'Z' { r = r - 'A' + 'a' }
            rr := bSub[j]
            if rr >= 'A' && rr <= 'Z' { rr = rr - 'A' + 'a' }
            if r != rr { ok = false; break }
        }
        if ok { return true }
    }
    return false
}

// Painel do entregador (dinÃ¢mica, com nome)
func PainelEntregador(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "entregador" {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    render(w, "delivery_panel.html", PageData{User: u})
}

func redirectByRole(w http.ResponseWriter, r *http.Request, role string) {
    switch role {
    case "cliente":
        http.Redirect(w, r, "/Templates/products.html", http.StatusFound)
    case "fornecedor":
        http.Redirect(w, r, "/Templates/dashboard_estabelecimento.html", http.StatusFound)
    case "entregador":
        http.Redirect(w, r, "/Templates/dashboard_entregador.html", http.StatusFound)
    default:
        http.Redirect(w, r, "/Templates/index.html", http.StatusFound)
    }
}

// Perfil do cliente: GET exibe dados; POST atualiza
func Perfil(w http.ResponseWriter, r *http.Request) {
    s, ok := GetSession(r)
    if !ok {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    usersMu.RLock()
    u, ok := users[s.Email]
    usersMu.RUnlock()
    if !ok {
        http.Redirect(w, r, "/logout", http.StatusFound)
        return
    }

    switch r.Method {
    case http.MethodGet:
        render(w, "perfil_cliente.html", PageData{User: u})
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            render(w, "perfil_cliente.html", PageData{Error: "RequisiÃ§Ã£o invÃ¡lida", User: u})
            return
        }
        u.Name = r.FormValue("name")
        u.Phone = normalizePhone(r.FormValue("telefone"))
        u.Rua = r.FormValue("rua")
        u.Bairro = r.FormValue("bairro")
        u.Numero = r.FormValue("numero")
        u.Cidade = r.FormValue("cidade")
        usersMu.Lock()
        users[u.Email] = u
        usersMu.Unlock()
        _ = saveUsers()
        render(w, "perfil_cliente.html", PageData{Success: "Perfil atualizado", User: u})
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

// PersistÃªncia simples em arquivo JSON
func usersFile() string { return filepath.Join("DB", "users.json") }

func saveUsers() error {
    usersMu.RLock()
    defer usersMu.RUnlock()
    if err := os.MkdirAll("DB", 0o755); err != nil { return err }
    f, err := os.Create(usersFile())
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(users)
}

func loadUsers() {
    b, err := os.ReadFile(usersFile())
    if err != nil { return }
    var m map[string]User
    if json.Unmarshal(b, &m) == nil {
        usersMu.Lock()
        users = m
        usersMu.Unlock()
    }
}
func init() { loadUsers(); loadProducts(); loadOrders(); loadDeliveries() }

// Normaliza telefone mantendo apenas dígitos
func normalizePhone(s string) string {
    if s == "" { return s }
    out := make([]rune, 0, len(s))
    for _, r := range s {
        if r >= '0' && r <= '9' { out = append(out, r) }
    }
    return string(out)
}

// ---------------- Produtos -----------------
type Produto struct {
    ID        string  `json:"id"`
    Nome      string  `json:"nome"`
    Preco     float64 `json:"preco"`
    Descricao string  `json:"descricao"`
    Imagem    string  `json:"imagem"` // URL relativa para /static/uploads/... ou URL externa
}

var (
    produtosMu sync.RWMutex
    produtos   = map[string]Produto{}
)

func productsFile() string { return filepath.Join("DB", "products.json") }

func saveProducts() error {
    produtosMu.RLock()
    defer produtosMu.RUnlock()
    if err := os.MkdirAll("DB", 0o755); err != nil { return err }
    f, err := os.Create(productsFile())
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(produtos)
}

func loadProducts() {
    b, err := os.ReadFile(productsFile())
    if err != nil { return }
    var m map[string]Produto
    if json.Unmarshal(b, &m) == nil {
        produtosMu.Lock()
        produtos = m
        produtosMu.Unlock()
    }
}

// GET lista, POST cadastra (fornecedor)
func Produtos(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        produtosMu.RLock()
        list := make([]Produto, 0, len(produtos))
        for _, p := range produtos { list = append(list, p) }
        produtosMu.RUnlock()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(list)
    case http.MethodPost:
        u, ok := currentUser(r)
        if !ok || u.Role != "fornecedor" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
        // aceita multipart para upload opcional de imagem
        r.ParseMultipartForm(10 << 20) // 10MB
        nome := r.FormValue("nome")
        precoStr := r.FormValue("preco")
        desc := r.FormValue("descricao")
        imgURL := r.FormValue("imagem_url")
        var preco float64
        fmt.Sscanf(precoStr, "%f", &preco)
        if nome == "" || preco <= 0 { http.Error(w, "invalid", http.StatusBadRequest); return }
        imgPath := imgURL
        if file, hdr, err := r.FormFile("imagem_arquivo"); err == nil {
            defer file.Close()
            _ = os.MkdirAll(filepath.Join("static", "uploads"), 0o755)
            id := randID()
            safe := filepath.Base(hdr.Filename)
            pathRel := filepath.Join("static", "uploads", id+"_"+safe)
            out, err := os.Create(pathRel)
            if err == nil {
                defer out.Close()
                _, _ = out.ReadFrom(file)
                imgPath = "/" + pathRel
            }
        }
        id := randID()
        p := Produto{ID: id, Nome: nome, Preco: preco, Descricao: desc, Imagem: imgPath}
        produtosMu.Lock(); produtos[id] = p; produtosMu.Unlock()
        _ = saveProducts()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(p)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

// POST /api/produtos/delete (fornecedor)
func ProdutoDelete(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "fornecedor" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
    if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
    if err := r.ParseForm(); err != nil { http.Error(w, "bad request", http.StatusBadRequest); return }
    id := r.FormValue("id")
    if id == "" { http.Error(w, "bad request", http.StatusBadRequest); return }
    produtosMu.Lock(); delete(produtos, id); produtosMu.Unlock()
    _ = saveProducts()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// POST /api/produtos/update (fornecedor)
func ProdutoUpdate(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "fornecedor" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
    if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
    r.ParseMultipartForm(10 << 20)
    id := r.FormValue("id")
    if id == "" { http.Error(w, "bad request", http.StatusBadRequest); return }
    produtosMu.Lock()
    p, ok := produtos[id]
    if !ok { produtosMu.Unlock(); http.Error(w, "not found", http.StatusNotFound); return }
    if v := r.FormValue("nome"); v != "" { p.Nome = v }
    if v := r.FormValue("preco"); v != "" { var pr float64; fmt.Sscanf(v, "%f", &pr); if pr > 0 { p.Preco = pr } }
    if v := r.FormValue("descricao"); v != "" { p.Descricao = v }
    if v := r.FormValue("imagem_url"); v != "" { p.Imagem = v }
    if file, hdr, err := r.FormFile("imagem_arquivo"); err == nil {
        defer file.Close()
        _ = os.MkdirAll(filepath.Join("static", "uploads"), 0o755)
        safe := filepath.Base(hdr.Filename)
        pathRel := filepath.Join("static", "uploads", id+"_"+safe)
        if out, err := os.Create(pathRel); err == nil { defer out.Close(); _, _ = out.ReadFrom(file); p.Imagem = "/"+pathRel }
    }
    produtos[id] = p
    produtosMu.Unlock()
    _ = saveProducts()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// ---------------- Entregas (painel do entregador) -----------------
type Corrida struct {
    ID        string  `json:"id"`
    Loja      string  `json:"loja"`
    Endereco  string  `json:"endereco"`
    Distancia float64 `json:"distancia"`
    Status    string  `json:"status"` // aceita | rota | concluida
    Phone     string  `json:"telefone"`
}

var (
    deliveriesMu   sync.RWMutex
    deliveries     = map[string]*Corrida{}
    deliveriesOnce bool
)

// Persistência simples das corridas do entregador
func deliveriesFile() string { return filepath.Join("DB", "deliveries.json") }

func saveDeliveries() error {
    deliveriesMu.RLock()
    defer deliveriesMu.RUnlock()
    if err := os.MkdirAll("DB", 0o755); err != nil { return err }
    f, err := os.Create(deliveriesFile())
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(deliveries)
}

func loadDeliveries() {
    b, err := os.ReadFile(deliveriesFile())
    if err != nil { return }
    var m map[string]*Corrida
    if json.Unmarshal(b, &m) == nil {
        deliveriesMu.Lock()
        deliveries = m
        deliveriesMu.Unlock()
        deliveriesOnce = true
    }
}

func seedDeliveries() {
    if deliveriesOnce { return }
    deliveriesMu.Lock()
    defer deliveriesMu.Unlock()
    if deliveriesOnce { return }
    deliveries["12344"] = &Corrida{ID: "12344", Loja: "RestÃ´ A", Endereco: "Av. Brasil, 100", Distancia: 3.2, Status: "aceita"}
    deliveries["12343"] = &Corrida{ID: "12343", Loja: "RestÃ´ B", Endereco: "Rua das Flores, 25", Distancia: 1.8, Status: "aceita"}
    deliveries["12340"] = &Corrida{ID: "12340", Loja: "RestÃ´ C", Endereco: "Rua 7, 330", Distancia: 4.6, Status: "rota"}
    deliveriesOnce = true
}

// GET /api/entregas -> lista de corridas do entregador
func EntregasList(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "entregador" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    seedDeliveries()
    deliveriesMu.RLock()
    list := make([]Corrida, 0, len(deliveries))
    for _, c := range deliveries {
        list = append(list, *c)
    }
    deliveriesMu.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)
}

// POST /api/entregas/status (id, status)
func EntregasUpdate(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "entregador" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    if err := r.ParseForm(); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    id := r.FormValue("id")
    status := r.FormValue("status") // aceita -> rota; rota -> concluida
    if id == "" || (status != "rota" && status != "concluida") {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    deliveriesMu.Lock()
    if c, ok := deliveries[id]; ok {
        if c.Status == "aceita" && status == "rota" {
            c.Status = "rota"
        } else if c.Status == "rota" && status == "concluida" {
            c.Status = "concluida"
        }
    }
    deliveriesMu.Unlock()
    _ = saveDeliveries()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// ---------------- Pedidos do cliente (API) -----------------
type PedidoResp struct {
    ID     string `json:"id"`
    Loja   string `json:"loja"`
    Itens  []Item `json:"itens"`
    Total  float64 `json:"total"`
    Status string `json:"status"` // agu | pre | ent
}

// ---------------- Perfil do usuÃ¡rio (API) -----------------
// GET /api/me -> retorna dados bÃ¡sicos do usuÃ¡rio logado
func Me(w http.ResponseWriter, r *http.Request) {
    s, ok := GetSession(r)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    usersMu.RLock()
    u, ok := users[s.Email]
    usersMu.RUnlock()
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(u)
}

// GET /api/cliente/pedidos
func PedidosClienteList(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "cliente" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    seedOrders()
    res := []PedidoResp{}
    ordersMu.RLock()
    // verifica se hÃ¡ associaÃ§Ã£o por cliente
    filterByClient := false
    for _, p := range orders { if p.Cliente != "" { filterByClient = true; break } }
    for id, p := range orders {
        if filterByClient && p.Cliente != u.Email { continue }
        // mapear status fornecedor -> cliente
        st := "agu"
        if p.Status == "preparando" { st = "pre" } else if p.Status == "pronto" { st = "ent" }
        // total
        var total float64
        for _, it := range p.Itens {
            total += float64(it.Qtd) * precoDe(it.Nome)
        }
        loja := p.Loja
        if loja == "" { loja = "Restaurante" }
        res = append(res, PedidoResp{ID: id, Loja: loja, Itens: p.Itens, Total: total, Status: st})
    }
    ordersMu.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}

// ---------------- Checkout: criar pedido (cliente) -----------------
type CheckoutReq struct {
    Loja    string `json:"loja"`
    Endereco string `json:"endereco"`
    Itens   []Item `json:"itens"`
}

// POST /api/checkout {loja,endereco,itens:[{Nome,Qtd}]}
func CheckoutCriarPedido(w http.ResponseWriter, r *http.Request) {
    u, ok := currentUser(r)
    if !ok || u.Role != "cliente" {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var req CheckoutReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if len(req.Itens) == 0 {
        http.Error(w, "empty cart", http.StatusBadRequest)
        return
    }
    if req.Loja == "" { req.Loja = "RestÃ´ A" }
    id := randID()
    p := &Pedido{ID: id, Itens: req.Itens, Status: "aguardando", Hora: time.Now().Hour(), Loja: req.Loja, Cliente: u.Email, Endereco: req.Endereco}
    ordersMu.Lock()
    orders[id] = p
    ordersMu.Unlock()
    _ = saveOrders()
    // resposta
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{"id": id, "status": p.Status})
}

// Persistência de pedidos em disco
func ordersFile() string { return filepath.Join("DB", "orders.json") }

func saveOrders() error {
    ordersMu.RLock()
    defer ordersMu.RUnlock()
    if err := os.MkdirAll("DB", 0o755); err != nil { return err }
    f, err := os.Create(ordersFile())
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(orders)
}

func loadOrders() {
    b, err := os.ReadFile(ordersFile())
    if err != nil { return }
    var m map[string]*Pedido
    if json.Unmarshal(b, &m) == nil {
        ordersMu.Lock()
        orders = m
        ordersMu.Unlock()
    }
}
