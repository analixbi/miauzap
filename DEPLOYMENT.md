# Miauzap Production Deployment Guide

## 📋 Overview

This guide covers deploying Miauzap to a VPS with Traefik reverse proxy and SSL certificates.

## 🚀 Quick Start - Production Deployment

### Prerequisites

- VPS with Docker and Docker Compose installed
- Domain name pointing to your VPS IP
- Traefik already running (or follow Traefik setup below)
- Portainer (optional, but recommended)

### Step 1: Prepare Environment

1. **Copy environment file:**
   ```bash
   cp .env.sample .env
   ```

2. **Edit `.env` with your production values:**
   ```bash
   nano .env
   ```

   **Critical settings to change:**
   ```env
   # Your domain URL (IMPORTANT!)
   SERVER_URL=https://miauzap.yourdomain.com
   
   # Strong security tokens
   MIAUZAP_ADMIN_TOKEN=your_very_secure_random_token_here
   MIAUZAP_GLOBAL_ENCRYPTION_KEY=generate_32_random_characters_here
   MIAUZAP_GLOBAL_HMAC_KEY=generate_another_32_random_chars
   
   # Database credentials
   DB_PASSWORD=strong_database_password
   
   # Traefik configuration
   TRAEFIK_DOMAIN=miauzap.yourdomain.com
   TRAEFIK_NETWORK=traefik-public
   TRAEFIK_ACME_EMAIL=your-email@domain.com
   ```

3. **Generate secure keys:**
   ```bash
   # Generate encryption key (32 bytes)
   openssl rand -base64 32
   
   # Generate HMAC key (32 bytes)
   openssl rand -base64 32
   
   # Generate admin token
   openssl rand -hex 32
   ```

### Step 2: Configure Docker Compose for Traefik

Edit `docker-compose.yml`:

1. **Uncomment Traefik labels** in `miauzap-server` service (lines ~70-95)
2. **Uncomment traefik-public network** in `miauzap-server` networks section
3. **Uncomment traefik-public** in the networks section at the bottom
4. **Comment out** the ports section in `miauzap-server` (since Traefik will handle routing)

### Step 3: Deploy

**Option A: Using Docker Compose directly**
```bash
docker compose up -d
```

**Option B: Using Portainer**
1. Go to Portainer UI
2. Stacks → Add Stack
3. Upload `docker-compose.yml`
4. Add environment variables from `.env`
5. Deploy

### Step 4: Verify Deployment

1. **Check containers:**
   ```bash
   docker compose ps
   ```

2. **Check logs:**
   ```bash
   docker compose logs -f miauzap-server
   ```

3. **Access your instance:**
   ```
   https://miauzap.yourdomain.com
   ```

## 🔧 Traefik Setup (If Not Already Running)

If you don't have Traefik running yet:

### 1. Create Traefik Network
```bash
docker network create traefik-public
```

### 2. Create Traefik Configuration

Create `traefik-compose.yml`:
```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    container_name: traefik
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    networks:
      - traefik-public
    ports:
      - "80:80"
      - "443:443"
    environment:
      - CF_API_EMAIL=${CF_API_EMAIL}
      - CF_DNS_API_TOKEN=${CF_DNS_API_TOKEN}
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik-data/traefik.yml:/traefik.yml:ro
      - ./traefik-data/acme.json:/acme.json
      - ./traefik-data/config.yml:/config.yml:ro
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik.entrypoints=http"
      - "traefik.http.routers.traefik.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.middlewares.traefik-auth.basicauth.users=admin:$$apr1$$..."
      - "traefik.http.middlewares.traefik-https-redirect.redirectscheme.scheme=https"
      - "traefik.http.middlewares.sslheader.headers.customrequestheaders.X-Forwarded-Proto=https"
      - "traefik.http.routers.traefik.middlewares=traefik-https-redirect"
      - "traefik.http.routers.traefik-secure.entrypoints=https"
      - "traefik.http.routers.traefik-secure.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.routers.traefik-secure.middlewares=traefik-auth"
      - "traefik.http.routers.traefik-secure.tls=true"
      - "traefik.http.routers.traefik-secure.tls.certresolver=cloudflare"
      - "traefik.http.routers.traefik-secure.tls.domains[0].main=yourdomain.com"
      - "traefik.http.routers.traefik-secure.tls.domains[0].sans=*.yourdomain.com"
      - "traefik.http.routers.traefik-secure.service=api@internal"

networks:
  traefik-public:
    external: true
```

### 3. Create `traefik-data/traefik.yml`:
```yaml
api:
  dashboard: true
  debug: true

entryPoints:
  http:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: https
          scheme: https
  https:
    address: ":443"

serversTransport:
  insecureSkipVerify: true

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    filename: /config.yml

certificatesResolvers:
  letsencrypt:
    acme:
      email: your-email@domain.com
      storage: acme.json
      httpChallenge:
        entryPoint: http
```

### 4. Start Traefik:
```bash
touch traefik-data/acme.json
chmod 600 traefik-data/acme.json
docker compose -f traefik-compose.yml up -d
```

## 🔐 Chatwoot Integration Setup

Once deployed, the webhook URL will be automatically generated:

1. **Access Miauzap dashboard:**
   ```
   https://miauzap.yourdomain.com
   ```

2. **Create an instance and configure Chatwoot:**
   - Go to instance settings
   - Click "Chatwoot Integration"
   - Fill in your Chatwoot details:
     - Chatwoot URL: `https://app.chatwoot.com`
     - Account ID: Your Chatwoot account ID
     - API Token: Your Chatwoot API token
   - Save configuration

3. **Copy the generated webhook URL:**
   - The webhook URL will be automatically generated as:
     ```
     https://miauzap.yourdomain.com/chatwoot/webhook/{instance_id}
     ```
   - This URL is shown in the "Webhook URL (Read-only)" field
   - Click the copy button to copy it

4. **Configure in Chatwoot:**
   - Go to your Chatwoot inbox settings
   - Paste the webhook URL
   - Save

## 📊 Monitoring

### View Logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f miauzap-server
docker compose logs -f db
docker compose logs -f rabbitmq
```

### Check Service Status
```bash
docker compose ps
```

### Access RabbitMQ Management (if ports exposed)
```
http://your-vps-ip:15672
Username: miauzap
Password: miauzap (or your configured password)
```

## 🔄 Updates

### Update Miauzap
```bash
# Pull latest changes
git pull

# Rebuild and restart
docker compose up -d --build
```

### Backup Database
```bash
# Create backup
docker compose exec db pg_dump -U miauzap miauzap > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
docker compose exec -T db psql -U miauzap miauzap < backup_file.sql
```

## 🛠️ Troubleshooting

### Container won't start
```bash
# Check logs
docker compose logs miauzap-server

# Check if ports are in use
netstat -tulpn | grep 8080
```

### Chatwoot webhook not working
1. Verify `SERVER_URL` is set correctly in `.env`
2. Check that domain DNS is pointing to VPS
3. Verify Traefik is routing correctly
4. Check Miauzap logs for webhook errors

### Database connection issues
```bash
# Check database is healthy
docker compose exec db pg_isready -U miauzap

# Restart database
docker compose restart db
```

### SSL Certificate issues
```bash
# Check Traefik logs
docker logs traefik

# Verify acme.json permissions
ls -la traefik-data/acme.json  # Should be 600
```

## 📝 Important Notes

1. **Always backup your `.env` file** - especially the encryption keys
2. **Keep `MIAUZAP_GLOBAL_ENCRYPTION_KEY` safe** - data cannot be recovered without it
3. **Use strong passwords** in production
4. **Set `SERVER_URL` correctly** - this is critical for Chatwoot webhook generation
5. **Monitor disk space** - database and RabbitMQ can grow over time

## 🆘 Support

For issues and questions:
- Check logs first: `docker compose logs -f`
- Verify environment variables in `.env`
- Ensure domain DNS is configured correctly
- Check Traefik routing and SSL certificates

## 🎉 Success!

If everything is working:
- ✅ Miauzap accessible at `https://miauzap.yourdomain.com`
- ✅ SSL certificate active (green padlock)
- ✅ Chatwoot webhook URL auto-generated correctly
- ✅ WhatsApp integration working
- ✅ Messages syncing with Chatwoot

You're all set! 🚀
