import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Plus, type LucideIcon } from 'lucide-react'
import { Button } from 'cheval-ui'

export interface SplitAction {
  label: string
  icon?: LucideIcon
  onClick: () => void
  disabled?: boolean
}

interface SplitButtonProps {
  label: string
  onClick: () => void
  icon?: LucideIcon
  actions?: SplitAction[]
}

export function SplitButton({
  label,
  onClick,
  icon: Icon = Plus,
  actions = [],
}: SplitButtonProps) {
  const [open, setOpen] = useState(false)
  const container = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return

    function dismiss(event: MouseEvent) {
      if (!container.current?.contains(event.target as Node)) setOpen(false)
    }

    function close(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false)
    }

    document.addEventListener('mousedown', dismiss)
    document.addEventListener('keydown', close)
    return () => {
      document.removeEventListener('mousedown', dismiss)
      document.removeEventListener('keydown', close)
    }
  }, [open])

  const hasActions = actions.length > 0

  function choose(action: SplitAction) {
    setOpen(false)
    action.onClick()
  }

  function renderAction(action: SplitAction) {
    const ActionIcon = action.icon
    return (
      <button
        key={action.label}
        type="button"
        role="menuitem"
        disabled={action.disabled}
        onClick={() => choose(action)}
        className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm text-foreground hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50"
      >
        {ActionIcon && <ActionIcon className="h-4 w-4 shrink-0" />}
        {action.label}
      </button>
    )
  }

  return (
    <div ref={container} className="relative inline-flex">
      <Button
        variant="suggested"
        onClick={onClick}
        className={hasActions ? 'gap-2 rounded-r-none' : 'gap-2'}
      >
        <Icon className="h-4 w-4 shrink-0" />
        {label}
      </Button>
      {hasActions && (
        <>
          <Button
            variant="suggested"
            aria-label="More actions"
            aria-haspopup="menu"
            aria-expanded={open}
            onClick={() => setOpen((value) => !value)}
            className="rounded-l-none border-l border-primary-foreground/20 px-2"
          >
            <ChevronDown className="h-4 w-4 shrink-0" />
          </Button>
          {open && (
            <div
              role="menu"
              className="absolute right-0 top-[calc(100%+0.25rem)] z-20 min-w-52 rounded-md border border-border bg-popover p-1 shadow-md"
            >
              {actions.map(renderAction)}
            </div>
          )}
        </>
      )}
    </div>
  )
}
