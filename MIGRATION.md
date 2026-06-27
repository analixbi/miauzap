# Migration Guide

This guide helps you understand the new Miauzap branding.

## Environment Variables

All environment variables use the `MIAUZAP_*` prefix.

| Variable | Description |
|----------|-------------|
| `MIAUZAP_ADDRESS` | Server listen address (default: 0.0.0.0) |
| `MIAUZAP_PORT` | Server listen port (default: 8080) |
| `MIAUZAP_ADMIN_TOKEN` | Token for admin routes |
| `MIAUZAP_GLOBAL_ENCRYPTION_KEY` | Key for encrypting session data |
| `MIAUZAP_GLOBAL_HMAC_KEY` | Key for HMAC signing |
| `MIAUZAP_GLOBAL_WEBHOOK` | Optional global webhook URL |

### Example `.env`

```env
MIAUZAP_PORT=8080
MIAUZAP_ADDRESS=0.0.0.0
MIAUZAP_ADMIN_TOKEN=your_token_here
MIAUZAP_GLOBAL_ENCRYPTION_KEY=your_key_here
MIAUZAP_GLOBAL_HMAC_KEY=your_hmac_key_here
```

## Docker

We now use the `analixbi/miauzap` image.

```yaml
services:
  miauzap-server:
    image: analixbi/miauzap:latest
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - ./miauzap_data:/app/files
    networks:
      - miauzap-network

networks:
  miauzap-network:

### Update Your Database Configuration

```env
DB_USER=miauzap
DB_PASSWORD=miauzap
DB_NAME=miauzap
```


## RabbitMQ Configuration

Update RabbitMQ connection strings:

```env
RABBITMQ_URL=amqp://miauzap:miauzap@localhost:5672/
```

## Session Device Name

The default device name is:

```env
SESSION_DEVICE_NAME=Miauzap
```

## API Endpoints

All API endpoints remain the same. No changes required to your API calls.

## Branding Changes

- Logo and favicons have been updated to Miauzap branding
- Dashboard title changed from "WuzAPI Manager" to "Miauzap Manager"
- All documentation updated with new branding

## Migration Checklist

- [ ] Update environment variables in `.env` file
- [ ] Update `docker-compose.yml` service names
- [ ] Update network names in Docker configuration
- [ ] Update database credentials (if using defaults)
- [ ] Update RabbitMQ connection string
- [ ] Rebuild Docker images with new branding
- [ ] Test application startup
- [ ] Verify API functionality



## Support

For issues or questions about the migration, please open an issue on the GitHub repository.
