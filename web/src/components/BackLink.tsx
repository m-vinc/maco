import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from 'cheval-ui'

interface BackLinkProps {
  to: string
  children: React.ReactNode
}

export function BackLink({ to, children }: BackLinkProps) {
  return (
    <Button variant="ghost" size="sm" asChild className="-ml-3.5 gap-2 self-start">
      <Link to={to}>
        <ArrowLeft className="h-4 w-4 shrink-0" />
        {children}
      </Link>
    </Button>
  )
}
