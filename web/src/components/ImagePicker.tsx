import { ChoiceRow } from './ChoiceRow'

export interface ImageOption {
  value: string
  distro: string
  label: string
  sub: string
}

interface ImagePickerProps {
  title: string
  options: ImageOption[]
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

export function ImagePicker({ title, options, value, onChange, disabled }: ImagePickerProps) {
  return (
    <ChoiceRow
      stacked
      title={title}
      value={value}
      choices={options.map((item) => ({ value: item.value, label: item.label }))}
      onChange={onChange}
      disabled={disabled}
      help={options.find((item) => item.value === value)?.sub}
    />
  )
}
