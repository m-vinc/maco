# Networking

Every VM has at least one network interface. A network mode decides what that
interface is connected to: the internet through your Mac, other VMs, or your
physical LAN.

The simplest option needs no setup. If you create a VM without a network, it
gets outbound internet access through user-mode NAT, the same way most laptops
reach the internet behind a home router. For anything more (VMs talking to each
other, or a VM appearing on your real network) you create a named network and
attach VMs to it:

```sh
maco networks new lab --mode bridge
maco vm new guest --network lab
```

## Which mode do I want?

| Mode | What the VM can reach | Setup needed | Use when |
|------|-----------------------|--------------|----------|
| `user` (default) | Outbound internet only. No inbound connections. | None. | You just need the VM online. |
| `switch` | Other VMs on the same switch. Not the LAN. | None. | You want a private network between VMs. |
| `bridge` | Your physical LAN, or a specific VLAN on it. | Admin rights. | The VM should appear as a real machine on your network. |
| `vmnet-bridged` | Your physical LAN, via Apple's vmnet. | Admin rights. | You prefer Apple's built-in bridging on a Wi-Fi or Ethernet port. |
| `vlan` | Nothing on its own. Building block for a bridge. | Admin rights. | You manage tagged VLAN interfaces separately. |

`user` and `switch` are self-contained and need no privileges. The bridge and
vmnet modes connect guests to real hardware and need administrator rights,
because maco has to configure host network interfaces.

## User NAT (default)

```sh
maco networks new outbound --mode user
maco vm new guest --network outbound
```

The guest gets an address and DNS by DHCP and reaches the internet through your
Mac, with outbound NAT. Nothing on your LAN can start a connection to the guest.
This mode changes no host interfaces and needs no admin rights.

Each interface gets its own private NAT, so two VMs on the same `user` network
cannot see each other. Use a `switch` for VM-to-VM traffic. You can also skip
creating a network and pass `--network user` directly.

## Isolated switch

```sh
maco networks new inside --mode switch
maco vm new a --network inside
maco vm new b --network inside
```

VMs on the same switch form a private Ethernet segment and can talk to each
other, but the switch has no connection to your LAN or the internet, and no DHCP
or router. Give the guests static addresses, or run those services on one of the
VMs.

## Native bridge (join your LAN or a VLAN)

```sh
# Plain bridge onto a physical port
maco networks new office --mode bridge --member en0
maco vm new guest --network office

# Bridge onto a tagged VLAN carried on a port
maco networks new vlan123 --mode bridge --vlan en0:123
maco vm new guest --network vlan123
```

A bridge puts the guest directly on a real Ethernet segment, so it appears as
its own machine on that network and gets its address from whatever DHCP or
router already runs there. `--member en0` bridges onto a physical port untagged.
`--vlan en0:123` puts the guest on VLAN 123 carried over `en0`: the guest sends
plain untagged frames and maco tags and untags them on the wire. Your physical
switch port must actually carry that VLAN.

A bridge provides no DHCP, DNS, routing, or NAT of its own. Add an optional host
address on the bridge with `--address 10.0.0.1/24`, or rely on the services
already on that network. Configuring host interfaces needs administrator rights.

If you prefer to manage the VLAN interface separately, create it on its own with
`--mode vlan --parent en0 --tag 123`, then add the resulting `vlanN` device to a
bridge with `--member vlanN`. A `vlan` network cannot be attached to a VM
directly; it only exists to be added to a bridge.

## vmnet bridged

```sh
maco networks new physical --mode vmnet-bridged --uplink en0
maco vm new guest --network physical
```

This uses Apple's own vmnet bridged backend on the named uplink, so the guest
joins your physical network and gets its address from it. vmnet decides which
interfaces it will accept; an interface simply existing is not enough. maco
checks the uplink exists when you create the network, and vmnet attaches when
the VM starts. `bridged` is accepted as an older spelling of this mode.

## How it works

For the native `bridge` and `vlan` modes, maco builds the connection on the host
so QEMU itself stays an ordinary unprivileged process. It creates a virtual
Ethernet pair, joins one end to a macOS bridge alongside your uplink or VLAN
interface, and carries the guest's frames to the other end through a small
packet helper. The helper needs privileges only briefly to open its capture
descriptor, then drops them. maco creates and cleans up these host interfaces
for you; the manifest you write stays the source of truth, and destroying a
network removes only what maco added.

`user` and `switch` modes need none of this: they are handled entirely inside
QEMU, which is why they require no admin rights.

## Changing a VM's interfaces while it runs

You can add, change, or remove a VM's interfaces without rebooting it, from the
VM interface editor in the web UI or the interfaces API. maco attaches or
detaches the device live and only saves the manifest once the change succeeds;
if it cannot confirm the result it reports that recovery is needed rather than
leaving the manifest out of step with the running VM.

Two caveats. Hotplug adds an Ethernet device but does not configure it inside
the guest, so set up DHCP or a static address on the new interface from within
the guest. And a VM created by an older version may need one stop and start to
gain the hotplug slots before live changes work.
