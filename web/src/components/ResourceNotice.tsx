import { Button, NoticeBanner, Spinner } from 'cheval-ui'

interface ResourceNoticeProps {
  resource: { loading: boolean; error: string; refresh: () => void }
  name: string
}

export function ResourceNotice({ resource, name }: ResourceNoticeProps) {
  if (resource.loading) {
    return (
      <div role="status" className="flex items-center gap-2 text-sm text-muted-foreground">
        <Spinner />
        Loading {name}…
      </div>
    )
  }

  if (!resource.error) return null

  return (
    <NoticeBanner intent="danger">
      <div className="space-y-2">
        <p>Couldn’t refresh {name}. Previously loaded information may be out of date.</p>
        <Button type="button" onClick={resource.refresh}>
          Retry
        </Button>
        <details>
          <summary className="cursor-pointer text-xs">Technical Details</summary>
          <p className="break-words text-sm">{resource.error}</p>
        </details>
      </div>
    </NoticeBanner>
  )
}
