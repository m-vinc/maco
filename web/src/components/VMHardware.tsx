import type { FormEvent } from 'react'
import { Button, NoticeBanner, PreferencesGroup } from 'cheval-ui'
import { getHost, updateHardware, type VMView, type Job } from '../api'
import { useRowAction } from '../hooks/useRowAction'
import { useResource } from '../hooks/useResource'
import { useDraft } from '../hooks/useDraft'
import { IntegerEntryRow } from './IntegerEntryRow'
import { MemoryEntryRow } from './MemoryEntryRow'
import { ActionStatus } from './ActionStatus'
import { ResourceNotice } from './ResourceNotice'
import { memoryLabel } from '../ux'

export function VMHardware({ vm, run }: { vm: VMView; run: (action: () => Promise<Job>) => Promise<Job | null> }) {
  const manifest = vm.manifest
  const draft = useDraft(`maco:draft:${manifest.id}:compute`, {
    cpus: manifest.cpus,
    memory_mib: manifest.memory_mib,
  })
  const host = useResource(getHost, null, 'host')
  const action = useRowAction(manifest.id, manifest.name, 'vm', run)
  const blocked = vm.phase === 'running' || action.busy

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (blocked || !draft.dirty || draft.conflict) return
    await action.execute('vm.hardware', () => updateHardware(manifest.id, draft.value))
  }

  const hostMemoryMib = host.data ? host.data.memory_bytes / (1024 * 1024) : 0
  const overCapacity =
    host.data && (draft.value.cpus > host.data.cpus || draft.value.memory_mib > hostMemoryMib)

  return (
    <form onSubmit={save} className="max-w-2xl space-y-4">
      <ResourceNotice resource={host} name="host capacity" />
      {host.data && (
        <p className="text-sm text-muted-foreground">
          Host capacity: {host.data.cpus} CPUs · {memoryLabel(hostMemoryMib)}. These are allocations,
          not measurements of guest usage.
        </p>
      )}
      {overCapacity && (
        <NoticeBanner intent="warning">
          This allocation exceeds the host’s physical capacity and can reduce performance or prevent
          startup.
        </NoticeBanner>
      )}
      {draft.conflict && (
        <NoticeBanner intent="warning">
          Saved settings changed while you were editing. Discard this draft to load the latest values
          before saving.
        </NoticeBanner>
      )}

      <fieldset disabled={blocked}>
        <PreferencesGroup title="CPU & Memory">
          <IntegerEntryRow
            id="hardware-cpus"
            title="CPUs"
            min={1}
            required
            value={draft.value.cpus}
            onValueChange={(cpus) => draft.set({ ...draft.value, cpus })}
          />
          <MemoryEntryRow
            id="hardware-memory"
            value={draft.value.memory_mib}
            onValueChange={(memory_mib) => draft.set({ ...draft.value, memory_mib })}
          />
        </PreferencesGroup>
      </fieldset>

      <div className="flex flex-wrap items-center gap-3">
        <Button type="submit" variant="suggested" disabled={blocked || !draft.dirty || draft.conflict}>
          {action.busy ? 'Saving…' : 'Save CPU & Memory'}
        </Button>
        <Button type="button" variant="outline" disabled={!draft.dirty || action.busy} onClick={draft.discard}>
          Discard Changes
        </Button>
        <span role="status" className="text-sm text-muted-foreground">
          {draft.dirty ? 'Unsaved changes · Draft kept in this browser tab' : 'Matches saved settings'}
        </span>
      </div>

      <ActionStatus job={action.job} error={action.error} />
    </form>
  )
}
