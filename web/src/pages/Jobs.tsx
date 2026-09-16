import { PageHeader } from 'cheval-ui'
import { useJobs } from '../hooks/useJobs'
import { JobsList } from '../components/JobsList'

export default function Jobs() {
  const resource = useJobs()

  return (
    <div className="space-y-6">
      <PageHeader title="Activity" />
      <p className="text-sm text-muted-foreground">
        Track operations and select one to see its result and saved logs.
      </p>
      <JobsList resource={resource} />
    </div>
  )
}
