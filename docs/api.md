# API contract

Maco uses a code-first workflow: Swag annotations on the Go HTTP
handlers generate OpenAPI 3.1, and openapi-typescript generates the frontend
contract. The frontend uses openapi-fetch with that generated contract for
JSON requests, PNG previews, and multipart uploads. API models in
`web/src/api.ts` are aliases of generated Go schemas. UI form types may make
optional API fields required locally to represent initialized controls.

The running server exposes the spec at `/api/docs/openapi.json`, without
requiring login. Obtain a JWT with `POST /api/login` and send
`Authorization: Bearer <token>` on protected HTTP requests. A `202` response
contains a queued job, not the completed result; follow `/api/jobs/{id}` for
its final state and logs.

Console, display, and notification routes upgrade to WebSocket. Send a JSON
object containing `token` as the first frame within five seconds. These
transports use WebSocket consumers rather than the HTTP client. Notification
messages follow `JobsEvent` in `pkg/api/job_stream.go`.

Install generators and frontend dependencies, then regenerate:

```sh
make api-tools
npm ci --prefix web
make api-generate
make api-check
npm --prefix web run build
```

Commit `pkg/api/docs/swagger.json` and the generated `web/src/api/generated/`
client with handler changes. `make api-check` regenerates into a temporary directory and fails if
either committed artifact differs; it also checks every mounted API operation
against the spec in both directions. Run this check in CI alongside API tests.

The spec is generated from Go handler annotations and Go request/response
structs by the Swag rc5 command. `binding:"optional"`
tags declare optional fields in the source; no script maintains schema field
lists. `make docs` is an alias for the full generation workflow.

`tools/openapi/normalize.py` repairs Swag rc5's malformed multipart file output
using the field name and required flag generated from the handler's `@Param`
annotation. It contains no API paths, model names, or field names.

Upload progress uses an XMLHttpRequest fetch adapter behind the same typed
client, authentication middleware, and error handling as other requests.
