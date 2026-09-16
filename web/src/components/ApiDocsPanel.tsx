import SwaggerUI from 'swagger-ui-react'
import 'swagger-ui-react/swagger-ui.css'

export default function ApiDocsPanel() {
  return (
    <div className="overflow-hidden rounded-xl border bg-white">
      <SwaggerUI url="/api/docs/openapi.json" docExpansion="none" deepLinking />
    </div>
  )
}
