import { useState } from 'react'
import { Trash2 } from 'lucide-react'
import { AlertDialog } from 'cheval-ui'
import { ActionButton } from './ActionButton'
import { deleteMedia, errorMessage, type Media } from '../api'
import { useSession } from '../hooks/useSession'

export function MediaRow({ media, onChange }: { media: Media; onChange?: () => void }) {
  const { admin } = useSession()
  const [confirm, setConfirm] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  async function remove() {
    setBusy(true)
    setError('')
    try {
      await deleteMedia(media.id)
      setConfirm(false)
      onChange?.()
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  const detail =
    media.kind === 'iso'
      ? 'Installation ISO'
      : `Custom disk image · ${media.size_gib} GiB`

  return (
    <div className="flex items-center justify-between gap-4 p-4">
      <div className="min-w-0">
        <p className="break-words">{media.name}</p>
        <p className="text-sm text-muted-foreground">
          {media.in_use ? `${detail} · In use` : detail}
        </p>
        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>

      {admin && (
        <ActionButton
          label={media.in_use ? `${media.name} is in use` : `Delete ${media.name}`}
          icon={Trash2}
          disabled={media.in_use}
          busy={busy}
          onClick={() => setConfirm(true)}
        />
      )}

      {admin && (
        <AlertDialog
          open={confirm}
          onCancel={() => setConfirm(false)}
          onConfirm={remove}
          busy={busy}
          destructive
          title={`Delete ${media.name}?`}
          description={
            media.kind === 'iso'
              ? 'The ISO file will be permanently removed.'
              : 'The disk image will be permanently removed.'
          }
          confirmLabel="Delete"
        />
      )}
    </div>
  )
}
