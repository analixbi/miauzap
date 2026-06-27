# MIAUZAP

<img src="static/favicon.ico" width="30"> Miauzap is a WhatsApp API implementation based on [@tulir/whatsmeow](https://github.com/tulir/whatsmeow) with **Chatwoot CRM integration**. It provides a simple RESTful API service with multiple device support, concurrent sessions, and seamless Chatwoot integration for managing WhatsApp conversations.

---

### 💡 Mantenha Este Projeto Vivo! / Keep This Project Alive! / ¡Mantén Este Proyecto Vivo!

| 🇧🇷 Português | 🇺🇸 English | 🇪🇸 Español |
| :--- | :--- | :--- |
| Manter uma API de alta qualidade e com updates constantes exige tempo e recursos. Se o MiauZap está a poupar tempo e dinheiro à sua empresa, **com certeza vale 10 dólares para ajudar a manter o projeto vivo e a evoluir!** Apoie-nos para garantir que tem sempre as melhores ferramentas à sua disposição. | Maintaining a high-quality, constantly updated API takes time and resources. If MiauZap is saving your business time and money, **it is definitely worth a $10 contribution to keep the project alive and thriving!** Support us to ensure you always have the best tools at your disposal. | Mantener una API de alta calidad con actualizaciones constantes requiere tiempo y recursos. Si MiauZap está ahorrando tiempo y dinero a tu negocio, **¡definitivamente vale una contribución de 10 dólares para mantener el proyecto vivo y en crecimiento!** Apóyanos para asegurarte de tener siempre las mejores herramientas a tu disposición. |

<div align="center">
  <img src="image (1).jpg" alt="PayPal QR Code" width="250"/>
  <br>
  <i>Scan to Donate via PayPal / Faça a leitura para doar via PayPal / Escanea para donar vía PayPal</i>
</div>
---

## 🎁 Brinde Especial: Plugin WooCommerce para WordPress! / Special Bonus: WooCommerce WordPress Plugin!

Miauzap includes a complete, production-ready WordPress plugin for WooCommerce integration as a free bonus! You can find it inside the [`wordpress-plugin`](wordpress-plugin/) directory (available as both source code and a pre-packaged `.zip` file).

### 🚀 Principais Funcionalidades do Plugin / Key Plugin Features:
* 🔐 **Login por OTP via WhatsApp:** Permita que seus clientes façam login seguro sem senhas tradicionais.
* 📦 **Mensagens automáticas por status de pedido:** Envie notificações automáticas de atualização de pedido (como códigos de rastreio, envio, etc.) usando variáveis dinâmicas.
* 🔀 **Instâncias com rotação ponderada:** Faça a rotação automática de múltiplas instâncias do Miauzap para balanceamento de carga e segurança contra banimentos.
* 🕒 **Fila inteligente com controle anti-ban:** Fila com atraso aleatório configurável, limites de disparo por dia por instância e horário silencioso (evita mensagens de madrugada).
* 📝 **Campos personalizados:** Suporta variáveis dinâmicas de anotações de pedidos e campos personalizados/meta do WooCommerce (como códigos de rastreio dos Correios).

Consulte o [**README do Plugin**](wordpress-plugin/miauzap-woocommerce/README.md) para obter instruções detalhadas de instalação e configuração.

---


Whatsmeow does not use Puppeteer on headless Chrome, nor an Android emulator. It communicates directly with WhatsApp’s WebSocket servers, making it significantly faster and much less demanding on memory and CPU than those solutions. The drawback is that any changes to the WhatsApp protocol could break connections, requiring a library update.

## :warning: Warning

**Using this software in violation of WhatsApp’s Terms of Service can get your number banned**:  
Be very careful—do not use this to send SPAM or anything similar. Use at your own risk. If you need to develop something for commercial purposes, contact a WhatsApp global solution provider and sign up for the WhatsApp Business API service instead.

## Available endpoints

* **Session:** Connect, disconnect, and log out from WhatsApp. Retrieve connection status and QR codes for scanning.
* **Messages:** Send text, image, audio, document, template, video, sticker, location, contact, and poll messages.
* **Users:** Check if phone numbers have WhatsApp, get user information and avatars, and retrieve the full contact list.
* **Chat:** Set presence (typing/paused, recording media), mark messages as read, download images from messages, send reactions.
* **Groups:** Create, delete and list groups, get info, get invite links, set participants, change group photos and names.
* **Webhooks:** Set and get webhooks that will be called whenever events or messages are received.
* **HMAC Configuration:** Configure HMAC keys for webhook security and signature verification.
* **Chatwoot Integration:** Full CRM integration with automatic contact sync, conversation management, and bidirectional messaging.

## Chatwoot Integration

Miauzap includes powerful Chatwoot integration features:

- ✅ **Automatic Contact Sync** - WhatsApp contacts are automatically created in Chatwoot
- ✅ **Bidirectional Messaging** - Send and receive messages between WhatsApp and Chatwoot
- ✅ **Conversation Management** - Conversations are automatically created and managed
- ✅ **Media Support** - Images, videos, audio, and documents are forwarded
- ✅ **Group Support** - Handle group conversations with participant tracking
- ✅ **Profile Picture Sync** - Contact avatars are synchronized
- ✅ **Brazilian Phone Numbers** - Smart handling of Brazilian number variations

See [CHATWOOT.md](CHATWOOT.md) for detailed integration documentation.


### Webhook HMAC Signing

When HMAC is configured, all webhooks include an `x-hmac-signature` header with SHA-256 HMAC signature.

#### Signature Generation by Content-Type:

**`application/json`**
* Signed data: Raw JSON request body
* Verification: Use the exact JSON received

**`application/x-www-form-urlencoded`**
* Signed data: URL-encoded form string (`key=value&key2=value2`)
* Verification: Reconstruct the form string from received parameters

**`multipart/form-data`** (file uploads)
* Signed data: JSON representation of form fields (excluding files)
* Verification: Create JSON from non-file form fields

* Always verify signatures before processing webhooks

## Prerequisites

**Required:**
* Go (Go Programming Language)

**Optional:**
* Docker (for containerization)

## Updating dependencies

This project uses the whatsmeow library to communicate with WhatsApp. To update the library to the latest version, run:

```bash
go get -u go.mau.fi/whatsmeow@latest
go mod tidy
```

## Building

```
go build .
```

## Homebrew installation

To install `miauzap` via [Homebrew](https://brew.sh) use:

```sh
brew install analixbi/miauzap/miauzap
```

## Run

By default it will start a REST service in port 8080. These are the parameters
you can use to alter behaviour

* -admintoken  : sets authentication token for admin endpoints. If not specified it will be read from .env
* -address  : sets the IP address to bind the server to (default 0.0.0.0)
* -port  : sets the port number (default 8080)
* -logtype : format for logs, either console (default) or json
* -color : enable colored output for console logs
* -osname : Connection OS Name in Whatsapp
* -skipmedia : Skip downloading media from messages
* -wadebug : enable whatsmeow debug, either INFO or DEBUG levels are suported

* -sslcertificate : SSL Certificate File
* -sslprivatekey : SSL Private Key File

Example:

To have colored logs:

```
./miauzap -logtype=console -color=true
```

For JSON logs:

```
./miauzap -logtype json 
```

With time zone: 

Set `TZ=America/New_York ./miauzap ...` in your shell or in your .env file or Docker Compose environment: `TZ=America/New_York`.  

## Configuration

Miauzap uses a `.env` file for configuration. You can use the provided `.env.sample` as a template:

```bash
cp .env.sample .env
```

### Environment Variables

#### Required Settings
```
MIAUZAP_ADMIN_TOKEN=your_admin_token_here
```

#### Security Settings

```
MIAUZAP_GLOBAL_ENCRYPTION_KEY=your_32_byte_encryption_key_here
MIAUZAP_GLOBAL_HMAC_KEY=your_global_hmac_key_here
```

#### Optional Settings

```
TZ=America/New_York
WEBHOOK_FORMAT=json
SESSION_DEVICE_NAME=Miauzap
MIAUZAP_PORT=8080
MIAUZAP_GLOBAL_WEBHOOK=https://your-global-webhook.url
WEBHOOK_RETRY_ENABLED=true
WEBHOOK_RETRY_COUNT=2
WEBHOOK_RETRY_DELAY_SECONDS=30
WEBHOOK_ERROR_QUEUE_NAME=miauzap_dead_letter_webhooks
```

### Important Notes

#### Auto-Generated Credentials
If the following settings are not provided, they will be auto-generated:
* `MIAUZAP_ADMIN_TOKEN`: Random 32-character token
* `MIAUZAP_GLOBAL_ENCRYPTION_KEY`: Random 32-byte key for AES-256 encryption

**Important**: Save auto-generated credentials to your `.env` file or you will lose access to encrypted data and admin functions on restart!

#### Webhook Security
* `MIAUZAP_GLOBAL_HMAC_KEY`: Global HMAC key for webhook signing (minimum 32 characters)

#### Database Configuration

**For PostgreSQL:**
```
DB_USER=miauzap
DB_PASSWORD=miauzap
DB_NAME=miauzap
DB_HOST=db  # Use 'db' when running with Docker Compose, or 'localhost' for native execution
DB_PORT=5432
DB_SSLMODE=false
```

**For SQLite (default):**
No database configuration needed - SQLite is used by default if no PostgreSQL settings are provided.

#### Optional Settings
```
TZ=America/New_York
WEBHOOK_FORMAT=json # or "form" for the default
SESSION_DEVICE_NAME=Miauzap
MIAUZAP_PORT=8080 # Port for the Miauzap server
MIAUZAP_GLOBAL_WEBHOOK= # Global webhook URL for all instances
```

### RabbitMQ Integration
Miauzap supports sending WhatsApp events to a RabbitMQ queue for global event distribution. When enabled, all WhatsApp events will be published to the specified queue regardless of individual user webhook configurations.

Set these environment variables to enable RabbitMQ integration:

```
RABBITMQ_URL=amqp://guest:guest@localhost:5672
RABBITMQ_QUEUE=whatsapp  # Optional (default: whatsapp_events)
```

When enabled:

* All WhatsApp events (messages, presence updates, etc.) will be published to the configured queue regardless of event subscritions for regular webhooks
* Events will include the userId and instanceName
* This works alongside webhook configurations - events will be sent to both RabbitMQ and any configured webhooks
* The integration is global and affects all instances

### Webhook Security with HMAC

Miauzap supports HMAC signatures for webhook verification:

* **Per-instance HMAC**: Configure unique HMAC keys for each user instance
* **Global HMAC**: Set a global HMAC key via `MIAUZAP_GLOBAL_HMAC_KEY` environment variable
* **Signature Header**: All signed webhooks include `x-hmac-signature` header
* **Key Security**: HMAC keys are never exposed after configuration

**Priority**: Instance HMAC > Global HMAC > No signature

Configure HMAC keys via the Dashboard or using the `/session/hmac/config` API endpoints.

#### Key configuration options:

* MIAUZAP_ADMIN_TOKEN: Required - Authentication token for admin endpoints
* TZ: Optional - Timezone for server operations (default: UTC)
* PostgreSQL-specific options: Only required when using PostgreSQL backend
* RabbitMQ options: Optional, only required if you want to publish events to RabbitMQ

### Docker Configuration

When using Docker Compose, `docker-compose.yml` automatically loads environment variables from a `.env` file when available. However, `docker-compose-swarm.yaml` uses `docker stack deploy`, which does not automatically load from `.env` files. Variables in the swarm file will only be substituted if they are exported in the shell environment where the deploy command is run. For managing secrets in Swarm, consider using Docker secrets.

The Docker configuration will:
1. First load variables from the `.env` file (if present and supported)
2. Use default values as fallback if variables are not defined
3. Override with any variables explicitly set in the `environment` section of the compose file

**Key differences for Docker deployment:**
- Set `DB_HOST=db` instead of `localhost` to connect to the PostgreSQL container
- The `MIAUZAP_PORT` variable controls the external port mapping in `docker-compose.yml`
- In swarm mode, `MIAUZAP_PORT` configures the Traefik load balancer port

**Note:** The `.env` file is already included in `.gitignore` to avoid committing sensitive information to your repository.

## Usage

To interact with the API, you must include the `Authorization` header in HTTP requests, containing the user's authentication token. You can have multiple users (different WhatsApp numbers) on the same server.  

* A Swagger API reference at [/api](/api)
* A sample web page to connect and scan QR codes at [/login](/login)
* A fully featured Dashboard to create, manage and test instances at [/dashboard](dashboard)

## ADMIN Actions

You can list, add and remove users using the admin endpoints. For that you must use the MIAUZAP_ADMIN_TOKEN in the Authorization header

Then you can use the /admin/users endpoint with the Authorization header containing the token to:

- `GET /admin/users` - List all users
- `POST /admin/users` - Create a new user
- `DELETE /admin/users/{id}` - Remove a user

The JSON body for creating a new user must contain:

- `name` [string] : User's name 
- `token` [string] : Security token to authorize/authenticate this user
- `webhook` [string] : URL to send events via POST (optional)
- `events` [string] : Comma-separated list of events to receive (required) - Valid events are: "Message", "ReadReceipt", "Presence", "HistorySync", "ChatPresence", "All"
- `expiration` [int] : Expiration timestamp (optional, not enforced by the system)

## User Creation with Optional Proxy and S3 Configuration

You can create a user with optional proxy and S3 storage configuration. All fields are optional and backward compatible. If you do not provide these fields, the user will be created with default settings.

### Example Payload

```json
{
  "name": "test_user",
  "token": "user_token",
  "proxyConfig": {
    "enabled": true,
    "proxyURL": "socks5://user:pass@host:port"
  },
  "s3Config": {
    "enabled": true,
    "endpoint": "https://s3.amazonaws.com",
    "region": "us-east-1",
    "bucket": "my-bucket",
    "accessKey": "AKIAIOSFODNN7EXAMPLE",
    "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "pathStyle": false,
    "publicURL": "https://cdn.yoursite.com",
    "mediaDelivery": "both",
    "retentionDays": 30
  }
}
```

- `proxyConfig` (object, optional):
  - `enabled` (boolean): Enable proxy for this user.
  - `proxyURL` (string): Proxy URL (e.g., `socks5://user:pass@host:port`).
- `s3Config` (object, optional):
  - `enabled` (boolean): Enable S3 storage for this user.
  - `endpoint` (string): S3 endpoint URL.
  - `region` (string): S3 region.
  - `bucket` (string): S3 bucket name.
  - `accessKey` (string): S3 access key.
  - `secretKey` (string): S3 secret key.
  - `pathStyle` (boolean): Use path style addressing.
  - `publicURL` (string): Public URL for accessing files.
  - `mediaDelivery` (string): Media delivery type (`base64`, `s3`, or `both`).
  - `retentionDays` (integer): Number of days to retain files.

If you omit `proxyConfig` or `s3Config`, the user will be created without proxy or S3 integration, maintaining full backward compatibility.

## Homebrew installation

To install `miauzap` via [Homebrew](https://brew.sh) use:

```sh
brew install analixbi/miauzap/miauzap
```

## API reference 

API calls should be made with content type json, and parameters sent into the
request body, always passing the Token header for authenticating the request.

Check the [API Reference](https://github.com/analixbi/miauzap/blob/main/API.md)

## Contributors

<table>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href="https://github.com/analixbi">
            <img src="https://avatars.githubusercontent.com/u/25182694?v=4" width="100;" style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt="analixbi"/>
            <br />
            <sub style="font-size:14px"><b>analixbi</b></sub>
        </a>
    </td>
</tr>
</table>

## Clients

- [miauzap TypeScript / Node Client](https://github.com/gusnips/miauzap-node)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=analixbi/miauzap&type=Date)](https://www.star-history.com/#analixbi/miauzap&Date)

## License

Copyright &copy; 2026 Analix

[MIT](https://choosealicense.com/licenses/mit/)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Icon Attribution

[Communication icons created by Vectors Market -
Flaticon](https://www.flaticon.com/free-icons/communication)

## Legal

This code is in no way affiliated with, authorized, maintained, sponsored or
endorsed by WhatsApp or any of its affiliates or subsidiaries. This is an
independent and unofficial software. Use at your own risk.

## Cryptography Notice

This distribution includes cryptographic software. The country in which you
currently reside may have restrictions on the import, possession, use, and/or
re-export to another country, of encryption software. BEFORE using any
encryption software, please check your country's laws, regulations and policies
concerning the import, possession, or use, and re-export of encryption
software, to see if this is permitted. See
[http://www.wassenaar.org/](http://www.wassenaar.org/) for more information.

The U.S. Government Department of Commerce, Bureau of Industry and Security
(BIS), has classified this software as Export Commodity Control Number (ECCN)
5D002.C.1, which includes information security software using or performing
cryptographic functions with asymmetric algorithms. The form and manner of this
distribution makes it eligible for export under the License Exception ENC
Technology Software Unrestricted (TSU) exception (see the BIS Export
Administration Regulations, Section 740.13) for both object code and source
code.
