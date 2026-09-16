import { useId, useRef, useState } from 'react'
import { Button, PreferencesGroup, NoticeBanner, fmtBytes } from 'cheval-ui'
import { uploadISO, uploadImage, errorMessage } from '../api'
import { useSession } from '../hooks/useSession'
import { JobNotice } from './JobNotice'

const config = {
  iso: {
    title: 'Upload Installation ISO',
    description: 'Choose an ARM64 ISO, up to 10 GiB.',
    accept: '.iso',
    label: 'Choose ISO…',
    upload: uploadISO,
  },
  image: {
    title: 'Upload Custom Disk Image',
    description: 'Choose an ARM64 .qcow2 or .img. New VMs receive an independent copy.',
    accept: '.qcow2,.img',
    label: 'Choose Image…',
    upload: uploadImage,
  },
}

interface MediaLibraryProps {
  mode?: 'image' | 'iso'
  onChange?: () => void
  onBusyChange?: (busy: boolean) => void
}

export function MediaLibrary({ mode = 'image', onChange, onBusyChange }: MediaLibraryProps) {
  const { admin } = useSession()
  const id = useId()
  const input = useRef<HTMLInputElement>(null)
  const [file, setFile] = useState<File | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [progress, setProgress] = useState(0)
  const [saved, setSaved] = useState('')
  const { title, description, accept, label, upload } = config[mode]

  if (!admin) return null

  async function perform() {
    if (!file || busy) return
    setBusy(true)
    onBusyChange?.(true)
    setError('')
    setSaved('')
    setProgress(0)
    try {
      await upload(file, setProgress)
      setSaved(`${file.name} is available in the library.`)
      setFile(null)
      onChange?.()
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
      onBusyChange?.(false)
    }
  }

  const percent = Math.round(progress * 100)

  return (
    <div className="space-y-3">
      <JobNotice error={error} />
      {saved && <NoticeBanner intent="success">{saved}</NoticeBanner>}

      <PreferencesGroup title={title} description={description}>
        <div className="space-y-3 p-4">
          <input
            ref={input}
            id={id}
            aria-label={label}
            type="file"
            accept={accept}
            disabled={busy}
            className="sr-only"
            onChange={(event) => {
              setFile(event.target.files?.[0] || null)
              setSaved('')
              setError('')
              event.target.value = ''
            }}
          />

          <div className="flex flex-wrap gap-2">
            <Button type="button" disabled={busy} onClick={() => input.current?.click()}>
              {label}
            </Button>
            {file && (
              <Button type="button" variant="suggested" disabled={busy} onClick={perform}>
                {busy ? 'Uploading…' : error ? 'Retry Upload' : 'Upload'}
              </Button>
            )}
          </div>

          {file && (
            <p className="break-words text-sm">
              {file.name} · {fmtBytes(file.size)}
            </p>
          )}

          {busy && (
            <div className="space-y-2">
              <div
                role="progressbar"
                aria-label={`Upload ${file?.name}`}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={percent}
                className="h-2 w-full overflow-hidden rounded-full bg-muted"
              >
                <div className="h-full bg-primary" style={{ width: `${percent}%` }} />
              </div>
              <p role="status" className="text-sm text-muted-foreground">
                {progress < 1
                  ? `${fmtBytes((file?.size || 0) * progress)} of ${fmtBytes(file?.size || 0)} uploaded`
                  : 'Transfer complete. Finalizing the library entry…'}
              </p>
            </div>
          )}
        </div>
      </PreferencesGroup>
    </div>
  )
}
