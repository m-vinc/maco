# Authentication and roles

The web UI and the HTTP API share one account system. Accounts live in the
SQLite database (`users` table) with bcrypt password hashes, independent of unix
users. The CLI, API, and UI all authenticate against the same table.

## Accounts

```bash
maco user add alice --password secret --role admin
maco user add bob   --password secret --role viewer
maco user list
maco user passwd alice --password new
maco user rm bob
```

On first run `maco serve` (and `maco install`) create an `admin` account from
`MACO_ADMIN_PASSWORD`, or a generated password logged once. A JWT signing secret
is created at `<data-dir>/jwt.secret` (0600).

## Roles

Two roles exist today. `--role` defaults to `admin` and only accepts a known
role.

| Role | Capabilities |
|------|--------------|
| `admin` | Full access: create, edit, start, stop, and delete VMs, networks, disks, media, and USB assignments; open the interactive graphical and serial consoles. |
| `viewer` | Read-only: list and inspect every resource, watch job and event streams, and see VM previews. Viewers cannot mutate anything and cannot open the interactive consoles. |

Roles are simple strings validated by `auth.ValidRole`. `admin` and `viewer`
are the only valid values; more roles can be added later (see
[Extending roles](#extending-roles)).

## Enforcement

Login (`POST /api/login`) returns a JWT valid for 24 hours. The token carries the
user id, a credential version, and the role. The browser sends it as an
`Authorization: Bearer` header; WebSocket endpoints receive it as the first
message.

Enforcement is authoritative on the server and does not trust the role embedded
in the token:

- Every authenticated request reloads the user from the database and takes the
  role from that row, so a role change takes effect immediately without a new
  login.
- Read-only HTTP methods (`GET`, `HEAD`) are allowed for any authenticated
  account. Mutating methods (`POST`, `PUT`, `PATCH`, `DELETE`) require `admin`
  and otherwise return `403`.
- The interactive console and display WebSockets require `admin`, since they
  send input to the guest. The job and event notification streams are read-only
  and allowed for any role.

A password change rotates the credential version and invalidates existing
tokens. A deleted or renamed account is rejected on the next request.

## Current user

`GET /api/me` returns `{ "username": ..., "role": ... }` for the authenticated
account. The UI calls it once after login and hides the actions a viewer cannot
perform (create buttons, row actions, interactive consoles, and the VM editing
tabs). The hidden controls are a convenience: the server still enforces every
rule, so a viewer that reaches a mutating endpoint directly receives `403`.

## Sessions

The browser stores the JWT in `sessionStorage`, so it is scoped to the tab and
cleared when the tab closes. A `401` from any API call clears the token and
redirects to the login page. Tokens are never placed in URLs; the console and
event WebSockets authenticate with the token in their first message and enforce
a same-origin check.

## Extending roles

The role is a plain string on the account and in the JWT. To add a role, add a
constant in `pkg/auth`, extend `auth.ValidRole`, and adjust the capability check
(`auth.CanMutate`, or a richer capability map if finer permissions are needed).
The `/api/me` contract already exposes the role to the UI.

Because identity is a signed token with a role claim, an external identity
provider (OIDC) can be added later by mapping provider claims to a maco role and
issuing the same token shape, without changing the API surface.
