import { Badge, type BadgeProps } from 'cheval-ui'
import { CheckCircle2, CircleAlert, Clock, LoaderCircle } from 'lucide-react'
import { type Job } from '../api'
import { activityState } from '../ux'

const jobVariants: Record<Job['state'], BadgeProps['variant']> = {
  pending: 'warning',
  running: 'info',
  succeeded: 'success',
  failed: 'destructive',
}

const jobIcons = {
  pending: Clock,
  running: LoaderCircle,
  succeeded: CheckCircle2,
  failed: CircleAlert,
}

export function JobStatusBadge({ state, job }: { state: Job['state']; job?: Job }) {
  const Icon = jobIcons[state]

  return (
    <Badge variant={jobVariants[state]} className="gap-1.5 whitespace-nowrap">
      <Icon
        aria-hidden="true"
        className={`h-3 w-3 ${state === 'running' ? 'animate-spin motion-reduce:animate-none' : ''}`}
      />
      {job ? activityState(job) : { pending: 'Queued', running: 'In Progress', succeeded: 'Completed', failed: 'Failed' }[state]}
    </Badge>
  )
}

export function VMStatusBadge({ phase }: { phase: string }) {
  return (
    <Badge
      variant={phase === 'running' ? 'success' : 'neutral'}
      className="whitespace-nowrap"
    >
      {phase === 'running'
        ? 'Running'
        : phase === 'stopped'
          ? 'Stopped'
          : phase}
    </Badge>
  )
}
