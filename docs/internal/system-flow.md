# UI, API, QEMU and filesystem flow

Current architecture, verified against the implementation on 2026-09-14.

```mermaid
flowchart TD
    UI["Browser UI<br/>React"]
    API["HTTP API<br/>JWT authentication"]
    Redis[("Redis<br/>Job queue, status and logs")]
    Worker["Asynq worker<br/>One operation at a time"]
    Engine["Engine<br/>VM and network orchestration"]
    CLI["CLI"]
    Driver["VM driver"]
    QEMU["QEMU process<br/>One per running VM"]
    Guest["Guest OS<br/>Applications + guest filesystem"]
    HVF["macOS Hypervisor.framework"]

    UI <-->|"HTTP requests / JSON responses"| API
    API -->|"Mutations: enqueue job; return 202 + job ID"| Redis
    Redis --> Worker
    Worker --> Engine
    Worker -->|"Progress and result"| Redis
    Redis -->|"Job updates via API WebSocket"| UI
    API <-->|"Immediate reads"| Engine
    CLI --> Engine
    Engine --> Driver
    Driver <-->|"Launch / PID checks / QMP socket"| QEMU
    QEMU --- HVF
    QEMU --- Guest

    subgraph FS["Host filesystem"]
        YAML[("YAML manifests<br/>VM + network configuration")]
        Assets[("Base images<br/>Firmware + cloud-init seed")]
        Disks[("QCOW2 files<br/>Boot + data disks")]
        DB[("SQLite · web.db<br/>Runtime cache + users")]
    end

    Engine <-->|"Read / save configuration"| YAML
    Engine -->|"Prepare on start"| Assets
    Engine -->|"Create / manage"| Disks
    Engine <-->|"Refresh runtime state"| DB
    Assets -->|"Read"| QEMU
    QEMU <-->|"Virtual disk I/O"| Disks
```

## Live console path

```mermaid
flowchart LR
    UI["Browser<br/>xterm.js / noVNC"]
    API["API<br/>Authenticated WebSocket proxy"]
    S["Unix sockets<br/>console.sock / vnc.sock"]
    Q["QEMU"]
    G["Guest<br/>Serial terminal / display"]
    UI <-->|"WebSocket"| API
    API <--> S
    S <--> Q
    Q <--> G
```

- **Start:** API queues a job; worker calls the engine; engine reads YAML and prepares disks; driver launches QEMU; job updates reach the UI.
- **Guest file write:** guest filesystem → virtual disk controller → QEMU → host QCOW2 file. The UI/API are outside this data path.
- **Persistence:** YAML stores configuration; SQLite caches VM runtime state and stores users; Redis stores jobs and their logs. Only the runtime cache portion of SQLite is rebuildable.
- **Runtime files:** QMP, serial and VNC sockets, PID files, logs and previews live under `/tmp/maco-<hash>/<uuid>/`.

Implementation: `pkg/engine/vm.go`, `pkg/jobs/jobs.go`, `pkg/vm/qemu.go`, `pkg/api/console.go`. See [architecture](../architecture.md) and the proposed [USB passthrough plan](usb-passthrough-plan.md).
