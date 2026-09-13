# BeeBox Frontend — End-user Auth MVP

Next.js application providing the end-user authentication experience for applications powered by BeeBox Identity.

## Product feel

Clerk-level simplicity × Linear-level polish × BeeBox box/layer identity.

This is **end-user UI**. Users never see project IDs, session tokens, internal error codes, or infrastructure concepts.

## Screens

| Route | Purpose |
|-------|---------|
| `/sign-in` | Email + password sign-in |
| `/sign-up` | Create account |
| `/forgot-password` | Request password reset |
| `/reset-password?reset_id=…&token=…` | Set new password |
| `/password-updated` | Confirmation after reset |
| `/welcome` | Minimal authenticated shell (**protected by middleware**) |
| `/session-expired` | Session expired state |
| `/service-unavailable` | Connection failure state |

## Architecture

```
UI (React Server + Client Components)
        ↓
Next.js Middleware (cookie gate for /welcome)
        ↓
Next.js BFF (Route Handlers)
        ↓
BeeBox Identity API (server-side only)
```

- Opaque session token stored in **httpOnly / Secure / SameSite** cookie.
- Secrets never reach the browser.
- Errors mapped to human-friendly copy.

## Local setup

### 1. Start backend services

You need at least **beebox-identity** running (port 8081 by default).

```bash
# Terminal 1 – Identity
cd services/beebox-identity
# set env from your config (PORT=8081, DATABASE_URL, VERIFICATION_CODE_SECRET, ...)
go run ./cmd/server
```

Optional: also run beebox-project (8080) and beebox-runtime (8084) if you need the full stack.

### 2. Start frontend

```bash
cd frontend
cp .env.example .env.local
# Edit BEEBOX_IDENTITY_URL if needed (default http://127.0.0.1:8081)

npm install
npm run dev
```

Open http://localhost:3000.

### 3. Test flows

1. **Sign up** → `/sign-up` → create account → redirected to sign-in
2. **Sign in** → `/sign-in` → success → `/welcome` (protected)
3. Open `/welcome` in a private window → should redirect to `/sign-in`
4. **Sign out** from welcome → back to sign-in, cookie cleared
5. **Forgot password** → request reset (email delivery depends on SMTP config)

## Environment

See `.env.example`. All BeeBox credentials are **server-only**.

## Scope

Implements only the flows in the current Identity OpenAPI:
- POST /auth/signup
- POST /auth/signin
- GET /auth/session
- POST /auth/signout
- POST /auth/password-reset/request
- POST /auth/password-reset/reset

No social login, magic links, passkeys, or MFA.
