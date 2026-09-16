import { AlertCircle } from 'lucide-react'
import { NoticeBanner } from 'cheval-ui'

interface JobNoticeProps {
  error?: string
}

export function JobNotice({ error }: JobNoticeProps) {
  if (!error) return null

  return (
    <NoticeBanner
      intent="danger"
      icon={<AlertCircle className="h-4 w-4" />}
      className="rounded-md"
    >
      {error}
    </NoticeBanner>
  )
}
