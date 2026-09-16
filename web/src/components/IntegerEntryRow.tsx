import { useEffect, useRef, useState, type ChangeEvent, type ComponentProps } from 'react'
import { EntryRow } from 'cheval-ui'

interface IntegerEntryRowProps
  extends Omit<
    ComponentProps<typeof EntryRow>,
    'value' | 'onChange' | 'type' | 'min' | 'max' | 'step' | 'error'
  > {
  value: number
  min: number
  max?: number
  onValueChange: (value: number) => void
}

function integerError(value: string, min: number, max: number) {
  if (!/^[0-9]+$/.test(value) || !Number.isSafeInteger(Number(value))) {
    return 'Enter a whole number using digits only.'
  }

  if (Number(value) < min) return `Enter a value of at least ${min}.`
  if (Number(value) > max) return `Enter a value no greater than ${max}.`

  return ''
}

export function IntegerEntryRow({
  value,
  min,
  max = Number.MAX_SAFE_INTEGER,
  onValueChange,
  ...props
}: IntegerEntryRowProps) {
  const [text, setText] = useState(String(value))
  const [edited, setEdited] = useState(false)
  const input = useRef<HTMLInputElement>(null)
  const error = integerError(text, min, max)
  useEffect(() => {
    setText(String(value))
    setEdited(false)
  }, [value])
  useEffect(() => {
    input.current?.setCustomValidity(error)
  }, [error])

  function change(event: ChangeEvent<HTMLInputElement>) {
    const next = event.target.value
    const message = integerError(next, min, max)
    setText(next)
    setEdited(true)
    event.target.setCustomValidity(message)
    if (!message) onValueChange(Number(next))
  }

  return (
    <EntryRow
      {...props}
      ref={input}
      type="text"
      inputMode="numeric"
      value={text}
      onChange={change}
      error={edited ? error : undefined}

    />
  )
}
