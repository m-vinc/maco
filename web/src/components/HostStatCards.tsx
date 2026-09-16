import { Cpu, MemoryStick } from 'lucide-react'
import { memoryLabel } from '../ux'
import { type HostInfo, type StorageStats, type VMView } from '../api'
import { StatCard } from './StatCard'
import { DiskStatCard } from './DiskStatCard'

const MIB = 1024 * 1024

interface HostStatCardsProps {
  host: HostInfo
  vms: VMView[]
  storage: StorageStats | null
}

export function HostStatCards({ host, vms, storage }: HostStatCardsProps) {
  const running = vms.filter((vm) => vm.phase === 'running')
  const usedCpus = running.reduce((sum, vm) => sum + vm.manifest.cpus, 0)
  const usedMem = running.reduce((sum, vm) => sum + vm.manifest.memory_mib, 0) * MIB
  const allocCpus = vms.reduce((sum, vm) => sum + vm.manifest.cpus, 0)
  const allocMem = vms.reduce((sum, vm) => sum + vm.manifest.memory_mib, 0) * MIB

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <StatCard
        icon={Cpu}
        label="CPU Assigned to Running VMs"
        value={`${usedCpus} vCPUs`}
        caption={`Host capacity: ${host.cpus} CPUs · ${running.length} running VMs`}
        meta={`${allocCpus} vCPUs configured across all VMs${usedCpus > host.cpus ? ' · Running VMs share physical CPUs' : ''}`}
      />
      <StatCard
        icon={MemoryStick}
        label="Memory Assigned to Running VMs"
        value={memoryLabel(usedMem / MIB)}
        caption={`Host capacity: ${memoryLabel(host.memory_bytes / MIB)}`}
        meta={`${memoryLabel(allocMem / MIB)} configured across all VMs`}
      />
      {storage && <DiskStatCard stats={storage} />}
    </div>
  )
}
