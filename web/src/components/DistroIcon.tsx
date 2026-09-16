const known = new Set(['ubuntu', 'debian', 'fedora', 'rocky-linux', 'almalinux', 'opensuse', 'alpine'])

export function DistroIcon({ distro, className = '' }: { distro: string; className?: string }) {
  const slug = known.has(distro) ? distro : 'tux'
  return <i aria-hidden="true" className={`fl-${slug} ${className}`} />
}
