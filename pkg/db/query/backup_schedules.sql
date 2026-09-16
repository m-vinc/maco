-- name: UpsertBackupSchedule :exec
INSERT INTO backup_schedules (vm_id, enabled, interval_hours, keep_last, max_age_days, last_run_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(vm_id) DO UPDATE SET
    enabled = excluded.enabled,
    interval_hours = excluded.interval_hours,
    keep_last = excluded.keep_last,
    max_age_days = excluded.max_age_days;

-- name: GetBackupSchedule :one
SELECT vm_id, enabled, interval_hours, keep_last, max_age_days, last_run_at FROM backup_schedules WHERE vm_id = ?;

-- name: ListBackupSchedules :many
SELECT vm_id, enabled, interval_hours, keep_last, max_age_days, last_run_at FROM backup_schedules;

-- name: MarkBackupScheduleRun :exec
UPDATE backup_schedules SET last_run_at = ? WHERE vm_id = ?;

-- name: DeleteBackupSchedule :exec
DELETE FROM backup_schedules WHERE vm_id = ?;
