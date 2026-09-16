# Running multiple instances

maco coordinates through Redis and the shared data directory, so more than one
`maco serve` process can run against the same state (several processes on one
host, or processes on different hosts that share Redis and the data directory).

## Requirements

- All instances point at the same Redis (`MACO_REDIS_URL` or `--redis-url`).
- All instances share the same data directory (manifests, `web.db`, the JWT
  secret, and TLS material). Keep it on local storage or a filesystem with
  correct POSIX locking. SQLite over most network filesystems is not
  recommended.
- A VM runs on the host whose QEMU started it. The instance on that host manages
  its lifecycle; other instances observe state through the shared cache.

## State database

`web.db` is SQLite in WAL mode. Each process opens one pooled `*sql.DB`,
process-wide and keyed by the database path, so all requests and background work
in a process share the same bounded connection pool instead of opening a handle
per call. The connection string sets:

- `journal_mode=WAL`: readers and a single writer proceed concurrently.
- `busy_timeout=10000`: a blocked writer waits up to ten seconds for the write
  lock instead of failing immediately.
- `synchronous=NORMAL`: durable under WAL with less fsync overhead.
- `_txlock=immediate`: write transactions take the write lock up front, which
  avoids the deferred-to-write upgrade deadlock between two connections.

Writers serialize on the single SQLite write lock, within a process and across
processes, and retry within the busy timeout. Migrations run once per process on
first use and are idempotent; a file lock beside the database serializes the
first initialization so two processes starting against a brand-new database do
not run migrations at the same time.

## Job execution

Background operations run through an Asynq queue in Redis, scoped to the data
directory. Asynq delivers each job once across all workers, and each worker runs
one job at a time.

On top of that, a Redis lock serializes job execution across every instance, so
host mutations never overlap even when several servers are running. The lock is
a `SET NX` key with a random token, refreshed while held and released with a
compare-and-delete script. If the holder crashes, the key expires after its TTL
and another instance proceeds. This lock is coordination, not a fencing
guarantee: the engine's per-VM lifecycle locks and network-store locks remain
the correctness boundary on each host.

Preview capture is scheduled by every instance, but a per-VM Redis lease ensures
only one capture is enqueued per interval.

## Job history

Job metadata and logs live in Redis, independent of Asynq task retention. The
newest jobs stay in a history index; on each submission the index is trimmed to
its most recent entries, and only finished (succeeded or failed) jobs are
removed, so a pending or running operation is never dropped. Preview capture
jobs are private, kept out of the history index, and expire on their own so they
do not consume history. Use Redis AOF persistence and include Redis in backups;
`web.db` does not contain job history.
