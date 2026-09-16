import { type LucideIcon } from 'lucide-react'
import { Card } from 'cheval-ui'

interface StatCardProps {
  icon: LucideIcon
  label: string
  value: string
  caption?: string
  pct?: number
  meta?: string
  warn?: boolean
  barLabel?: string
}

export function StatCard({ icon: Icon, label, value, caption, pct, meta, warn, barLabel }: StatCardProps) {
  const fill = Math.max(0, Math.min(100, Math.round(pct || 0)))

  return (
    <Card className="space-y-3 p-4">
      <div className="flex items-center gap-2 text-muted-foreground">
        <Icon aria-hidden="true" className="h-4 w-4 shrink-0" />
        <span className="text-sm font-semibold">{label}</span>
      </div>

      <div>
        <p className="text-2xl font-semibold leading-tight tabular-nums">{value}</p>
        {caption && <p className="mt-1 text-sm text-muted-foreground">{caption}</p>}
      </div>

      {pct !== undefined && (
        <div
          role="meter"
          aria-label={barLabel || label}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={fill}
          aria-valuetext={`${Math.round(pct)}%`}
          className="h-1.5 w-full overflow-hidden rounded-full bg-muted"
        >
          <div
            className={`h-full rounded-full ${warn ? 'bg-warning' : 'bg-primary'}`}
            style={{ width: `${fill}%` }}
          />
        </div>
      )}

      {meta && <p className="text-sm text-muted-foreground tabular-nums">{meta}</p>}
    </Card>
  )
}
