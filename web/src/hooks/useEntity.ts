import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useResource } from './useResource'

// Loads a single entity and redirects to `fallback` once it no longer exists,
// i.e. the loader resolves to null (deleted or never found). Transient errors
// keep the user on the page so a blip does not eject them.
export function useEntity<T>(load: () => Promise<T | null>, domain: string, fallback: string) {
  const navigate = useNavigate()
  const resource = useResource<T | null>(load, null, domain)
  useEffect(() => {
    if (!resource.loading && resource.data === null && !resource.error) navigate(fallback, { replace: true })
  }, [resource.loading, resource.data, resource.error, navigate, fallback])
  return resource
}
