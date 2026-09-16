import { type NetworkInterface } from '../api'

interface InterfaceMembersProps {
  title: string
  interfaces: NetworkInterface[]
  selected: string[]
  onChange: (members: string[]) => void
}

export function InterfaceMembers({
  title,
  interfaces,
  selected,
  onChange,
}: InterfaceMembersProps) {
  function toggle(device: string) {
    onChange(
      selected.includes(device)
        ? selected.filter((entry) => entry !== device)
        : [...selected, device],
    )
  }

  return (
    <fieldset className="flex min-w-0 flex-col gap-1.5 px-4 py-2">
      <legend className="text-xs font-medium leading-4 text-muted-foreground">
        {title}
      </legend>
      {interfaces.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No host interfaces detected.
        </p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {interfaces.map((iface) => {
            const active = selected.includes(iface.device)
            return (
              <label
                key={iface.device}
                className="flex min-h-10 items-center gap-2 rounded-md border px-3 py-2 text-sm focus-within:ring-2 focus-within:ring-inset focus-within:ring-ring"
              >
                <input
                  type="checkbox"
                  checked={active}
                  onChange={() => toggle(iface.device)}
                  className="h-4 w-4 shrink-0"
                />
                <span className="break-words font-medium">
                  {iface.device}
                  {iface.hardware_port ? ` · ${iface.hardware_port}` : ' · Not currently detected'}
                </span>
              </label>
            )
          })}
        </div>
      )}
    </fieldset>
  )
}
