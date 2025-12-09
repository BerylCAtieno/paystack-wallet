# Wallet Service with Paystack Integration

A production-ready wallet service backend built with Go, featuring Paystack integration, Google OAuth authentication, JWT tokens, and API key management.

## Features

- **Google OAuth Authentication** - Sign in with Google to generate JWT tokens
- **Wallet Management** - Create wallets, check balance, view transaction history
- **Paystack Integration** - Deposit funds using Paystack payment gateway
- **Wallet Transfers** - Transfer funds between users
- **API Key System** - Service-to-service authentication with permission-based access
- **Webhook Support** - Real-time transaction updates from Paystack
- **SQLite Database** - Lightweight, embedded database with WAL mode for concurrency

## Architecture

```
/wallet-service
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   ├── database/                   # Database connection & migrations
│   ├── domain/                     # Business logic
│   │   ├── auth/                   # API key & JWT logic
│   │   ├── user/                   # User models
│   │   └── wallet/                 # Wallet & transaction logic
│   ├── repository/                 # Database operations
│   ├── paystack/                   # Paystack client & webhooks
│   ├── api/                        # HTTP handlers & routing
│   │   ├── handlers/               # Request handlers
│   │   └── middleware/             # Authentication middleware
│   ├── security/                   # JWT & hashing utilities
│   └── utils/                      # Helper functions
├── go.mod
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- Paystack account (test or live keys)
- Google OAuth credentials

## Setup

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd wallet-service
go mod download
```

### 2. Configure Environment

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```env
PORT=8080
DB_PATH=./wallet.db

# Get from Google Cloud Console
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# Generate a secure random string
JWT_SECRET=your_jwt_secret_key

# Get from Paystack Dashboard
PAYSTACK_SECRET_KEY=sk_test_your_key
PAYSTACK_PUBLIC_KEY=pk_test_your_key
```

### 3. Get Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Add authorized redirect URI: `http://localhost:8080/auth/google/callback`
6. Copy Client ID and Client Secret to `.env`

### 4. Get Paystack Credentials

1. Sign up at [Paystack](https://paystack.com/)
2. Go to Settings → API Keys & Webhooks
3. Copy your test keys to `.env`

### 5. Run the Service

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Documentation

### Authentication

#### Google Sign-In
```
GET /auth/google
```
Redirects to Google OAuth consent screen.

#### Google Callback
```
GET /auth/google/callback
```
Handles OAuth callback and returns JWT token.

**Response:**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "user_id",
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

### API Key Management

All API key endpoints require JWT authentication.

#### Create API Key
```
POST /keys/create
Authorization: Bearer <jwt_token>

{
  "name": "wallet-service",
  "permissions": ["deposit", "transfer", "read"],
  "expiry": "1D"
}
```

**Expiry Options:** `1H`, `1D`, `1M`, `1Y`

**Permissions:**
- `deposit` - Can initiate deposits
- `transfer` - Can transfer funds
- `read` - Can view balance and transactions

**Response:**
```json
{
  "data": {
    "api_key": "sk_live_abc123...",
    "expires_at": "2025-12-10T10:00:00Z"
  }
}
```

⚠️ **Important:** Save the API key immediately - it won't be shown again!

#### Rollover Expired API Key
```
POST /keys/rollover
Authorization: Bearer <jwt_token>

{
  "expired_key_id": "key_id",
  "expiry": "1M"
}
```

Creates a new API key with same permissions as an expired key.

### Wallet Operations

Wallet endpoints accept either JWT or API Key authentication:
- **JWT:** `Authorization: Bearer <token>`
- **API Key:** `x-api-key: <api_key>`

#### Initiate Deposit
```
POST /wallet/deposit
Authorization: Bearer <jwt_token>
# OR
x-api-key: <api_key>

{
  "amount": 5000
}
```

**Response:**
```json
{
  "data": {
    "reference": "DEP_user_id_5000",
    "authorization_url": "https://checkout.paystack.com/..."
  }
}
```

User completes payment at `authorization_url`. Paystack sends webhook to credit wallet.

#### Get Balance
```
GET /wallet/balance
Authorization: Bearer <jwt_token>
# OR
x-api-key: <api_key>
```

**Requires:** `read` permission for API keys

**Response:**
```json
{
  "data": {
    "balance": 15000
  }
}
```

#### Transfer Funds
```
POST /wallet/transfer
Authorization: Bearer <jwt_token>
# OR
x-api-key: <api_key>

{
  "wallet_number": "4566678954356",
  "amount": 3000
}
```

**Requires:** `transfer` permission for API keys

**Response:**
```json
{
  "data": {
    "status": "success",
    "message": "Transfer completed"
  }
}
```

#### Get Transaction History
```
GET /wallet/transactions
Authorization: Bearer <jwt_token>
# OR
x-api-key: <api_key>
```

**Requires:** `read` permission for API keys

**Response:**
```json
{
  "data": [
    {
      "id": "txn_id",
      "type": "deposit",
      "amount": 5000,
      "status": "success",
      "reference": "DEP_...",
      "created_at": "2025-12-09T10:00:00Z"
    },
    {
      "id": "txn_id_2",
      "type": "transfer",
      "amount": -3000,
      "status": "success",
      "recipient_wallet": "4566678954356",
      "created_at": "2025-12-09T11:00:00Z"
    }
  ]
}
```

### Webhooks

#### Paystack Webhook
```
POST /wallet/paystack/webhook
```

Receives payment notifications from Paystack. Validates signature and credits wallet on successful payment.

**Configure in Paystack Dashboard:**
1. Go to Settings → API Keys & Webhooks
2. Add webhook URL: `https://your-domain.com/wallet/paystack/webhook`
3. Select events: `charge.success`

#### Verify Deposit Status
```
GET /wallet/deposit/{reference}/status?reference=DEP_...
```

Manual verification endpoint (for debugging). Does NOT credit wallet - only webhook credits wallets.

## Testing

### Manual Testing Flow

1. **Sign in with Google**
   ```bash
   # Open in browser
   http://localhost:8080/auth/google
   ```

2. **Create API Key**
   ```bash
   curl -X POST http://localhost:8080/keys/create \
     -H "Authorization: Bearer <your_jwt>" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "test-key",
       "permissions": ["deposit", "transfer", "read"],
       "expiry": "1D"
     }'
   ```

3. **Initiate Deposit**
   ```bash
   curl -X POST http://localhost:8080/wallet/deposit \
     -H "x-api-key: <your_api_key>" \
     -H "Content-Type: application/json" \
     -d '{"amount": 10000}'
   ```

4. **Check Balance**
   ```bash
   curl http://localhost:8080/wallet/balance \
     -H "x-api-key: <your_api_key>"
   ```

5. **Transfer Funds**
   ```bash
   curl -X POST http://localhost:8080/wallet/transfer \
     -H "x-api-key: <your_api_key>" \
     -H "Content-Type: application/json" \
     -d '{
       "wallet_number": "recipient_wallet_number",
       "amount": 5000
     }'
   ```

### Testing Webhooks Locally

Use ngrok to expose your local server:

```bash
ngrok http 8080
```

Add the ngrok URL to Paystack webhook settings:
```
https://your-ngrok-url.ngrok.io/wallet/paystack/webhook
```




TODO: Fix migration error: 2025/12/09 22:30:26 Failed to run migrations: failed to execute migration internal/database/migrations/001_create_users.down.sql: FOREIGN KEY constraint failed

TODO: Test user registration with Google