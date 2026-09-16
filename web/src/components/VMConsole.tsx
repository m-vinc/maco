import { lazy, Suspense, useState } from 'react'
import { Button } from 'cheval-ui'
import { Monitor, Terminal } from 'lucide-react'

const SerialConsole = lazy(loadSerial)
const GraphicalConsole = lazy(loadGraphical)

async function loadSerial() {
  const module = await import('./SerialConsole')
  return { default: module.SerialConsole }
}

async function loadGraphical() {
  const module = await import('./GraphicalConsole')
  return { default: module.GraphicalConsole }
}

export function VMConsole({ id }: { id: string }) {
  const [mode, setMode] = useState<'serial' | 'graphical'>('serial')

  return (
    <div className="space-y-4">
      <div role="group" aria-label="Console type" className="flex gap-2">
        <Button
          variant={mode === 'serial' ? 'default' : 'ghost'}
          aria-pressed={mode === 'serial'}
          onClick={() => setMode('serial')}
        >
          <Terminal aria-hidden="true" className="mr-2 h-4 w-4" />
          Serial
        </Button>
        <Button
          variant={mode === 'graphical' ? 'default' : 'ghost'}
          aria-pressed={mode === 'graphical'}
          onClick={() => setMode('graphical')}
        >
          <Monitor aria-hidden="true" className="mr-2 h-4 w-4" />
          Graphical
        </Button>
      </div>
      <Suspense fallback={<p role="status">Loading console…</p>}>
        {mode === 'serial' ? (
          <SerialConsole id={id} />
        ) : (
          <GraphicalConsole id={id} />
        )}
      </Suspense>
    </div>
  )
}
