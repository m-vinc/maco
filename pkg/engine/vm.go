package engine

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/m-vinc/maco/pkg/cloudinit"
	"github.com/m-vinc/maco/pkg/firmware"
	"github.com/m-vinc/maco/pkg/image"
	"github.com/m-vinc/maco/pkg/network"
	"github.com/m-vinc/maco/pkg/types"
	"github.com/m-vinc/maco/pkg/vm"
	"github.com/rs/zerolog/log"
)

type VMView struct {
	Manifest *types.VMManifest `json:"manifest"`
	Phase    string            `json:"phase"`
	PID      int               `json:"pid"`
	BootTime int64             `json:"boot_time"`
}

type CreateVMParams struct {
	ISOs        []string `json:"isos,omitempty" binding:"optional"`
	BootOrder   []string `json:"boot_order,omitempty" binding:"optional"`
	Name        string   `json:"name" binding:"optional"`
	Image       string   `json:"image" binding:"optional"`
	CPUs        int      `json:"cpus" binding:"optional"`
	MemoryMiB   int      `json:"memory_mib" binding:"optional"`
	DiskSizeGiB int      `json:"disk_size_gib" binding:"optional"`
	Network     string   `json:"network" binding:"optional"`
	Addresses   []string `json:"addresses" binding:"optional"`
	Username    string   `json:"username" binding:"optional"`
	Password    string   `json:"password" binding:"optional"`
	SSHKey      string   `json:"ssh_key" binding:"optional"`
	Autostart   bool     `json:"autostart" binding:"optional"`
}

var vmName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,62}$`)

func (e *Engine) CreateVM(params CreateVMParams) (*types.VMManifest, error) {
	if !vmName.MatchString(params.Name) {
		return nil, fmt.Errorf("name must be 1 to 63 characters of letters, digits, dot, dash or underscore")
	}

	for _, address := range params.Addresses {
		if _, _, err := net.ParseCIDR(address); err != nil {
			return nil, fmt.Errorf("address %q must use CIDR notation", address)
		}
	}

	if (params.Image != "" && len(params.ISOs) != 0) || (params.Image == "" && len(params.ISOs) != 1) {
		return nil, fmt.Errorf("select either one image or one installation ISO")
	}

	lock, err := e.lockMedia(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Close() }()

	if params.Image != "" {
		if _, err := e.imageBase(context.Background(), params.Image, false); err != nil {
			return nil, err
		}
	}

	networkRef := params.Network
	if networkRef != "" && networkRef != "user" && networkRef != "vmnet-shared" && networkRef != "vmnet-host" {
		n, err := e.nets.Resolve(networkRef)
		if err != nil {
			return nil, fmt.Errorf("resolve network %q: %w", networkRef, err)
		}

		networkRef = n.ID
	}

	if err := e.validateMedia(params.ISOs, params.BootOrder, nil); err != nil {
		return nil, err
	}
	m := &types.VMManifest{
		ISOs: params.ISOs, BootOrder: params.BootOrder,
		ID:          uuid.NewString(),
		Name:        params.Name,
		Image:       params.Image,
		CPUs:        params.CPUs,
		MemoryMiB:   params.MemoryMiB,
		DiskSizeGiB: params.DiskSizeGiB,
		Network:     networkRef,
		Addresses:   params.Addresses,
		Username:    params.Username,
		Password:    params.Password,
		SSHKey:      params.SSHKey,
		Autostart:   params.Autostart,
	}

	exposeInterfaces(m)
	if err := e.vms.Save(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (e *Engine) ListVMs(ctx context.Context) ([]VMView, error) {
	manifests, err := e.vms.List()
	if err != nil {
		return nil, err
	}

	database, err := e.OpenDB(ctx)
	if err != nil {
		return nil, err
	}
	cached, err := database.ListVMStates(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]VMView, 0, len(manifests))
	for _, m := range manifests {
		st := e.driver.Status(m.ID)
		boot := int64(0)
		if st.Phase == vm.PhaseRunning {
			if prev, ok := cached[m.ID]; ok && prev.Phase == string(vm.PhaseRunning) {
				boot = prev.BootTime
			}
			if boot == 0 {
				boot = time.Now().Unix()
			}
		}

		if err := database.SetVMState(ctx, types.VMState{
			ID: m.ID, Phase: string(st.Phase), PID: st.PID, BootTime: boot, SeenAt: time.Now().Unix(),
		}); err != nil {
			return nil, err
		}

		exposeInterfaces(m)
		views = append(views, VMView{Manifest: m, Phase: string(st.Phase), PID: st.PID, BootTime: boot})
	}

	return views, nil
}

func (e *Engine) StartVM(ctx context.Context, ref string) (vm.Status, error) {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return vm.Status{}, err
	}

	if st := e.driver.Status(m.ID); st.Phase == vm.PhaseRunning {
		return st, nil
	}

	st, err := e.driver.StartPreparedContext(ctx, m.ID, func() (vm.Spec, error) {
		current, err := e.vms.Load(m.ID)
		if err != nil {
			return vm.Spec{}, err
		}
		m = current
		if err := e.checkUSBAssignments(ctx, m); err != nil {
			return vm.Spec{}, err
		}
		spec, err := e.prepareSpec(ctx, m)
		if err != nil {
			return vm.Spec{}, err
		}
		log.Ctx(ctx).Info().Msg("Starting QEMU")
		return spec, nil
	})
	if err != nil {
		return vm.Status{}, err
	}

	if err := e.attachAssignedUSB(ctx, m); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Stopping VM after USB assignment failed")
		_ = e.driver.Stop(m.ID, 5*time.Second)
		_ = e.cacheState(ctx, m.ID, vm.Status{Phase: vm.PhaseStopped}, 0)
		return vm.Status{}, err
	}

	_ = e.cacheState(ctx, m.ID, st, time.Now().Unix())
	e.captureInitialPreview(ctx, m.ID)
	return st, nil
}

func (e *Engine) StopVM(ctx context.Context, ref string, timeout time.Duration) error {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Msg("Requesting guest shutdown")
	if err := e.driver.Stop(m.ID, timeout); err != nil {
		return err
	}

	_ = e.cacheState(ctx, m.ID, vm.Status{Phase: vm.PhaseStopped}, 0)
	return nil
}

func (e *Engine) StatusVM(ref string) (*types.VMManifest, vm.Status, error) {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return nil, vm.Status{}, err
	}

	exposeInterfaces(m)
	return m, e.driver.Status(m.ID), nil
}

func (e *Engine) ConsoleVM(ref string) error {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}

	return e.driver.Console(m.ID)
}

func (e *Engine) DeleteVM(ctx context.Context, ref string) error {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}

	lock, err := e.driver.LockContext(ctx, m.ID)
	if err != nil {
		return err
	}
	defer lock.Close()
	m, err = e.vms.Load(m.ID)
	if err != nil {
		return err
	}
	if e.driver.Status(m.ID).Phase == vm.PhaseRunning {
		return fmt.Errorf("vm %s is running; stop it first", m.Name)
	}

	if err := os.RemoveAll(e.paths.VMDiskDir(m.ID)); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(e.paths.VMRunDir(m.ID), "preview.png")); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := e.vms.Delete(m.ID); err != nil {
		return err
	}

	database, err := e.OpenDB(ctx)
	if err != nil {
		return err
	}
	if err := database.DeleteBackupSchedule(ctx, m.ID); err != nil {
		return err
	}
	return database.DeleteVMState(ctx, m.ID)
}

func (e *Engine) Reconcile(ctx context.Context) error {
	manifests, err := e.vms.List()
	if err != nil {
		return err
	}

	database, err := e.OpenDB(ctx)
	if err != nil {
		return err
	}
	cached, err := database.ListVMStates(ctx)
	if err != nil {
		return err
	}

	for _, m := range manifests {
		st := e.driver.Status(m.ID)
		startedAt := int64(0)
		if st.Phase != vm.PhaseRunning && m.Autostart {
			st, err = e.StartVM(ctx, m.ID)
			if err != nil {
				return fmt.Errorf("autostart %s: %w", m.Name, err)
			}
			startedAt = time.Now().Unix()
		}

		boot := startedAt
		if st.Phase == vm.PhaseRunning {
			if prev, ok := cached[m.ID]; boot == 0 && ok && prev.Phase == string(vm.PhaseRunning) {
				boot = prev.BootTime
			}
			if boot == 0 {
				boot = time.Now().Unix()
			}
		}

		if err := database.SetVMState(ctx, types.VMState{
			ID: m.ID, Phase: string(st.Phase), PID: st.PID, BootTime: boot, SeenAt: time.Now().Unix(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) cacheState(ctx context.Context, id string, st vm.Status, boot int64) error {
	database, err := e.OpenDB(ctx)
	if err != nil {
		return err
	}
	return database.SetVMState(ctx, types.VMState{
		ID: id, Phase: string(st.Phase), PID: st.PID, BootTime: boot, SeenAt: time.Now().Unix(),
	})
}

func (e *Engine) prepareSpec(ctx context.Context, m *types.VMManifest) (vm.Spec, error) {
	log.Ctx(ctx).Info().Msg("Resolving VM network")
	spec := vm.Spec{Interfaces: []vm.InterfaceSpec{}}
	interfaces, err := e.normalizedInterfaces(m)
	if err == nil {
		for _, nic := range interfaces {
			network, resolveErr := e.resolveNetwork(nic.Network)
			if resolveErr != nil {
				err = resolveErr
				break
			}
			network.ID = nic.ID
			network.MAC = nic.MAC
			spec.Interfaces = append(spec.Interfaces, network)
		}
	}
	if err != nil {
		return vm.Spec{}, err
	}

	fw, err := firmware.EnsureContext(ctx, e.paths.FirmwareDir())
	if err != nil {
		return vm.Spec{}, err
	}

	diskPath, err := e.prepareVMDisk(ctx, m)
	if err != nil {
		return vm.Spec{}, err
	}
	seedPath := ""
	if m.Image != "" {
		log.Ctx(ctx).Info().Msg("Preparing cloud-init seed")
		seedPath = filepath.Join(e.paths.VMDiskDir(m.ID), "seed.iso")
		if _, err := cloudinit.BuildSeedISO(seedPath, cloudinit.Seed{Hostname: m.Name, Username: m.Username, Password: m.Password, SSHKey: m.SSHKey, NetworkConfig: guestInterfacesConfig(interfaces)}); err != nil {
			return vm.Spec{}, err
		}
	}

	for _, disk := range m.Disks {
		if _, err := uuid.Parse(disk.ID); err != nil {
			return vm.Spec{}, fmt.Errorf("invalid data disk ID: %w", err)
		}

		path := filepath.Join(e.paths.VMDiskDir(m.ID), disk.ID+".qcow2")
		if _, err := os.Stat(path); err != nil {
			return vm.Spec{}, fmt.Errorf("data disk %s unavailable: %w", disk.Name, err)
		}

		spec.Disks = append(spec.Disks, vm.DiskSpec{ID: disk.ID, Path: path})
	}

	for _, id := range m.ISOs {
		media, err := e.GetMedia(id)
		if err != nil {
			return vm.Spec{}, err
		}
		spec.ISOs = append(spec.ISOs, vm.DiskSpec{ID: id, Path: media.Path})
	}
	spec.BootOrder = m.EffectiveBootOrder()
	spec.ID, spec.Name = m.ID, m.Name
	spec.CPUs, spec.MemoryMiB = m.CPUs, m.MemoryMiB
	spec.DiskPath, spec.SeedPath, spec.Firmware = diskPath, seedPath, fw
	return spec, nil
}

func (e *Engine) resolveNetwork(ref string) (vm.InterfaceSpec, error) {
	switch ref {
	case "", "user":
		return vm.InterfaceSpec{Network: vm.NetworkUser}, nil
	case "vmnet-host", "vmnet-shared":
		return vm.InterfaceSpec{Network: vm.NetworkMode(ref)}, nil
	}

	n, err := e.nets.Resolve(ref)
	if err != nil {
		return vm.InterfaceSpec{}, fmt.Errorf("resolve network %q: %w", ref, err)
	}

	switch n.Mode {
	case types.NetworkBridge:
		if err := network.Ensure(e.nets, n, false); err != nil {
			return vm.InterfaceSpec{}, err
		}

		return vm.InterfaceSpec{Network: vm.NetworkBridge, Bridge: n.Device}, nil
	case types.NetworkSwitch:
		return vm.InterfaceSpec{Network: vm.NetworkSwitch, Group: n.Group}, nil
	case types.NetworkUser:
		return vm.InterfaceSpec{Network: vm.NetworkUser}, nil
	case types.NetworkBridged, types.NetworkVmnetBridged:
		return vm.InterfaceSpec{Network: vm.NetworkVmnetBridged, Uplink: n.Uplink}, nil
	default:
		return vm.InterfaceSpec{}, fmt.Errorf("unknown network mode %q", n.Mode)
	}
}

func guestNetworkConfig(mac string, addresses []string) string {
	config := "version: 2\nrenderer: networkd\nethernets:\n  lab:\n    match:\n      macaddress: " + mac + "\n    set-name: lab0\n    optional: true\n"
	if len(addresses) == 0 {
		return config + "    dhcp4: true\n"
	}

	config += "    dhcp4: false\n    addresses:\n"
	for _, address := range addresses {
		config += "      - " + address + "\n"
	}

	return config
}

func (e *Engine) ScreenshotVM(ref string) error {
	manifest, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}

	return e.driver.Screenshot(manifest.ID)
}

func (e *Engine) ShutdownVM(ref string) error {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}
	return e.driver.Shutdown(m.ID)
}

func (e *Engine) ForceStopVM(ctx context.Context, ref string) error {
	m, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}
	if err := e.driver.ForceStop(m.ID); err != nil {
		return err
	}
	_ = e.cacheState(ctx, m.ID, vm.Status{Phase: vm.PhaseStopped}, 0)
	return nil
}

func (e *Engine) prepareVMDisk(ctx context.Context, m *types.VMManifest) (string, error) {
	diskPath := filepath.Join(e.paths.VMDiskDir(m.ID), "disk.qcow2")
	if _, err := os.Stat(diskPath); err == nil {
		return diskPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	base := ""
	if m.Image != "" {
		var err error
		base, err = e.imageBase(ctx, m.Image, true)
		if err != nil {
			return "", err
		}
	}
	if err := image.CreateVMDisk(ctx, base, diskPath, m.DiskSizeGiB); err != nil {
		return "", err
	}
	return diskPath, nil
}
