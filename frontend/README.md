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
| `/welcome` | Minimal authenticated shell |
| `/session-expired` | Session expired state |
| `/service-unavailable` | Connection failure state |

## Architecture

```
UI (React Server + Client Components)
        ↓
Next.js BFF (Route Handlers)
        ↓
BeeBox Identity API (server-side only)
```

- Opaque session token stored in **httpOnly / Secure / SameSite** cookie.
- Secrets never reach the browser.
- Errors mapped to human-friendly copy.

## Local setup

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Open http://localhost:3000.

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
