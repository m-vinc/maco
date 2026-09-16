package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/m-vinc/maco/pkg/image"
	"github.com/m-vinc/maco/pkg/types"
	"github.com/m-vinc/maco/pkg/vm"
)

type DiskParams struct {
	ImageID string `json:"image_id,omitempty" binding:"optional"`
	ID      string `json:"id" binding:"optional"`
	Name    string `json:"name" binding:"optional"`
	SizeGiB int    `json:"size_gib" binding:"optional"`
}

func (e *Engine) ManageDisk(ctx context.Context, ref, action string, params DiskParams) error {
	manifest, err := e.vms.Resolve(ref)
	if err != nil {
		return err
	}

	lock, err := e.driver.LockContext(ctx, manifest.ID)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	manifest, err = e.vms.Load(manifest.ID)
	if err != nil {
		return err
	}

	if action == "vm.disk.add" {
		return e.addDataDisk(ctx, manifest, params)
	}

	for index, disk := range manifest.Disks {
		if disk.ID != params.ID {
			continue
		}

		if _, err := uuid.Parse(disk.ID); err != nil {
			return fmt.Errorf("invalid disk ID: %w", err)
		}

		path := filepath.Join(e.paths.VMDiskDir(manifest.ID), disk.ID+".qcow2")
		switch action {
		case "vm.disk.grow":
			if params.SizeGiB <= disk.SizeGiB {
				return fmt.Errorf("new capacity must be larger than %d GiB", disk.SizeGiB)
			}

			if err := e.resizeDisk(ctx, manifest.ID, path, params.SizeGiB); err != nil {
				return err
			}

			manifest.Disks[index].SizeGiB = params.SizeGiB
			return e.vms.Save(manifest)
		case "vm.disk.remove":
			if e.driver.Status(manifest.ID).Phase == vm.PhaseRunning {
				return fmt.Errorf("stop the VM before removing disks")
			}

			removed := path + ".removed"
			if err := os.Rename(path, removed); err != nil {
				return err
			}

			manifest.Disks = append(manifest.Disks[:index], manifest.Disks[index+1:]...)
			order := manifest.BootOrder[:0]
			for _, device := range manifest.BootOrder {
				if device != "disk:"+disk.ID {
					order = append(order, device)
				}
			}
			manifest.BootOrder = order
			if err := e.vms.Save(manifest); err != nil {
				if restoreErr := os.Rename(removed, path); restoreErr != nil {
					return fmt.Errorf("save disk removal: %w; restore disk: %v", err, restoreErr)
				}

				return err
			}

			return os.Remove(removed)
		default:
			return fmt.Errorf("unknown disk action")
		}
	}

	return fmt.Errorf("data disk not found; the boot disk cannot be removed")
}

func (e *Engine) addDataDisk(ctx context.Context, manifest *types.VMManifest, params DiskParams) error {
	if params.SizeGiB < 1 || strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("disk name and capacity of at least 1 GiB are required")
	}

	disk := types.VMDisk{ID: uuid.NewString(), Name: strings.TrimSpace(params.Name), SizeGiB: params.SizeGiB}
	path := filepath.Join(e.paths.VMDiskDir(manifest.ID), disk.ID+".qcow2")

	base := ""
	if params.ImageID != "" {
		mediaLock, err := e.lockMedia(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = mediaLock.Close() }()

		media, err := e.GetMedia(params.ImageID)
		if err != nil {
			return err
		}
		if media.Kind != "image" || params.SizeGiB < media.SizeGiB {
			return fmt.Errorf("select a disk image and sufficient capacity")
		}
		base = media.Path
	}

	if err := image.CreateVMDisk(ctx, base, path, disk.SizeGiB); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("create data disk: %w", err)
	}

	manifest.Disks = append(manifest.Disks, disk)
	if err := e.vms.Save(manifest); err != nil {
		_ = os.Remove(path)
		return err
	}

	if e.driver.Status(manifest.ID).Phase == vm.PhaseRunning {
		if err := e.driver.AttachDisk(manifest.ID, disk.ID, path); err != nil {
			manifest.Disks = manifest.Disks[:len(manifest.Disks)-1]
			if saveErr := e.vms.Save(manifest); saveErr != nil {
				return fmt.Errorf("attach disk: %w; restore manifest: %v", err, saveErr)
			}

			if !errors.Is(err, vm.ErrDiskBackendOpen) {
				_ = os.Remove(path)
			}
			return err
		}
	}

	return nil
}
