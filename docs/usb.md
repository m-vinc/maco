# USB devices

In VM details, open **USB devices → Add device**. Select a device by name and serial/port. Technical information is under **Device details**.

- **Stopped VM:** the device is saved with an **On startup** badge and connects on every start. **Remove** clears the assignment.
- **Running VM:** the device connects immediately for the current session. **Detach** releases it. Removing a saved assignment also disconnects its live attachment.

Keep saved devices plugged in when starting the VM. Search and refresh are available in the picker.

Attach and detach are background jobs. A successful attach means QEMU reports the device attached to its virtual USB bus; guest drivers and useful device I/O still depend on the device and guest. Eject or unmount storage inside the guest before detaching it. Whole composite devices move together. Host access can be interrupted while a device is assigned.

## Requirements

- Build on macOS with cgo, pkg-config and libusb **1.0.30 or newer**. Install the build dependencies with `brew install pkg-config libusb` if absent. The binary links to libusb, so the library must also be present at runtime.
- QEMU must expose `usb-host`. The inventory API reports whether the installed binary supports it.
- Stop/start VMs created by older binaries once to get the named USB controller. The usual emulated keyboard and tablet remain enabled.
- maco normally runs as root. Host drivers or mounted storage can prevent capture even with root access. maco does not unmount filesystems itself. QEMU/libusb may attempt to release host interfaces as part of capture.

## CLI

```sh
maco usb list
maco usb list --json
maco vm usb list my-vm
maco vm usb attach my-vm <device-id-from-list>
maco vm usb detach my-vm <attachment-id>
maco vm usb assign my-vm <device-id-from-list>
maco vm usb unassign my-vm <vendor:product[:serial]>
maco vm usb assigned my-vm
```

Device IDs identify the current physical connection. Unplugging/reconnecting can invalidate a previously listed ID. The UI also submits the expected fingerprint; the worker rechecks it immediately before and after capture. The CLI resolves the current fingerprint from the selected connection ID.

`attach`/`detach` are session-only. `assign`/`unassign` edit the manifest so a device reattaches on every start. `assign` records the serial plus vendor and product IDs of the selected device; it never stores the transient bus/address. `assigned` prints the manifest entries and the key `unassign` expects.

## API

All routes use the usual bearer authentication.

| Method and path | Result |
| --- | --- |
| `GET /api/usb/devices` | `devices`, `supported`, optional `reason` |
| `GET /api/vms/{id}/usb` | Saved and live assignments, with `assignment_key` for saved entries |
| `POST /api/vms/{id}/usb` | Body: `device_id`, `fingerprint`; returns job with HTTP 202 |
| `DELETE /api/vms/{id}/usb/{attachment-id}` | Returns detach job with HTTP 202 |
| `POST /api/vms/{id}/usb/assignments` | Body: `device_id`, `fingerprint`; queues a saved assignment, including while stopped |
| `DELETE /api/vms/{id}/usb/assignments/{key}` | Removes a saved assignment and any matching live attachment |

The mutation jobs are `vm.usb.attach`, `vm.usb.detach`, `vm.usb.assign` and `vm.usb.unassign`. Validation failures return before queuing; changes between submission and execution fail the job. The UI reports job results through the existing jobs stream.

## Persistent assignments and boot policy

A manifest may list `usb:` assignments, each identified by vendor and product IDs plus an optional serial. On every start maco resolves each assignment against the live host inventory **before launching QEMU**:

- If an assigned device is not connected, matches more than one connected device, is unavailable, or is already attached to another VM, the start is refused and QEMU is not launched. Fix the device situation and start again.
- When all assignments resolve, maco starts the VM and captures each device. If a capture fails after launch, maco stops the VM rather than leaving it running without a required device.

Manifest assignments never store the transient bus/address; matching uses serial plus vendor/product IDs. Use a serial to disambiguate two identical models. Manage assignments with `maco vm usb assign|unassign|assigned` or by editing the manifest.

## Lifetime and ownership

Ad-hoc `attach`/`detach` assignments are session-only. VM stop releases the device. Manifest `usb:` assignments are reattached on the next start; ad-hoc attachments are not. Backend restart preserves assignments to QEMU processes that are still running. Runtime records live in the private `/tmp/maco-usb-<uid>/claims.json` registry, shared by maco instances running as the same user, including different data directories. The normal root service therefore shares one registry. Separate operating-system users do not share this application registry; libusb/host ownership restrictions still apply.

A global registry lock precedes the VM lifecycle lock for USB mutations. Records are written before capture, inspected against the QEMU object and PID-file generation, and removed only after confirmed device deletion or VM exit. If QEMU cannot be inspected, ownership remains reserved and the UI displays **Check connection**. Do not delete the registry while assigned VMs are running. Reads reconcile stale records after stop/exit.

QEMU capture is checked with the USB object's `attached` property. Detach waits until the QEMU object is absent, handles guest unplug errors, and bounds the wait. A command acknowledgement by itself is not treated as removal completion.

## Physical disconnect and limits

Physical unplug does **not** remove the assignment and does **not** stop the VM. maco takes no action on unplug; the guest observes an ordinary USB disconnect and its behavior is undefined. QEMU keeps the object with `attached` false, and the assignment shows **Disconnected** until you detach it. QEMU can reconnect a device matching the same bus, address, port and vendor/product IDs while the object exists. Stock QEMU does not use the host serial number or libusb connection ID as a capture filter. Detach before replacing a device, including one of the same model in the same port. Refresh or reopen the USB tab to observe changes.

Inventory serial numbers can be unavailable when the host will not open the device. Hubs and devices without precise location/connection identity are ineligible. Device-specific host-driver restrictions are reported during capture. No automatic host storage unmount or guest filesystem management is provided.

## Limits

USB passthrough is validated against a QEMU virtual USB bus and an automated
test suite; it has not been through broad physical-hardware acceptance across
device classes. A successful attach means QEMU holds the device on the guest's
USB bus. Whether a given peripheral then works depends on the guest's own
drivers. Serial and storage transfers and hardware-token authentication in
particular should be confirmed with your actual device.
