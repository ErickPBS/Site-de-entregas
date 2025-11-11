let carrinho = [];
const FRETE_FIXO = 8.0;

function carregarCarrinho() {
  try {
    const salvo = localStorage.getItem("carrinho");
    carrinho = salvo ? JSON.parse(salvo) : [];
  } catch (_) {
    carrinho = [];
  }
}

function salvarCarrinho() {
  localStorage.setItem("carrinho", JSON.stringify(carrinho));
}

function addCarrinho(produto, preco) {
  carregarCarrinho();
  const idx = carrinho.findIndex(i => i.produto === produto);
  if (idx >= 0) {
    carrinho[idx].qtd = (carrinho[idx].qtd || 1) + 1;
  } else {
    carrinho.push({ produto, preco: Number(preco), qtd: 1 });
  }
  salvarCarrinho();
  alert(produto + " adicionado ao carrinho!");
}

function removerItem(index) {
  carregarCarrinho();
  carrinho.splice(index, 1);
  salvarCarrinho();
  renderCarrinho();
}

function incItem(index) {
  carregarCarrinho();
  carrinho[index].qtd = (carrinho[index].qtd || 1) + 1;
  salvarCarrinho();
  renderCarrinho();
}

function decItem(index) {
  carregarCarrinho();
  const atual = carrinho[index];
  const novaQtd = (atual.qtd || 1) - 1;
  if (novaQtd <= 0) {
    carrinho.splice(index, 1);
  } else {
    atual.qtd = novaQtd;
  }
  salvarCarrinho();
  renderCarrinho();
}

function renderCarrinho() {
  carregarCarrinho();
  const tbody = document.getElementById("cart-body");
  const totalEl = document.getElementById("cart-total");
  const subtotalEl = document.getElementById("cart-subtotal");
  const freteEl = document.getElementById("cart-frete");
  const checkoutBtn = document.querySelector('a.checkout');
  if (!tbody) return;
  tbody.innerHTML = "";
  let subtotal = 0;
  carrinho.forEach((item, idx) => {
    const qtd = item.qtd || 1;
    const linha = (Number(item.preco) || 0) * qtd;
    subtotal += linha;
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${item.produto}</td>
      <td>
        <button onclick="decItem(${idx})">-</button>
        <span style="display:inline-block;width:24px;text-align:center;">${qtd}</span>
        <button onclick="incItem(${idx})">+</button>
      </td>
      <td>R$ ${linha.toFixed(2)}</td>
      <td><button class="danger" onclick="removerItem(${idx})">Remover</button></td>
    `;
    tbody.appendChild(tr);
  });
  const frete = carrinho.length ? FRETE_FIXO : 0;
  const total = subtotal + frete;
  if (subtotalEl) subtotalEl.textContent = `R$ ${subtotal.toFixed(2)}`;
  if (freteEl) freteEl.textContent = `R$ ${frete.toFixed(2)}`;
  if (totalEl) totalEl.textContent = `R$ ${total.toFixed(2)}`;

  // Oculta o botão de checkout quando o carrinho estiver vazio
  if (checkoutBtn) {
    if (carrinho.length === 0) {
      checkoutBtn.style.display = 'none';
    } else {
      checkoutBtn.style.display = 'inline-block';
    }
  }
}

document.addEventListener("DOMContentLoaded", () => {
  if (document.getElementById("cart-body")) {
    renderCarrinho();
  }
});
