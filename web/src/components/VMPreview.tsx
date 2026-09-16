import { useEffect, useState } from 'react'
import { Monitor, Terminal } from 'lucide-react'
import { getVMPreview } from '../api'
import { subscribeResourceEvents } from '../hooks/resourceEvents'
import { type ConsoleMode } from './VMConsoleDialog'

interface VMPreviewProps {
  id: string
  name: string
  running: boolean
  bootTime: number
  large?: boolean
  onOpenConsole?: (mode: ConsoleMode) => void
}

interface ConsoleButtonProps {
  label: string
  icon: typeof Monitor
  large?: boolean
  onClick: () => void
}

function ConsoleButton({ label, icon: Icon, large, onClick }: ConsoleButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className="flex flex-col items-center gap-1 rounded-lg bg-background/90 px-3 py-2 text-xs font-medium text-foreground shadow-sm ring-1 ring-border transition hover:bg-background focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
    >
      <Icon aria-hidden="true" className={large ? 'h-6 w-6' : 'h-4 w-4'} />
      {large && <span>{label}</span>}
    </button>
  )
}

export function VMPreview({
  id,
  name,
  running,
  bootTime,
  large,
  onOpenConsole,
}: VMPreviewProps) {
  const [image, setImage] = useState('')
  const [imageBoot, setImageBoot] = useState(0)
  const [ratio, setRatio] = useState(8 / 5)

  function subscribe() {
    setImage('')
    setRatio(8 / 5)
    if (!running || bootTime <= 0) return

    const controller = new AbortController()
    let current = ''
    let loading = false
    let dirty = false

    async function refresh() {
      if (loading) { dirty = true; return }
      dirty = false

      loading = true
      try {
        const { data: blob, response } = await getVMPreview(id, controller.signal)
        if (!response?.ok || !blob) return

        const captureHeader = response.headers.get('X-Preview-Captured-At')
        const capturedAt = captureHeader
          ? Number(captureHeader)
          : Date.parse(response.headers.get('Last-Modified') || '')
        if (!Number.isFinite(capturedAt) || capturedAt < bootTime * 1000) return

        if (controller.signal.aborted) return

        const next = URL.createObjectURL(blob)
        if (current) URL.revokeObjectURL(current)
        current = next
        setImage(next)
        setImageBoot(bootTime)
      } catch {
        return
      } finally {
        loading = false
        if (!controller.signal.aborted && dirty) void refresh()
      }
    }

    const unsubscribe = subscribeResourceEvents(resources => {
      if (resources.includes('all') || resources.includes(`preview:${id}`)) void refresh()
    })
    void refresh()
    return () => {
      controller.abort()
      unsubscribe()
      if (current) URL.revokeObjectURL(current)
    }
  }

  useEffect(subscribe, [id, running, bootTime])

  const consoles = running && onOpenConsole

  return (
    <div
      className={`group relative flex items-center justify-center overflow-hidden rounded-lg bg-muted ring-1 ring-border ${large ? 'w-full max-w-[22rem]' : 'w-28'}`}
      style={{ aspectRatio: running ? ratio : 8 / 5 }}
      title={running ? 'Latest captured VM display' : 'VM is stopped'}
    >
      {running && image && imageBoot === bootTime ? (
        <img
          src={image}
          alt={`${name} display preview`}
          className="absolute inset-0 h-full w-full rounded-lg object-contain"
          onLoad={(event) => {
            const image = event.currentTarget
            if (image.naturalHeight)
              setRatio(image.naturalWidth / image.naturalHeight)
          }}
        />
      ) : (
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <Monitor
            aria-label={running ? 'Preview not available yet' : 'VM is stopped'}
            className={large ? 'h-8 w-8' : 'h-5 w-5'}
          />
          {large && (
            <p className="text-sm">
              {running ? 'No screenshot available yet' : 'VM is stopped'}
            </p>
          )}
        </div>
      )}
      {consoles && (
        <div className={`absolute inset-0 flex items-center justify-center bg-background/60 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100 ${large ? 'gap-4' : 'gap-2'}`}>
          <ConsoleButton
            label="Graphical console"
            icon={Monitor}
            large={large}
            onClick={() => onOpenConsole('graphical')}
          />
          <ConsoleButton
            label="Serial console"
            icon={Terminal}
            large={large}
            onClick={() => onOpenConsole('serial')}
          />
        </div>
      )}
    </div>
  )
}
