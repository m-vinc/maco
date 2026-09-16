package types

type VMDisk struct {
	ID      string `yaml:"id" json:"id" validate:"required,uuid"`
	Name    string `yaml:"name" json:"name" validate:"required"`
	SizeGiB int    `yaml:"size_gib" json:"size_gib" validate:"min=1"`
}

type VMUSBAssignment struct {
	VendorID  uint16 `yaml:"vendor_id" json:"vendor_id" validate:"required"`
	ProductID uint16 `yaml:"product_id" json:"product_id" validate:"required"`
	Serial    string `yaml:"serial,omitempty" json:"serial,omitempty" binding:"optional"`
	Product   string `yaml:"product,omitempty" json:"product,omitempty" binding:"optional"`
}

type VMInterface struct {
	ID        string   `yaml:"id" json:"id" validate:"required"`
	Network   string   `yaml:"network" json:"network"`
	MAC       string   `yaml:"mac,omitempty" json:"mac,omitempty" validate:"omitempty,mac" binding:"optional"`
	Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty" validate:"dive,cidr" binding:"optional"`
}

type VMManifest struct {
	Interfaces  *[]VMInterface    `yaml:"interfaces,omitempty" json:"interfaces,omitempty" validate:"omitempty,dive" binding:"optional"`
	ISOs        []string          `yaml:"isos,omitempty" json:"isos,omitempty" binding:"optional"`
	BootOrder   []string          `yaml:"boot_order,omitempty" json:"boot_order,omitempty" binding:"optional"`
	Disks       []VMDisk          `yaml:"disks,omitempty" json:"disks,omitempty" validate:"dive" binding:"optional"`
	USB         []VMUSBAssignment `yaml:"usb,omitempty" json:"usb,omitempty" validate:"dive" binding:"optional"`
	ID          string            `yaml:"id" json:"id" validate:"required"`
	Name        string            `yaml:"name" json:"name" validate:"required,hostname_rfc1123"`
	Image       string            `yaml:"image,omitempty" json:"image"`
	CPUs        int               `yaml:"cpus" json:"cpus" validate:"min=1"`
	MemoryMiB   int               `yaml:"memory_mib" json:"memory_mib" validate:"min=64"`
	DiskSizeGiB int               `yaml:"disk_size_gib" json:"disk_size_gib" validate:"min=1"`
	Network     string            `yaml:"network,omitempty" json:"network,omitempty" binding:"optional"`
	Addresses   []string          `yaml:"addresses,omitempty" json:"addresses,omitempty" validate:"dive,cidr" binding:"optional"`
	Username    string            `yaml:"username" json:"username" validate:"required"`
	Password    string            `yaml:"password,omitempty" json:"-"`
	SSHKey      string            `yaml:"ssh_key,omitempty" json:"ssh_key,omitempty" binding:"optional"`
	Autostart   bool              `yaml:"autostart,omitempty" json:"autostart,omitempty" binding:"optional"`
}

type BackupSchedule struct {
	VMID          string `json:"vm_id"`
	Enabled       bool   `json:"enabled"`
	IntervalHours int    `json:"interval_hours"`
	KeepLast      int    `json:"keep_last"`
	MaxAgeDays    int    `json:"max_age_days"`
	LastRunAt     int64  `json:"last_run_at"`
}

type VMState struct {
	ID       string
	Phase    string
	PID      int
	BootTime int64
	SeenAt   int64
}

func (m *VMManifest) EffectiveBootOrder() []string {
	devices := []string{"disk"}
	for _, disk := range m.Disks {
		devices = append(devices, "disk:"+disk.ID)
	}
	for _, id := range m.ISOs {
		devices = append(devices, "iso:"+id)
	}
	preferred := m.BootOrder
	if len(preferred) == 0 {
		preferred = []string{}
		for _, id := range m.ISOs {
			preferred = append(preferred, "iso:"+id)
		}
		preferred = append(preferred, "disk")
	}
	valid := map[string]bool{}
	for _, id := range devices {
		valid[id] = true
	}
	seen := map[string]bool{}
	result := []string{}
	for _, id := range append(append([]string{}, preferred...), devices...) {
		if valid[id] && !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

func (m *VMManifest) EffectiveInterfaces() []VMInterface {
	if m.Interfaces != nil {
		return *m.Interfaces
	}
	return []VMInterface{{ID: "net0", Network: m.Network, Addresses: m.Addresses}}
}
