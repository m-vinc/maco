import { JobTarget } from './JobTarget'
import { JobStatusBadge } from './StatusBadge'
import { Link } from 'react-router-dom'
import { TableRow, TableCell } from 'cheval-ui'
import { type Job } from '../api'
import { activityName, unfinishedTime } from '../ux'

interface JobRowProps {
  job: Job
  highlighted?: boolean
}

export function JobRow({ job, highlighted }: JobRowProps) {
  return (
    <TableRow
      aria-selected={highlighted || undefined}
      className={highlighted ? 'bg-primary/10 hover:bg-primary/15' : undefined}
    >
      <TableCell>
        <Link className="text-primary hover:underline" to={`/jobs/${job.id}`}>
          {activityName(job)}
        </Link>
      </TableCell>
      <TableCell>
        <JobTarget job={job} />
      </TableCell>
      <TableCell>
        <JobStatusBadge state={job.state} job={job} />
      </TableCell>
      <TableCell>{new Date(job.created_at).toLocaleString()}</TableCell>
      <TableCell>
        {job.finished_at
          ? new Date(job.finished_at).toLocaleString()
          : unfinishedTime(job)}
      </TableCell>
    </TableRow>
  )
}
