# Service Marketplace

Marketplace integration microservice for Shopee and TikTok Shop.

## Features

- 🔐 **OAuth 2.0 Authentication** - Secure connection to Shopee and TikTok Shop
- 📦 **Product Sync** - Push products to marketplaces with category mapping
- 📊 **Inventory Sync** - Real-time stock updates via NATS events
- 🛒 **Order Import** - Webhook-driven order synchronization
- 🔒 **Token Encryption** - AES-256 encryption for access tokens

## Quick Start

### 1. Run Database Migration

```bash
psql -U postgres -d ecommerce_db -f migrations/001_create_marketplace_schema.sql
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your credentials
```

### 3. Get Marketplace Credentials

#### Shopee Open Platform
1. Register at https://open.shopee.com/
2. Create an App (use Sandbox for testing)
3. Get Partner ID and Partner Key
4. Set redirect URL to: `http://your-domain/api/v1/admin/marketplace/shopee/callback`

#### TikTok Shop Partner API
1. Register at https://partner.tiktokshop.com/
2. Create a Test App
3. Get App Key and App Secret
4. Set redirect URL to: `http://your-domain/api/v1/admin/marketplace/tiktok/callback`

### 4. Generate Encryption Key

```bash
# Generate a 32-byte key for AES-256
openssl rand -hex 32
```

Add to `.env`:
```
MARKETPLACE_ENCRYPTION_KEY=<your-32-byte-hex-key>
```

### 5. Run the Service

```bash
go run cmd/server/main.go
```

## API Endpoints

### Connections
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/marketplace/connections` | List all connections |
| GET | `/admin/marketplace/connections/:id` | Get connection details |
| DELETE | `/admin/marketplace/connections/:id` | Disconnect marketplace |
| POST | `/admin/marketplace/connections/:id/refresh` | Refresh OAuth token |

### OAuth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/marketplace/:platform/auth-url` | Get OAuth URL |
| GET | `/admin/marketplace/:platform/callback` | OAuth callback |

### Products
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/marketplace/connections/:id/products` | List synced products |
| POST | `/admin/marketplace/connections/:id/products/push` | Push products |

### Categories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/marketplace/connections/:id/categories` | List mappings |
| GET | `/admin/marketplace/connections/:id/categories/external` | Get marketplace categories |
| POST | `/admin/marketplace/connections/:id/categories` | Create mapping |

### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/marketplace/connections/:id/orders` | List orders |
| POST | `/admin/marketplace/connections/:id/orders/sync` | Manual sync |
| PUT | `/admin/marketplace/connections/:id/orders/:id/status` | Update status |

### Inventory
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/marketplace/connections/:id/inventory/push` | Push stock |
| POST | `/admin/marketplace/connections/:id/inventory/status` | Get stock |

### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/webhooks/shopee` | Shopee webhook receiver |
| POST | `/api/v1/webhooks/tiktok` | TikTok webhook receiver |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `APP_PORT` | Service port (default: 8009) | No |
| `DB_HOST` | Database host | Yes |
| `DB_PORT` | Database port | Yes |
| `DB_USER` | Database user | Yes |
| `DB_PASSWORD` | Database password | Yes |
| `DB_NAME` | Database name | Yes |
| `NATS_URL` | NATS server URL | Yes |
| `JWT_SECRET` | JWT secret for auth | Yes |
| `SHOPEE_PARTNER_ID` | Shopee Partner ID | For Shopee |
| `SHOPEE_PARTNER_KEY` | Shopee Partner Key | For Shopee |
| `SHOPEE_SANDBOX` | Use sandbox API | No |
| `TIKTOK_APP_KEY` | TikTok App Key | For TikTok |
| `TIKTOK_APP_SECRET` | TikTok App Secret | For TikTok |
| `MARKETPLACE_ENCRYPTION_KEY` | 32-byte AES key | Yes |
| `SERVICE_CATALOG_URL` | Catalog service URL | Yes |
| `SERVICE_ORDER_URL` | Order service URL | Yes |

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│  Frontend Admin │────▶│ service-gateway │
└─────────────────┘     └────────┬────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │  service-marketplace   │
                    │                        │
                    │  ┌──────────────────┐  │
                    │  │ OAuth Handlers   │  │
                    │  │ Product Sync     │  │
                    │  │ Order Sync       │  │
                    │  │ Inventory Sync   │  │
                    │  │ Webhook Handler  │  │
                    │  └──────────────────┘  │
                    └───────────┬────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                 ▼
      ┌───────────┐     ┌───────────┐     ┌───────────┐
      │  Shopee   │     │  TikTok   │     │   NATS    │
      │    API    │     │    API    │     │  Events   │
      └───────────┘     └───────────┘     └───────────┘
```

## Testing

### Sandbox Testing

Both Shopee and TikTok provide sandbox environments:

- **Shopee**: Set `SHOPEE_SANDBOX=true`
- **TikTok**: Use test app credentials

### Manual Testing

```bash
# Get OAuth URL
curl -X POST http://localhost:8009/api/v1/admin/marketplace/shopee/auth-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"redirect_url": "http://localhost:3001/marketplace/callback"}'
```

## License

Proprietary - ecommerce Desa Murni Batik
