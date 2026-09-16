import { PageHeader, NoticeBanner } from 'cheval-ui'
import { HostConsole } from '../components/HostConsole'
import { useSession } from '../hooks/useSession'

export default function Console() {
  const { admin } = useSession()

  return (
    <div className="space-y-6">
      <PageHeader
        title="Console"
        description="A terminal on the maco host machine."
      />
      {admin ? (
        <HostConsole />
      ) : (
        <NoticeBanner intent="warning">
          The host console requires an administrator account.
        </NoticeBanner>
      )}
    </div>
  )
}
