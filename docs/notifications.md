# UI notifications

## Transport choice

Use the existing `github.com/coder/websocket` Go implementation with the
browser's native WebSocket client. It integrates with our Chi/net/http server,
JWT authentication, Redis pub/sub, and Asynq jobs. The application notification
layer supplies named events, reconnect backoff, heartbeat, and snapshot recovery.
This provides the notification behavior we need without introducing a separate
Socket.IO/Engine.IO protocol or HTTP long-polling fallback.

This endpoint uses our JSON event protocol and is not wire-compatible with
`socket.io-client`. Consumers must use a standard WebSocket client. Resource
notifications trigger targeted HTTP snapshot reads; images remain HTTP assets.

## Protocol

The authenticated `/api/events` WebSocket is the shared notification connection
for the application. Send `{"token":"JWT"}` as the first message. Authentication
failures close with code 1008. The legacy `/api/jobs/stream` endpoint remains
available for job-only clients.

Messages:

- `snapshot`: initial public job state. Sent again on each connection, allowing
  recovery after a disconnect.
- `update`: a public job, including its state and logs.
- `invalidate`: a `resources` array identifying changed domains (`vms`,
  `networks`, `media`, `disks`, `storage`, `interfaces`, `usb`), or a specific
  `preview:VM_UUID`. `all` requests a full refresh.

Resource hooks fetch an initial HTTP snapshot, then refresh only for matching
notifications. Concurrent loads of the same shared loader are deduplicated.
Changes received during an in-flight request trigger a follow-up refresh, so a
late response cannot swallow a newer change. Preview images use the existing
HTTP image endpoint, fetched on screenshot completion rather than a timer.
Private screenshot jobs remain hidden from public job history and updates.

Job transitions are delivered through Redis pub/sub. Terminal mutations signal
the affected domains. Direct media mutations signal resource changes as well.
For changes originating outside the API (CLI manifest edits, VM process exits,
USB hotplug, and filesystem usage), one shared server observer samples state
while notification clients are connected and publishes only changed domains.
This observer still uses a two-second timer; browser HTTP polling is removed.
It stops when the final client disconnects. Slow clients receive `all` instead
of silently missing resource changes. WebSocket ping frames keep connections
healthy without repeating job snapshots. The browser reconnects with backoff
and refreshes mounted resources after the new snapshot.

Network creation remains a durable API job. Its execution applies the network
before marking the job successful. The creation dialog waits for terminal state
through the same notification connection; apply failures stay in the dialog.
