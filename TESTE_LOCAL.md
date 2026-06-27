# Guia de Teste Local - Miauzap

## Pré-requisitos

✅ Go instalado (versão 1.21+)
✅ Código compilado com sucesso (`miauzap.exe` presente)

## Passo a Passo para Testar Localmente

### 1. Criar Arquivo de Configuração

Copie o arquivo `.env.sample` para `.env`:

```bash
copy .env.sample .env
```

### 2. Editar o Arquivo `.env`

Abra o arquivo `.env` e configure:

```env
# Configuração Mínima para Teste Local
MIAUZAP_PORT=8080
MIAUZAP_ADDRESS=0.0.0.0
MIAUZAP_ADMIN_TOKEN=meu_token_teste_123

# Chaves de segurança (32 caracteres mínimo)
MIAUZAP_GLOBAL_ENCRYPTION_KEY=12345678901234567890123456789012
MIAUZAP_GLOBAL_HMAC_KEY=minha_chave_hmac_teste_32_chars

# Dispositivo
SESSION_DEVICE_NAME=Miauzap

# Timezone
TZ=America/Sao_Paulo
```

**Nota:** O Miauzap usará SQLite por padrão se você não configurar PostgreSQL.

### 3. Executar o Miauzap

```bash
.\miauzap.exe
```

Ou com logs coloridos:

```bash
.\miauzap.exe -color
```

### 4. Acessar a Interface Web

Abra seu navegador em:

- **Dashboard:** http://localhost:8080/static/dashboard/
- **Login Simples:** http://localhost:8080/static/login/
- **Documentação:** http://localhost:8080/static/docs/
- **API Swagger:** http://localhost:8080/static/api/

### 5. Criar Primeiro Usuário (via API)

Use o token admin que você configurou no `.env`:

```bash
curl -X POST http://localhost:8080/admin/users ^
  -H "Authorization: meu_token_teste_123" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"Usuario1\",\"token\":\"token_usuario_1\"}"
```

**Resposta esperada:**
```json
{
  "code": 200,
  "data": {
    "id": "1",
    "name": "Usuario1",
    "token": "token_usuario_1"
  }
}
```

### 6. Fazer Login no Dashboard

1. Acesse: http://localhost:8080/static/dashboard/
2. Escolha "Admin Login"
3. Digite o token admin: `meu_token_teste_123`
4. Você verá a lista de instâncias

### 7. Conectar WhatsApp

**Opção A: Via Dashboard**
1. Clique na instância criada
2. Clique em "Login with QR"
3. Escaneie o QR Code com WhatsApp

**Opção B: Via API**
```bash
# Conectar
curl -X POST http://localhost:8080/session/connect ^
  -H "token: token_usuario_1" ^
  -H "Content-Type: application/json" ^
  -d "{\"subscribe\":[\"Message\"],\"immediate\":true}"

# Obter QR Code
curl -X GET http://localhost:8080/session/qr ^
  -H "token: token_usuario_1"
```

### 8. Verificar Status

```bash
curl -X GET http://localhost:8080/session/status ^
  -H "token: token_usuario_1"
```

### 9. Enviar Mensagem de Teste

```bash
curl -X POST http://localhost:8080/chat/send/text ^
  -H "token: token_usuario_1" ^
  -H "Content-Type: application/json" ^
  -d "{\"phone\":\"5511999999999\",\"body\":\"Olá do Miauzap!\"}"
```

## Verificações de Funcionamento

### ✅ Checklist de Teste

- [ ] Servidor inicia sem erros
- [ ] Dashboard carrega com branding "Miauzap Manager"
- [ ] Login com token admin funciona
- [ ] Criação de usuário via API funciona
- [ ] QR Code é gerado
- [ ] Conexão WhatsApp estabelecida
- [ ] Status mostra "connected" e "loggedIn"
- [ ] Envio de mensagem funciona
- [ ] Webhook recebe eventos (se configurado)

## Estrutura de Arquivos Criados

Após executar, o Miauzap criará:

```
miauzap/
├── dbdata/           # Banco de dados SQLite
│   └── main.db
└── .env              # Suas configurações (não versionado)
```

## Logs Importantes

O Miauzap mostrará logs como:

```
{"level":"info","time":"...","message":"Server started. Waiting for connections..."}
{"level":"info","address":"0.0.0.0","port":"8080","message":"Server started..."}
```

## Troubleshooting

### Erro: "unauthorized"
- Verifique se o token está correto
- Para admin: use header `Authorization: seu_token_admin`
- Para usuário: use header `token: token_do_usuario`

### Erro: "no session"
- Execute `/session/connect` antes de outras operações
- Verifique se o cliente foi criado corretamente

### Porta 8080 já em uso
- Altere `MIAUZAP_PORT` no `.env`
- Ou pare o processo que está usando a porta

### Banco de dados não cria
- Verifique permissões na pasta `dbdata/`
- Certifique-se que não há outro processo usando o arquivo

## Parar o Servidor

Pressione `Ctrl+C` no terminal onde o Miauzap está rodando.

## Próximos Passos

1. **Configurar Webhook:** Para receber eventos do WhatsApp
2. **Integrar Chatwoot:** Para CRM completo
3. **Configurar S3:** Para armazenamento de mídia
4. **Deploy em Produção:** Usar Docker ou VPS

## Recursos Adicionais

- **API Completa:** http://localhost:8080/static/api/
- **Documentação:** http://localhost:8080/static/docs/
- **README:** [README.md](README.md)
- **Guia Chatwoot:** [CHATWOOT.md](CHATWOOT.md)
