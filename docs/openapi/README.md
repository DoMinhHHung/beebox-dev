# OpenAPI Convention

This repository documents each HTTP service with its own OpenAPI 3.1 contract.

The source code remains the runtime truth. These files describe the current HTTP surface as implemented today and do not change service behavior.

## Locations

- `services/beebox-identity/docs/openapi.yaml`
- `services/beebox-project/docs/openapi.yaml`
- `services/beebox-runtime/docs/openapi.yaml`

Only services that actually expose HTTP endpoints get an OpenAPI document.

## Naming

- Use OpenAPI 3.1.x.
- Keep operationIds in lower camel case.
- Keep reusable schema names in Upper Camel Case.
- Use service-specific tags that reflect the resource or concern being documented.
- Keep JSON property names exact and match the wire format from source code.
- Preserve required versus optional fields exactly as implemented.
- Model enums only when the source code constrains the values.

## Security Schemes

Name security schemes by the credential semantics, not just the transport.

- Identity session bearer tokens stay separate from project developer bearer tokens.
- Internal service bearer tokens stay separate from end-user bearer tokens.
- Header-based project credentials should use an apiKey security scheme with the actual header name.

If one operation requires more than one credential, represent that with a single security requirement object containing all required schemes.

## Error Schema

All service OpenAPI documents use the same error envelope shape:

- `error.code`
- `error.message`

Document the actual status codes each service returns. Do not collapse service-specific behavior into a fake shared taxonomy.

## Validation

The OpenAPI files are static documentation. They should be checked without adding runtime dependencies or middleware.

Typical local checks are:

```bash
python3 - <<'PY'
import yaml
from pathlib import Path
for path in Path('services').glob('*/docs/openapi.yaml'):
    yaml.safe_load(path.read_text())
    print(path)
PY
```

```bash
go test ./...
```

Use your local editor or any installed OpenAPI linter/viewer to inspect the YAML. No Swagger UI or runtime OpenAPI middleware is required in the Go services.
