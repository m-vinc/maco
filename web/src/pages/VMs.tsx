import { useState } from 'react'
import {
  EmptyState,
  PageHeader,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
} from 'cheval-ui'
import { Server } from 'lucide-react'
import { listVMs, getHost, getDiskStorage } from '../api'
import { useResource } from '../hooks/useResource'
import { useJobAction } from '../hooks/useJobAction'
import { CreateVMDialog } from '../components/CreateVMDialog'
import { HostStatCards } from '../components/HostStatCards'
import { SplitButton } from '../components/SplitButton'
import { VMRow } from '../components/VMRow'
import { ResourceNotice } from '../components/ResourceNotice'
import { JobNotice } from '../components/JobNotice'
import { useSession } from '../hooks/useSession'

export default function VMs() {
  const { admin } = useSession()
  const [creating, setCreating] = useState(false)
  const vms = useResource(listVMs, [], 'vms')
  const host = useResource(getHost, null, 'host')
  const storage = useResource(getDiskStorage, null, 'storage')
  const action = useJobAction()

  return (
    <div className="space-y-6">
      <PageHeader
        title="Virtual machines"
        action={
          admin ? (
            <SplitButton
              label="Create virtual machine"
              onClick={() => setCreating(true)}
            />
          ) : undefined
        }
      />
      <ResourceNotice resource={host} name="host capacity" />
      <ResourceNotice resource={storage} name="host storage" />
      {host.data && (
        <HostStatCards host={host.data} vms={vms.data} storage={storage.data} />
      )}
      <ResourceNotice resource={vms} name="virtual machines" />
      <JobNotice error={action.error} />
      {!vms.loading && (vms.data.length ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Preview</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Source</TableHead>
              <TableHead>Memory</TableHead>
              <TableHead>State</TableHead>
              <TableHead>Automatic Startup</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {vms.data.map((vm) => (
              <VMRow key={vm.manifest.id} vm={vm} run={action.run} />
            ))}
          </TableBody>
        </Table>
      ) : (
        !vms.error && (
          <EmptyState
            icon={<Server />}
            title="No virtual machines yet"
            message={
              admin
                ? 'Create your first virtual machine to get started.'
                : 'An administrator can create virtual machines.'
            }
          />
        )
      ))}
      {creating && <CreateVMDialog onClose={() => setCreating(false)} />}
    </div>
  )
}
