# Configuration

A VM is described by one YAML manifest at `<data-dir>/vms/<uuid>.yml`. The
manifest set is maco's source of truth; the SQLite cache never overrides it.

## VM manifest

```yaml
id: 20d20d55-0370-4beb-a89c-8cdc03b607d0
name: web1
image: ubuntu-24.04-arm64
cpus: 2
memory_mib: 2048
disk_size_gib: 20
network: user
username: maco
password: maco
autostart: false
```

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | uuid, assigned by `maco vm new`, names the manifest and disk directory |
| `name` | string | RFC 1123 hostname, also the guest hostname |
| `image` | string | catalog image key (see `maco image list`) |
| `cpus` | int | vCPUs, at least 1 |
| `memory_mib` | int | memory in MiB, at least 64 |
| `disk_size_gib` | int | virtual disk size for the qcow2 overlay, at least 1 |
| `network` | string | named network or UUID; empty/`user` retains user-mode NAT; `vmnet-host` and `vmnet-shared` remain available |
| `addresses` | list | optional static IPv4/IPv6 CIDRs for `lab0`; otherwise DHCPv4 |
| `username` | string | primary login user provisioned by cloud-init |
| `password` | string | optional console/ssh password |
| `ssh_key` | string | optional authorized SSH public key |
| `autostart` | bool | start this VM automatically on `maco reconcile` |

Manifests are written with `0600` permissions. Editing one by hand is
supported; the change takes effect the next time maco acts on that VM.

## Network manifest

A network is described by one YAML manifest at `<data-dir>/networks/<uuid>.yml`.
It is the desired topology; `maco net apply` reconciles the host to match.

```yaml
id: 214125b7-1ba8-47db-9796-708952310ed7
name: lab
mode: bridge
address: 10.100.0.1/24
members: []
vlans:
  - parent: en10
    tag: 123
```

`mode: bridge` creates a native macOS bridge. `members` lists existing host
interfaces to join; `vlans` creates native VLAN interfaces on Ethernet parents
and adds them as bridge members. Use separate bridges for separate access
VLANs. A plain bridge does not provide VLAN filtering between its ports.

`address` is an optional host IPv4 CIDR. Setting it does not create a DHCP
server, NAT, DNS, or routing service. Use `maco vm new --network lab --address
10.100.0.2/24` for a static guest or provide DHCP through an uplink.

`device` may initially name an existing bridge. Such a bridge is borrowed and
is retained on network destruction. Otherwise maco creates and owns a bridge.
Reconciliation records `owned`, `applied_members`, `applied_vlans`,
`applied_address`, and VLAN `device` values for cleanup; do not edit these
runtime fields. Existing member interfaces and VLAN parents are never destroyed.

Network manifests also retain `mode: switch` with `group` for the multicast
backend, and `mode: bridged` with `uplink` for Apple's physical vmnet bridge
backend. VM creation resolves a network name to its UUID so renaming a network
does not detach VMs created through the CLI.

## Autostart and state

A VM manifest does not record a desired power state. It carries only
`autostart`: when true, `maco reconcile` boots the VM if it is not already
running. `maco vm start`/`stop` remain the explicit controls.

The actual runtime state (phase, pid, boot time) is not stored in the manifest.
It lives in the SQLite cache at `<data-dir>/web.db` (`vm_state` table), written
on start/stop and refreshed from live process liveness by `maco vm list` and
`maco reconcile`. The cache is derivable: deleting `web.db` and reconciling
rebuilds it from the manifests and the running QEMU processes.
