import { HardDrive } from 'lucide-react'
import { fmtBytes, NoticeBanner } from 'cheval-ui'
import { type StorageStats } from '../api'
import { StatCard } from './StatCard'

export function DiskStatCard({ stats }: { stats: StorageStats }) {
  const pct =
    stats.total_bytes > 0
      ? ((stats.total_bytes - stats.free_bytes) / stats.total_bytes) * 100
      : 0

  return (
    <div className="space-y-2">
      <StatCard
        icon={HardDrive}
        label="Host Storage"
        value={`${fmtBytes(stats.free_bytes)} free`}
        caption={`Host volume capacity: ${fmtBytes(stats.total_bytes)}`}
        pct={pct}
        barLabel="Host volume space used"
        warn={pct >= 90}
        meta={`VM disk files: ${fmtBytes(stats.disks_bytes)} · ${stats.disk_count} disks`}
      />
      {pct >= 90 && (
        <NoticeBanner intent="warning">
          Host storage is low. Other files on this Mac also use this volume.
        </NoticeBanner>
      )}
    </div>
  )
}
