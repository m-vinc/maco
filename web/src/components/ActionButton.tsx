import { Button } from 'cheval-ui'
import { LoaderCircle, type LucideIcon } from 'lucide-react'

interface ActionButtonProps {
  label: string
  icon: LucideIcon
  disabled?: boolean
  busy?: boolean
  onClick: () => void
}

export function ActionButton({
  label,
  icon: Icon,
  disabled,
  busy,
  onClick,
}: ActionButtonProps) {
  return (
    <span title={label} className="inline-flex">
      <Button
        variant="ghost"
        size="icon"
        aria-label={label}
        aria-busy={busy || undefined}
        disabled={disabled || busy}
        onClick={onClick}
      >
        {busy ? (
          <LoaderCircle aria-hidden="true" className="h-4 w-4 animate-spin" />
        ) : (
          <Icon aria-hidden="true" className="h-4 w-4" />
        )}
      </Button>
    </span>
  )
}
