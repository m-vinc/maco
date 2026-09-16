package engine

import (
	"context"

	"github.com/m-vinc/maco/pkg/config"
	"github.com/m-vinc/maco/pkg/db"
	"github.com/m-vinc/maco/pkg/manifest"
	"github.com/m-vinc/maco/pkg/network"
	"github.com/m-vinc/maco/pkg/types"
	"github.com/m-vinc/maco/pkg/usb"
	"github.com/m-vinc/maco/pkg/vm"
)

type Engine struct {
	ensureNetwork func(*network.Store, *types.NetworkManifest, bool) error
	usbDevices    func() ([]usb.Device, error)
	usbClaimsDir  string
	paths         *config.Paths
	vms           *manifest.Store
	nets          *network.Store
	driver        *vm.Driver
}

func New(paths *config.Paths) *Engine {
	return &Engine{
		paths:         paths,
		ensureNetwork: network.Ensure,
		vms:           manifest.NewStore(paths.VMsDir()),
		nets:          network.NewStore(paths.NetworksDir()),
		driver:        vm.NewDriver(paths.RunDir()),
	}
}

func (e *Engine) Paths() *config.Paths     { return e.paths }
func (e *Engine) VMStore() *manifest.Store { return e.vms }
func (e *Engine) NetStore() *network.Store { return e.nets }
func (e *Engine) Driver() *vm.Driver       { return e.driver }

func (e *Engine) OpenDB(ctx context.Context) (*db.DB, error) {
	return db.Shared(ctx, e.paths.DBPath())
}
