import type { VMManifest } from './api'

export function bootOrder(manifest: VMManifest, isos = manifest.isos || []): string[] {
  const devices = ['disk', ...(manifest.disks || []).map(disk => `disk:${disk.id}`), ...isos.map(id => `iso:${id}`)]
  const preferred = manifest.boot_order?.length
    ? manifest.boot_order
    : [...isos.map(id => `iso:${id}`), 'disk']
  return [...new Set([...preferred, ...devices])].filter(id => devices.includes(id))
}
