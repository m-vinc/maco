import { useEffect, useRef, type ComponentProps } from 'react'
import { EntryRow } from 'cheval-ui'

export function ValidatedEntryRow({ help, error, ...props }: ComponentProps<typeof EntryRow> & { help?: string }) {
  const input = useRef<HTMLInputElement>(null)
  useEffect(() => { input.current?.setCustomValidity(error || '') }, [error])
  return <EntryRow {...props} ref={input} error={error} help={help} />
}
