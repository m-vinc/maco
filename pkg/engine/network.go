package engine

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/m-vinc/maco/pkg/network"
	"github.com/m-vinc/maco/pkg/types"
	"github.com/m-vinc/maco/pkg/vm"
)

type CreateNetworkParams struct {
	Name    string       `json:"name" binding:"optional"`
	Mode    string       `json:"mode" binding:"optional"`
	Uplink  string       `json:"uplink" binding:"optional"`
	Parent  string       `json:"parent" binding:"optional"`
	Tag     int          `json:"tag" binding:"optional"`
	Address string       `json:"address" binding:"optional"`
	Device  string       `json:"device" binding:"optional"`
	Members []string     `json:"members" binding:"optional"`
	VLANs   []types.VLAN `json:"vlans" binding:"optional"`
}

func mcastGroup(id string) string {
	sum := sha256.Sum256([]byte(id))
	port := 20000 + int(binary.BigEndian.Uint16(sum[3:5])%20000)
	return fmt.Sprintf("239.%d.%d.%d:%d", sum[0], sum[1], sum[2], port)
}

func (e *Engine) ListNetworks() ([]*types.NetworkManifest, error) {
	return e.nets.List()
}

func (e *Engine) CreateNetwork(params CreateNetworkParams) (*types.NetworkManifest, error) {
	if params.Name == "user" || strings.HasPrefix(params.Name, "vmnet-") {
		return nil, fmt.Errorf("network name %q is reserved", params.Name)
	}

	if existing, err := e.nets.Resolve(params.Name); err == nil {
		return nil, fmt.Errorf("network %s already exists (%s)", params.Name, existing.ID)
	}

	if params.Mode == "" {
		params.Mode = "bridge"
	}

	if params.Mode != "bridged" && params.Mode != "vmnet-bridged" && params.Uplink != "" {
		return nil, fmt.Errorf("uplink requires vmnet-bridged mode; use members for a native bridge")
	}

	n := &types.NetworkManifest{
		ID:      uuid.NewString(),
		Name:    params.Name,
		Mode:    types.NetworkMode(params.Mode),
		Address: params.Address,
		Device:  params.Device,
		Members: params.Members,
		VLANs:   params.VLANs,
		Parent:  params.Parent,
		Tag:     params.Tag,
	}

	switch n.Mode {
	case types.NetworkSwitch:
		n.Group = mcastGroup(n.ID)
	case types.NetworkBridged, types.NetworkVmnetBridged:
		if params.Uplink == "" {
			return nil, fmt.Errorf("vmnet-bridged network requires an uplink (e.g. en0)")
		}

		n.Uplink = params.Uplink
	case types.NetworkVLAN:
		if params.Parent == "" || params.Tag < 1 || params.Tag > 4094 {
			return nil, fmt.Errorf("vlan network requires a parent interface and a tag between 1 and 4094")
		}
	}

	if err := e.nets.Save(n); err != nil {
		return nil, err
	}

	if err := e.ensureNetwork(e.nets, n, false); err != nil {
		current, loadErr := e.nets.Load(n.ID)
		if loadErr != nil {
			return nil, fmt.Errorf("apply network: %w (reload for rollback: %v)", err, loadErr)
		}
		if cleanupErr := network.Destroy(current); cleanupErr != nil {
			return nil, fmt.Errorf("apply network: %w (rollback failed; retained manifest %s: %v)", err, n.ID, cleanupErr)
		}
		if cleanupErr := e.nets.Delete(n.ID); cleanupErr != nil {
			return nil, fmt.Errorf("apply network: %w (remove failed manifest: %v)", err, cleanupErr)
		}
		return nil, fmt.Errorf("apply network: %w", err)
	}
	return n, nil
}

func (e *Engine) ApplyNetworks(dryRun bool) error {
	return network.Reconcile(e.nets, dryRun)
}

func (e *Engine) ApplyNetwork(ref string, dryRun bool) error {
	n, err := e.nets.Resolve(ref)
	if err != nil {
		return err
	}

	return network.Ensure(e.nets, n, dryRun)
}

func (e *Engine) DestroyNetwork(ref string) error {
	lock, err := e.nets.Lock()
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()

	n, err := e.nets.Resolve(ref)
	if err != nil {
		return err
	}

	manifests, err := e.vms.List()
	if err != nil {
		return err
	}

	for _, m := range manifests {
		if e.driver.Status(m.ID).Phase == vm.PhaseRunning {
			refs, err := e.driver.PendingNetworkReferences(m.ID)
			if err != nil {
				return fmt.Errorf("inspect pending VM network change: %w", err)
			}
			if slices.Contains(refs, n.ID) || slices.Contains(refs, n.Name) {
				return fmt.Errorf("network %s is involved in an unfinished change on VM %s; stop it first", n.Name, m.Name)
			}
		}
		for _, nic := range m.EffectiveInterfaces() {
			if (nic.Network == n.ID || nic.Network == n.Name) && e.driver.Status(m.ID).Phase == vm.PhaseRunning {
				return fmt.Errorf("network %s is used by running VM %s; stop it first", n.Name, m.Name)
			}
		}
	}

	if n.Mode == types.NetworkVLAN && n.Device != "" {
		networks, err := e.nets.List()
		if err != nil {
			return err
		}

		for _, other := range networks {
			if other.ID == n.ID {
				continue
			}

			if slices.Contains(other.Members, n.Device) || slices.Contains(other.AppliedMembers, n.Device) {
				return fmt.Errorf("VLAN %s is attached to bridge %s; remove it from the bridge first", n.Name, other.Name)
			}
		}
	}

	if err := network.Destroy(n); err != nil {
		return err
	}

	return e.nets.Delete(n.ID)
}

func (e *Engine) UpdateNetwork(ref string, params CreateNetworkParams) (*types.NetworkManifest, error) {
	lock, err := e.nets.Lock()
	if err != nil {
		return nil, err
	}
	n, err := func() (*types.NetworkManifest, error) {
		n, err := e.nets.Resolve(ref)
		if err != nil {
			return nil, err
		}
		if params.Name != n.Name || types.NetworkMode(params.Mode) != n.Mode {
			return nil, fmt.Errorf("network name and mode must not change")
		}
		if params.Device != "" && params.Device != n.Device {
			return nil, fmt.Errorf("applied device must not change")
		}
		manifests, err := e.vms.List()
		if err != nil {
			return nil, err
		}
		for _, m := range manifests {
			if e.driver.Status(m.ID).Phase != vm.PhaseRunning {
				continue
			}
			refs, err := e.driver.PendingNetworkReferences(m.ID)
			if err != nil {
				return nil, fmt.Errorf("inspect pending VM network change: %w", err)
			}
			if slices.Contains(refs, n.ID) || slices.Contains(refs, n.Name) {
				return nil, fmt.Errorf("network %s is involved in an unfinished change on VM %s; stop it first", n.Name, m.Name)
			}
			for _, nic := range m.EffectiveInterfaces() {
				if nic.Network == n.ID || nic.Network == n.Name {
					return nil, fmt.Errorf("network %s is used by running VM %s; stop it first", n.Name, m.Name)
				}
			}
		}

		if n.Mode == types.NetworkVLAN && (n.Parent != params.Parent || n.Tag != params.Tag) && n.Device != "" {
			if !n.Owned {
				return nil, fmt.Errorf("cannot retag a borrowed VLAN interface")
			}
			networks, err := e.nets.List()
			if err != nil {
				return nil, err
			}
			for _, other := range networks {
				if other.ID != n.ID && (slices.Contains(other.Members, n.Device) || slices.Contains(other.AppliedMembers, n.Device)) {
					return nil, fmt.Errorf("VLAN %s is attached to bridge %s; remove it from the bridge first", n.Name, other.Name)
				}
			}
		}
		n.Address, n.Uplink, n.Parent, n.Tag = params.Address, params.Uplink, params.Parent, params.Tag
		n.Members = slices.Clone(params.Members)
		n.VLANs = make([]types.VLAN, len(params.VLANs))
		for i, vlan := range params.VLANs {
			n.VLANs[i] = types.VLAN{Parent: vlan.Parent, Tag: vlan.Tag}
		}
		if err := e.nets.Save(n); err != nil {
			return nil, err
		}
		return n, nil
	}()
	_ = lock.Close()
	if err != nil {
		return nil, err
	}
	if err := e.ensureNetwork(e.nets, n, false); err != nil {
		return nil, fmt.Errorf("apply network (configuration saved; retry to reconcile): %w", err)
	}
	return n, nil
}
