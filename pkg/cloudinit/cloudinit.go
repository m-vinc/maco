package cloudinit

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
)

type Seed struct {
	Hostname      string
	Username      string
	Password      string
	SSHKey        string
	NetworkConfig string
	UserData      string
}

type cloudUser struct {
	Name         string   `yaml:"name"`
	Sudo         string   `yaml:"sudo"`
	Shell        string   `yaml:"shell"`
	LockPassword bool     `yaml:"lock_passwd"`
	SSHKeys      []string `yaml:"ssh_authorized_keys,omitempty"`
}
type cloudPassword struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Type     string `yaml:"type"`
}
type passwordConfig struct {
	Expire bool            `yaml:"expire"`
	Users  []cloudPassword `yaml:"users"`
}
type cloudConfig struct {
	Hostname        string          `yaml:"hostname"`
	Users           []cloudUser     `yaml:"users"`
	SSHPasswordAuth bool            `yaml:"ssh_pwauth,omitempty"`
	Passwords       *passwordConfig `yaml:"chpasswd,omitempty"`
	Packages        []string        `yaml:"packages"`
	Commands        [][]string      `yaml:"runcmd"`
}

func (s Seed) defaultUserData() string {
	user := s.Username
	if user == "" {
		user = "maco"
	}
	account := cloudUser{Name: user, Sudo: "ALL=(ALL) NOPASSWD:ALL", Shell: "/bin/bash"}
	if s.SSHKey != "" {
		account.SSHKeys = []string{s.SSHKey}
	}
	config := cloudConfig{Hostname: s.Hostname, Users: []cloudUser{account}, Packages: []string{"qemu-guest-agent"}, Commands: [][]string{{"systemctl", "enable", "--now", "qemu-guest-agent"}}}
	if s.Password != "" {
		config.SSHPasswordAuth = true
		config.Passwords = &passwordConfig{Users: []cloudPassword{{Name: user, Password: s.Password, Type: "text"}}}
	}
	data, _ := yaml.Marshal(config)
	return "#cloud-config\n" + string(data)
}
func (s Seed) metaData() string {
	data, _ := yaml.Marshal(struct {
		InstanceID string `yaml:"instance-id"`
		Hostname   string `yaml:"local-hostname"`
	}{s.Hostname, s.Hostname})
	return string(data)
}

func BuildSeedISO(isoPath string, seed Seed) (string, error) {
	if err := os.MkdirAll(filepath.Dir(isoPath), 0o700); err != nil {
		return "", err
	}

	staging, err := os.MkdirTemp("", "maco-cidata-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(staging) }()

	userData := seed.UserData
	if userData == "" {
		userData = seed.defaultUserData()
	}

	if err := os.WriteFile(filepath.Join(staging, "user-data"), []byte(userData), 0o600); err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(staging, "meta-data"), []byte(seed.metaData()), 0o600); err != nil {
		return "", err
	}

	if seed.NetworkConfig != "" {
		if err := os.WriteFile(filepath.Join(staging, "network-config"), []byte(seed.NetworkConfig), 0o600); err != nil {
			return "", err
		}
	}

	if err := makeHybridISO(isoPath, staging); err != nil {
		return "", err
	}

	return isoPath, nil
}

func makeHybridISO(isoPath, srcDir string) error {
	hdiutil, err := exec.LookPath("hdiutil")
	if err != nil {
		return fmt.Errorf("hdiutil not found: %w", err)
	}

	_ = os.Remove(isoPath)

	args := []string{
		"makehybrid",
		"-o", isoPath,
		"-iso", "-joliet",
		"-default-volume-name", "CIDATA",
		srcDir,
	}

	out, err := exec.Command(hdiutil, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("hdiutil makehybrid: %w: %s", err, out)
	}

	return nil
}
