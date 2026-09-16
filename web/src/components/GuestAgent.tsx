import { useCallback } from 'react'
import { Button, CopyButton } from 'cheval-ui'
import { getGuestAgent, type GuestAgent as GuestAgentData } from '../api'
import { useResource } from '../hooks/useResource'
import { ResourceNotice } from './ResourceNotice'
import type { ReactNode } from 'react'

interface GuestAgentProps {
  id: string
}

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`
}

function Row({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="grid grid-cols-[7rem_minmax(0,1fr)] gap-3 py-2 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="break-words font-medium">{value}</dd>
    </div>
  )
}

export function GuestAgent({ id }: GuestAgentProps) {
  const load = useCallback(() => getGuestAgent(id), [id])
  const agent = useResource<GuestAgentData | null>(load, null, `guest-agent:${id}`)
  const data = agent.data
  const os = data?.os
  const interfaces = (data?.interfaces || []).filter(
    (nic) => nic.name !== 'lo' && (nic.ip_addresses?.length || 0) > 0,
  )
  const filesystems = (data?.filesystems || []).filter((fs) => fs.total_bytes > 0)

  return (
    <section
      className="rounded-xl border bg-card p-4"
      aria-label="Guest agent"
    >
      <div className="mb-2 flex items-center gap-2">
        <h2 className="flex-1 font-semibold">Guest Information</h2>
        {data?.available && data.version && (
          <span className="text-xs text-muted-foreground">
            qemu-guest-agent {data.version}
          </span>
        )}
        <Button size="sm" variant="outline" onClick={agent.refresh}>Refresh</Button>
      </div>

      <ResourceNotice resource={agent} name="guest information" />

      {!agent.loading && data && !data.available && (
        <p className="text-sm text-muted-foreground">
          The QEMU guest agent is not responding. Install and start
          qemu-guest-agent inside the guest to see OS details, addresses, and
          filesystem usage here. Supported distribution images attempt to install it during automatic setup; custom images may need manual installation.
        </p>
      )}

      {data?.available && (
        <div className="space-y-5">
          <dl className="divide-y divide-border">
            {data.hostname && <Row label="Hostname" value={data.hostname} />}
            {os?.pretty_name && <Row label="OS" value={os.pretty_name} />}
            {os?.kernel_release && (
              <Row
                label="Kernel"
                value={`${os.kernel_release}${os.machine ? ` (${os.machine})` : ''}`}
              />
            )}
          </dl>

          {interfaces.length > 0 && (
            <div>
              <h3 className="mb-1 text-sm font-semibold text-muted-foreground">
                Network
              </h3>
              <dl className="divide-y divide-border">
                {interfaces.map((nic) => (
                  <Row
                    key={nic.name}
                    label={nic.name}
                    value={
                      <div className="space-y-1">
                        {(nic.ip_addresses || []).map((addr) => (
                          <span key={addr.address} className="flex flex-wrap items-center gap-2">
                            <span className="break-all">
                              {addr.address}/{addr.prefix}
                            </span>
                            <CopyButton text={addr.address} label={`Copy IP address ${addr.address}`} />
                          </span>
                        ))}
                      </div>
                    }
                  />
                ))}
              </dl>
            </div>
          )}

          {!interfaces.length && (
            <p className="text-sm text-muted-foreground">
              No guest network addresses reported. Use the console to check guest connectivity.
            </p>
          )}
          {filesystems.length > 0 && (
            <div>
              <h3 className="mb-1 text-sm font-semibold text-muted-foreground">
                Filesystems
              </h3>
              <dl className="divide-y divide-border">
                {filesystems.map((fs) => (
                  <Row
                    key={fs.mountpoint}
                    label={fs.mountpoint}
                    value={`${fs.type} · ${formatBytes(fs.used_bytes)} used of ${formatBytes(fs.total_bytes)}`}
                  />
                ))}
              </dl>
            </div>
          )}
        </div>
      )}
    </section>
  )
}
