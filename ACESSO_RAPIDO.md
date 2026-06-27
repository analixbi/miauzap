# 🚀 Miauzap - URLs Corretas para Acesso

## ✅ Servidor Rodando

**Status:** ATIVO  
**Endereço:** http://localhost:8080  
**Token Admin:** `pog6H7NvnlYTyVWEBRhpOFdmzQkdCn8p`

---

## 🌐 URLs Corretas de Acesso

### Páginas Principais

| Página | URL Correta | Descrição |
|--------|-------------|-----------|
| **Dashboard** | http://localhost:8080/dashboard | Interface principal de gerenciamento |
| **Login** | http://localhost:8080/login | Página de login simples |
| **Página Inicial** | http://localhost:8080/ | Página inicial com links |
| **Health Check** | http://localhost:8080/health | Status do servidor (JSON) |

### Documentação

| Recurso | URL | Descrição |
|---------|-----|-----------|
| **API Docs** | http://localhost:8080/static/api/ | Documentação Swagger da API |
| **Docs** | http://localhost:8080/static/docs/ | Documentação completa |

---

## 🔐 Como Fazer Login

### Opção 1: Login Admin (Recomendado)

1. Acesse: http://localhost:8080/dashboard
2. Clique no botão **"Admin Login"**
3. Digite o token admin: `pog6H7NvnlYTyVWEBRhpOFdmzQkdCn8p`
4. Clique em **"Login"**

### Opção 2: Login de Usuário

1. Primeiro, crie um usuário via API (veja abaixo)
2. Acesse: http://localhost:8080/dashboard
3. Clique no botão **"Login"** (não "Admin Login")
4. Digite o token do usuário criado

---

## 📝 Criar Primeiro Usuário (via API)

```bash
curl -X POST http://localhost:8080/admin/users ^
  -H "Authorization: pog6H7NvnlYTyVWEBRhpOFdmzQkdCn8p" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"Usuario1\",\"token\":\"meu_token_usuario_123\"}"
```

**Resposta esperada:**
```json
{
  "code": 200,
  "data": {
    "id": "1",
    "name": "Usuario1",
    "token": "meu_token_usuario_123"
  }
}
```

---

## 🎯 Próximos Passos

### 1. Acessar Dashboard
- Abra: http://localhost:8080/dashboard
- Faça login com o token admin

### 2. Criar Instância WhatsApp
- No dashboard, clique em **"Add Instance"**
- Ou use a API para criar usuário

### 3. Conectar WhatsApp
- Clique na instância criada
- Escolha **"Login with QR"**
- Escaneie o QR Code com WhatsApp

### 4. Testar Envio de Mensagem
```bash
curl -X POST http://localhost:8080/chat/send/text ^
  -H "token: meu_token_usuario_123" ^
  -H "Content-Type: application/json" ^
  -d "{\"phone\":\"5511999999999\",\"body\":\"Olá do Miauzap!\"}"
```

---

## ⚠️ Observações Importantes

### Problema com `.env` Detectado
O arquivo `.env` tem um caractere BOM que impede a leitura. Por isso o servidor gerou tokens aleatórios.

**Para corrigir:**
1. Abra o `.env` no VS Code ou Notepad++
2. Salve com codificação **UTF-8 sem BOM**
3. Reinicie o servidor

### Tokens Temporários
Os tokens atuais são temporários e serão perdidos ao reiniciar o servidor. Corrija o `.env` para usar tokens fixos.

---

## 🔍 Verificar Status

### Via Browser
- http://localhost:8080/health

### Via API
```bash
curl http://localhost:8080/health
```

**Resposta:**
```json
{
  "status": "ok",
  "version": "1.0.5",
  "active_connections": 0,
  ...
}
```

---

## 🛑 Parar o Servidor

Pressione **Ctrl+C** no terminal onde o Miauzap está rodando.

---

## ✅ Checklist de Teste

- [x] Servidor iniciado
- [x] Dashboard acessível em http://localhost:8080/dashboard
- [x] Login page acessível em http://localhost:8080/login
- [x] Health check funcionando
- [ ] Login admin realizado
- [ ] Usuário criado via API
- [ ] WhatsApp conectado
- [ ] Mensagem enviada com sucesso

---

## 📚 Recursos Adicionais

- **Documentação Completa:** [README.md](README.md)
- **Guia de Teste Local:** [TESTE_LOCAL.md](TESTE_LOCAL.md)
- **API Reference:** [API.md](API.md)
- **Integração Chatwoot:** [CHATWOOT.md](CHATWOOT.md)
