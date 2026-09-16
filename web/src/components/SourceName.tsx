import { listCatalog, listMedia } from '../api'
import { useResource } from '../hooks/useResource'

export function SourceName({ image }: { image?: string }) {
  const catalog = useResource(listCatalog, [], 'catalog')
  const media = useResource(listMedia, [], 'media')

  if (!image) return <>ISO Installation</>

  const preset = catalog.data.find((item) => item.id === image)
  const custom = media.data.find((item) => `media:${item.id}` === image || item.id === image)

  if (preset) return <>{`${preset.display_name} · ${preset.arch}`}</>
  if (custom) return <>{`${custom.name} · Custom Image`}</>
  if (catalog.loading || media.loading) return <>Loading Source…</>
  return <>{`Source Unavailable (${image})`}</>
}
