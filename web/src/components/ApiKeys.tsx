import { useState, type FormEvent } from 'react'
import {
  AlertDialog,
  Badge,
  Button,
  CopyButton,
  EntryRow,
  NoticeBanner,
  PreferencesGroup,
} from 'cheval-ui'
import { Trash2 } from 'lucide-react'
import { createAPIKey, listAPIKeys, revokeAPIKey, errorMessage, type APIKey } from '../api'
import { useResource } from '../hooks/useResource'
import { ResourceNotice } from './ResourceNotice'
import { JobNotice } from './JobNotice'

function formatDate(seconds?: number | null) {
  if (!seconds) return 'Never'
  return new Date(seconds * 1000).toLocaleString()
}

export function ApiKeys() {
  const keys = useResource(listAPIKeys, [], 'api-keys')
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [created, setCreated] = useState<{ key: APIKey; token: string } | null>(null)
  const [revoking, setRevoking] = useState<APIKey | null>(null)

  async function create(event: FormEvent) {
    event.preventDefault()
    if (!name.trim() || busy) return
    setBusy(true)
    setError('')
    try {
      const result = await createAPIKey(name.trim())
      setCreated({ key: result.api_key, token: result.token })
      setName('')
      keys.refresh()
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  async function revoke() {
    if (!revoking || busy) return
    setBusy(true)
    setError('')
    try {
      await revokeAPIKey(revoking.id)
      setRevoking(null)
      keys.refresh()
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="max-w-3xl space-y-6">
      <JobNotice error={error} />

      {created && (
        <NoticeBanner intent="success">
          <div className="space-y-2">
            <p className="font-medium">Copy your new API key now; it will not be shown again.</p>
            <div className="flex flex-wrap items-center gap-2">
              <code className="min-w-0 break-all rounded bg-background/60 px-2 py-1 text-sm">
                {created.token}
              </code>
              <CopyButton text={created.token} label="Copy API key" />
            </div>
            <p className="text-sm text-muted-foreground">
              Use it as a bearer token: <code>Authorization: Bearer {created.key.prefix}…</code>. It
              carries your account’s permissions.
            </p>
            <Button variant="outline" onClick={() => setCreated(null)}>
              Done
            </Button>
          </div>
        </NoticeBanner>
      )}

      <form onSubmit={create}>
        <PreferencesGroup
          title="Create API Key"
          description="Generate a token bound to your account for programmatic access to this API."
        >
          <EntryRow
            id="api-key-name"
            title="Name"
            placeholder="For example, CI pipeline"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
          <div className="p-4">
            <Button type="submit" variant="suggested" disabled={!name.trim() || busy}>
              {busy ? 'Creating…' : 'Create API Key'}
            </Button>
          </div>
        </PreferencesGroup>
      </form>

      <ResourceNotice resource={keys} name="API keys" />

      <PreferencesGroup title="Active API Keys">
        {keys.data.map((key) => (
          <div key={key.id} className="flex flex-wrap items-center gap-3 p-4">
            <div className="min-w-0 flex-1">
              <p className="flex flex-wrap items-center gap-2 font-medium">
                <span className="break-words">{key.name}</span>
                <Badge variant="neutral">{key.prefix}…</Badge>
              </p>
              <p className="text-sm text-muted-foreground">
                Created {formatDate(key.created_at)} · Last used {formatDate(key.last_used_at)}
              </p>
            </div>
            <Button size="sm" variant="outline" disabled={busy} onClick={() => setRevoking(key)}>
              <Trash2 aria-hidden="true" className="mr-2 h-4 w-4" />
              Revoke…
            </Button>
          </div>
        ))}
        {!keys.loading && !keys.error && keys.data.length === 0 && (
          <p className="p-4 text-sm text-muted-foreground">
            No API keys yet. Create one above to access the API without signing in.
          </p>
        )}
      </PreferencesGroup>

      <AlertDialog
        open={!!revoking}
        onCancel={() => setRevoking(null)}
        onConfirm={revoke}
        busy={busy}
        destructive
        title={`Revoke ${revoking?.name || 'API key'}?`}
        description="Any client using this key immediately loses access. This cannot be undone."
        confirmLabel="Revoke Key"
      />
    </div>
  )
}
