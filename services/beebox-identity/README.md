# beebox-identity

Identity and authentication service for BeeBox.

## Deployment sequence

1. Provision PostgreSQL and set `DATABASE_URL`.
2. Apply schema migrations before starting the server:

```bash
cd services/beebox-identity
DATABASE_URL='postgres://...' go run ./cmd/migrate
```

Migrations live in `migrations/`. The migrate command records applied files in `schema_migrations` and is idempotent.

3. Start the service:

```bash
DATABASE_URL='postgres://...' go run ./cmd/server
```

The server does not apply migrations on boot. Missing schema causes runtime query failures.

## Configuration

See `env.example`.

- `DATABASE_URL` (required)
- SMTP settings for email verification and password reset delivery
- Twilio settings for phone verification SMS (`TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM_NUMBER`)

Missing SMS or SMTP configuration causes delivery attempts to fail. Delivery failures on verification and password-reset request paths do not change the public success response (enumeration-safe). Codes and reset tokens remain hashed at rest and expire normally.

## Password reset contract

Password reset requests require an email-shaped identifier. The delivery target is that email address. Public responses remain enumeration-safe (`{"status":"accepted"}` for both known and unknown accounts). Reset tokens are never returned over HTTP.

## Verification request contract

`POST /auth/verification/request` returns the same public success shape for existing and unknown accounts. Verification records and delivery occur only when the account exists.
