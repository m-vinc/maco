import { useState, type FormEvent } from 'react'
import { Dialog, Button, PreferencesGroup, EntryRow, NoticeBanner, SwitchRow } from 'cheval-ui'
import { createVM, listCatalog, listMedia, listNetworks, getHost, type CreateVMParams } from '../api'
import { useResource } from '../hooks/useResource'
import { useJobAction } from '../hooks/useJobAction'
import { ChoiceRow } from './ChoiceRow'
import { ImagePicker, type ImageOption } from './ImagePicker'
import { vmNetworkChoices } from './VMNetwork'
import { JobNotice } from './JobNotice'
import { ResourceNotice } from './ResourceNotice'
import { IntegerEntryRow } from './IntegerEntryRow'
import { MemoryEntryRow } from './MemoryEntryRow'
import { ValidatedEntryRow } from './ValidatedEntryRow'
import { MediaLibrary } from './MediaLibrary'
import { automaticStartupHelp, networkHelp, memoryLabel, nameError, cidrError } from '../ux'

type VMForm = Required<Omit<CreateVMParams, 'isos' | 'boot_order'>> & Pick<CreateVMParams, 'isos' | 'boot_order'>

const defaults: VMForm = {
  name: '',
  image: 'ubuntu-24.04-arm64',
  cpus: 2,
  memory_mib: 2048,
  disk_size_gib: 20,
  network: 'user',
  addresses: [],
  username: 'maco',
  password: '',
  ssh_key: '',
  autostart: false,
}

const sshKeyPattern = /^(ssh-|ecdsa-|sk-)[^\s]+\s+[A-Za-z0-9+/]+={0,3}(\s.*)?$/

export function CreateVMDialog({ onClose }: { onClose: () => void }) {
  const [form, setForm] = useState(defaults)
  const [source, setSource] = useState<'image' | 'iso'>('image')
  const [addresses, setAddresses] = useState<string[]>([])
  const [uploading, setUploading] = useState(false)
  const catalog = useResource(listCatalog, [], 'catalog')
  const media = useResource(listMedia, [], 'media')
  const networks = useResource(listNetworks, [], 'networks')
  const host = useResource(getHost, null, 'host')
  const action = useJobAction()

  const imageOptions: ImageOption[] = [
    ...catalog.data.map((image) => ({
      value: image.id,
      distro: image.distro,
      label: `${image.display_name} · ${image.arch}`,
      sub: image.downloaded
        ? 'Downloaded. A new VM receives an independent copy.'
        : 'Downloads on first use. Creation continues in Activity.',
    })),
    ...media.data
      .filter((item) => item.kind === 'image')
      .map((item) => ({
        value: `media:${item.id}`,
        distro: '',
        label: item.name,
        sub: `Custom disk image · ${item.size_gib} GiB. A new VM receives an independent copy.`,
      })),
  ]
  const isos = media.data.filter((item) => item.kind === 'iso')

  // Keep the user's choice stable when the library refreshes.
  const selectedImage = form.image
  const selectedISO = form.isos?.[0] || ''
  const ready = !media.loading && (source === 'iso' || !catalog.loading) && !networks.loading
  const sourceError = source === 'image' ? catalog.error || media.error : media.error
  const sourceAvailable =
    source === 'image'
      ? imageOptions.some((item) => item.value === selectedImage)
      : isos.some((item) => item.id === selectedISO)
  const custom = source === 'image' && selectedImage.startsWith('media:')
  const imageSize =
    source === 'image'
      ? media.data.find((item) => `media:${item.id}` === selectedImage)?.size_gib || 1
      : 1
  const diskSize = Math.max(form.disk_size_gib, imageSize)
  const validNetwork = vmNetworkChoices(networks.data).some(
    (item) => item.value === form.network && !item.disabled,
  )
  const invalidAddresses = addresses.some((address) => !!cidrError(address))
  const blocked = action.busy || uploading

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (
      !ready ||
      !sourceAvailable ||
      !validNetwork ||
      sourceError ||
      networks.error ||
      invalidAddresses ||
      nameError(form.name)
    ) {
      return
    }
    const job = await action.run(() =>
      createVM({
        ...form,
        image: source === 'image' ? selectedImage : '',
        isos: source === 'iso' ? [selectedISO] : [],
        boot_order: source === 'iso' ? [`iso:${selectedISO}`, 'disk'] : ['disk'],
        disk_size_gib: diskSize,
        addresses: source === 'image' ? addresses.map((item) => item.trim()).filter(Boolean) : [],
      }),
    )
    if (job) onClose()
  }

  function close() {
    if (!blocked) onClose()
  }

  const overCapacity =
    host.data && (form.cpus > host.data.cpus || form.memory_mib * 1048576 > host.data.memory_bytes)

  return (
    <Dialog
      open
      onClose={close}
      title="Create Virtual Machine"
      description="Choose an installation source and guest access. Hardware defaults can be adjusted below."
      className="max-w-2xl"
      footer={
        <>
          <Button variant="outline" onClick={close} disabled={blocked}>
            Cancel
          </Button>
          <Button
            variant="suggested"
            type="submit"
            form="create-vm"
            disabled={
              blocked ||
              !ready ||
              !sourceAvailable ||
              !!sourceError ||
              !!networks.error ||
              !validNetwork
            }
          >
            {action.busy ? 'Submitting…' : 'Create'}
          </Button>
        </>
      }
    >
      <form id="create-vm" onSubmit={submit} className="space-y-4">
        <JobNotice error={action.error} />

        <fieldset disabled={blocked} className="space-y-4">
          <PreferencesGroup title="Virtual Machine">
            <ValidatedEntryRow
              id="vm-name"
              title="Name"
              required
              autoFocus
              value={form.name}
              error={form.name ? nameError(form.name) : ''}
              help="1–63 letters, digits, dots, underscores or hyphens"
              onChange={(event) => setForm({ ...form, name: event.target.value })}
            />
            <ChoiceRow
              stacked
              title="Install from"
              value={source}
              choices={[
                { value: 'image', label: 'Disk image' },
                { value: 'iso', label: 'Installation ISO' },
              ]}
              onChange={(value) => setSource(value as 'image' | 'iso')}
            />
            <ResourceNotice resource={media} name="installation media" />
            {source === 'image' && <ResourceNotice resource={catalog} name="system images" />}

            {ready && !sourceError && (
              source === 'image' ? (
                <ImagePicker
                  title="Image"
                  value={selectedImage}
                  options={imageOptions}
                  disabled={blocked}
                  onChange={(image) => setForm({ ...form, image })}
                />
              ) : (
                <ChoiceRow
                  stacked
                  title="Installation ISO"
                  value={selectedISO}
                  choices={isos.map((item) => ({ value: item.id, label: item.name }))}
                  onChange={(id) => setForm({ ...form, isos: [id] })}
                  help="Install the operating system and create its account using the guest installer."
                />
              )
            )}

            {ready && !sourceError && !sourceAvailable && (
              <p role="status" className="px-4 py-2 text-sm">
                {selectedImage && source === 'image'
                  ? 'The selected image is unavailable. Choose another image or upload one below.'
                  : 'Choose available installation media, or upload it below.'}
              </p>
            )}
          </PreferencesGroup>

          {source === 'image' && (
            <PreferencesGroup
              title="Guest Setup"
              description={
                custom
                  ? 'Automatic setup is unverified for custom images. Existing guest accounts may apply; these settings only take effect if the image supports cloud-init.'
                  : 'Create an account inside the VM. This is separate from the maco sign-in account.'
              }
            >
              <EntryRow
                id="vm-user"
                title="Guest username"
                required
                value={form.username}
                onChange={(event) => setForm({ ...form, username: event.target.value })}
              />
              <EntryRow
                id="vm-password"
                title="Guest password (optional)"
                type="password"
                autoComplete="new-password"
                value={form.password}
                help="A password enables guest password sign-in when automatic setup is supported."
                onChange={(event) => setForm({ ...form, password: event.target.value })}
              />
              <ValidatedEntryRow
                id="vm-key"
                title="SSH public key (optional)"
                value={form.ssh_key}
                help="Paste a public key, for example ssh-ed25519 AAAA…; never a private key."
                error={
                  form.ssh_key && !sshKeyPattern.test(form.ssh_key.trim())
                    ? 'Enter a complete SSH public key.'
                    : ''
                }
                onChange={(event) => setForm({ ...form, ssh_key: event.target.value })}
              />
              {!form.password && !form.ssh_key && (
                <p className="px-4 py-2 text-sm text-muted-foreground">
                  No password or key is configured. Automatic setup will not provide a known sign-in
                  method; add one for ordinary guest access.
                </p>
              )}
            </PreferencesGroup>
          )}

          <PreferencesGroup title="Connectivity">
            <ResourceNotice resource={networks} name="networks" />
            <ChoiceRow
              stacked
              title="Network"
              value={form.network}
              choices={vmNetworkChoices(networks.data, form.network, networks.loading)}
              disabled={networks.loading}
              onChange={(network) => setForm({ ...form, network })}
              help={networkHelp(form.network, networks.data)}
            />
          </PreferencesGroup>

          <details className="rounded-xl border p-4">
            <summary className="cursor-pointer font-semibold">
              Hardware · {form.cpus} virtual CPUs · {memoryLabel(form.memory_mib)} · {diskSize} GiB disk
            </summary>
            <div className="mt-3 space-y-3">
              <ResourceNotice resource={host} name="host capacity" />
              {host.data && (
                <p className="text-sm text-muted-foreground">
                  Host capacity: {host.data.cpus} CPUs · {memoryLabel(host.data.memory_bytes / 1048576)}{' '}
                  memory. Assigned resources are shared with this Mac and other VMs.
                </p>
              )}
              <PreferencesGroup>
                <IntegerEntryRow
                  id="vm-cpus"
                  title="Virtual CPUs"
                  min={1}
                  required
                  value={form.cpus}
                  onValueChange={(cpus) => setForm({ ...form, cpus })}
                />
                <MemoryEntryRow
                  id="vm-memory"
                  value={form.memory_mib}
                  onValueChange={(memory_mib) => setForm({ ...form, memory_mib })}
                />
                <IntegerEntryRow
                  id="vm-disk"
                  title="Disk capacity (GiB)"
                  min={imageSize}
                  required
                  value={diskSize}
                  help={
                    imageSize > 1
                      ? `The image needs at least ${imageSize} GiB.`
                      : 'Virtual capacity; actual host disk use grows as data is written.'
                  }
                  onValueChange={(disk_size_gib) => setForm({ ...form, disk_size_gib })}
                />
              </PreferencesGroup>
              {overCapacity && (
                <NoticeBanner intent="warning">
                  The requested assignment exceeds host capacity. Reduce it or consider other running
                  workloads.
                </NoticeBanner>
              )}
            </div>
          </details>

          <details className="rounded-xl border p-4">
            <summary className="cursor-pointer font-semibold">Advanced Options</summary>
            <div className="mt-3 space-y-3">
              <PreferencesGroup>
                <SwitchRow
                  title="Automatic Startup"
                  subtitle={automaticStartupHelp}
                  checked={form.autostart}
                  onCheckedChange={(autostart) => setForm({ ...form, autostart })}
                />
              </PreferencesGroup>
              {source === 'image' && (
                <PreferencesGroup
                  title="Static Guest Addresses"
                  description="Leave empty for automatic addressing. These addresses belong to the guest and require automatic guest setup support."
                >
                  {addresses.map((address, index) => (
                    <div key={index} className="flex items-start">
                      <div className="min-w-0 flex-1">
                        <ValidatedEntryRow
                          id={`vm-address-${index}`}
                          title={`Guest IP address ${index + 1}`}
                          required
                          value={address}
                          error={cidrError(address)}
                          help="For example 192.168.1.20/24 or 2001:db8::20/64"
                          onChange={(event) =>
                            setAddresses(
                              addresses.map((item, i) => (i === index ? event.target.value : item)),
                            )
                          }
                        />
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        className="mt-3 mr-2"
                        aria-label={`Remove address ${index + 1}`}
                        onClick={() => setAddresses(addresses.filter((_, i) => i !== index))}
                      >
                        Remove
                      </Button>
                    </div>
                  ))}
                  <Button type="button" className="m-3" onClick={() => setAddresses([...addresses, ''])}>
                    Add Address
                  </Button>
                </PreferencesGroup>
              )}
            </div>
          </details>
        </fieldset>

        <details className="rounded-xl border p-4">
          <summary className="cursor-pointer font-semibold">Upload Installation Media…</summary>
          <div className="mt-3">
            <MediaLibrary mode={source} onBusyChange={setUploading} onChange={media.refresh} />
          </div>
        </details>

        <p className="text-sm text-muted-foreground">
          Creation is submitted to Activity. Preparing a VM does not start it unless Automatic Startup
          is enabled.
        </p>
      </form>
    </Dialog>
  )
}
