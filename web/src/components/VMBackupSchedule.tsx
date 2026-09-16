import { useCallback, useEffect, useState } from 'react'
import { Button, NoticeBanner, PreferencesGroup, SwitchRow } from 'cheval-ui'
import {
  getBackupSchedule,
  setBackupSchedule,
  errorMessage,
  type BackupSchedule,
} from '../api'
import { useResource } from '../hooks/useResource'
import { IntegerEntryRow } from './IntegerEntryRow'

interface VMBackupScheduleProps {
  vmID: string
}

export function VMBackupSchedule({ vmID }: VMBackupScheduleProps) {
  const load = useCallback(() => getBackupSchedule(vmID), [vmID])
  const schedule = useResource<BackupSchedule | null>(load, null, 'vms')
  const [enabled, setEnabled] = useState(false)
  const [interval, setInterval] = useState(24)
  const [keepLast, setKeepLast] = useState(0)
  const [maxAgeDays, setMaxAgeDays] = useState(0)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    const data = schedule.data
    if (!data) return
    setEnabled(data.enabled)
    setInterval(data.interval_hours || 24)
    setKeepLast(data.keep_last)
    setMaxAgeDays(data.max_age_days)
  }, [schedule.data])

  const valid = !enabled || interval >= 1

  async function save() {
    if (saving || !valid) return
    setSaving(true)
    setError('')
    setSaved(false)
    try {
      await setBackupSchedule(vmID, {
        enabled,
        interval_hours: interval,
        keep_last: keepLast,
        max_age_days: maxAgeDays,
      })
      setSaved(true)
      schedule.refresh()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  const lastRun = schedule.data?.last_run_at
    ? new Date(schedule.data.last_run_at * 1000).toLocaleString()
    : 'Never'

  return (
    <div className="space-y-3">
      <h2 className="font-semibold">Scheduled backups</h2>
      <p className="text-sm text-muted-foreground">
        Automatically back up this VM on a fixed interval and clean up old backups with a retention
        policy. Last automatic backup: {lastRun}.
      </p>
      {schedule.error && <NoticeBanner intent="danger">{schedule.error}</NoticeBanner>}
      {error && <NoticeBanner intent="danger">{error}</NoticeBanner>}
      {saved && <NoticeBanner intent="success">Schedule saved.</NoticeBanner>}
      <fieldset disabled={saving}>
        <PreferencesGroup>
          <SwitchRow
            title="Enable scheduled backups"
            checked={enabled}
            onCheckedChange={setEnabled}
          />
          <IntegerEntryRow
            id="backup-interval"
            title="Interval (hours)"
            min={1}
            value={interval}
            onValueChange={setInterval}
          />
          <IntegerEntryRow
            id="backup-keep-last"
            title="Keep last (0 = unlimited)"
            min={0}
            value={keepLast}
            onValueChange={setKeepLast}
          />
          <IntegerEntryRow
            id="backup-max-age"
            title="Delete older than (days, 0 = never)"
            min={0}
            value={maxAgeDays}
            onValueChange={setMaxAgeDays}
          />
        </PreferencesGroup>
      </fieldset>
      <div className="flex justify-end">
        <Button variant="suggested" disabled={saving || !valid} onClick={save}>
          {saving ? 'Saving…' : 'Save schedule'}
        </Button>
      </div>
    </div>
  )
}
