package vm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func serveMockAgent(t *testing.T, sock string) {
	t.Helper()
	listener, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	replies := map[string]string{
		"guest-info":                   `{"return":{"version":"8.2.2"}}`,
		"guest-get-host-name":          `{"return":{"host-name":"testhost"}}`,
		"guest-get-osinfo":             `{"return":{"id":"ubuntu","pretty-name":"Ubuntu 24.04 LTS","kernel-release":"6.8.0-40-generic","machine":"aarch64"}}`,
		"guest-network-get-interfaces": `{"return":[{"name":"lo","ip-addresses":[]},{"name":"enp0s1","hardware-address":"02:aa:bb:cc:dd:ee","ip-addresses":[{"ip-address":"10.0.0.5","ip-address-type":"ipv4","prefix":24}]}]}`,
		"guest-get-fsinfo":             `{"return":[{"name":"vda1","mountpoint":"/","type":"ext4","total-bytes":1000,"used-bytes":400}]}`,
		"guest-fsfreeze-freeze":        `{"return":2}`,
		"guest-fsfreeze-thaw":          `{"return":2}`,
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			var cmd struct {
				Execute   string `json:"execute"`
				Arguments struct {
					ID json.RawMessage `json:"id"`
				} `json:"arguments"`
			}
			if json.Unmarshal(bytes.TrimLeft(line, "\xff"), &cmd) != nil {
				return
			}

			if cmd.Execute == "guest-sync-delimited" {
				_, _ = conn.Write([]byte{0xff})
				_, _ = conn.Write([]byte(`{"return":` + string(cmd.Arguments.ID) + "}\n"))
				continue
			}
			if reply, ok := replies[cmd.Execute]; ok {
				_, _ = conn.Write([]byte(reply + "\n"))
			}
		}
	}()
}

func TestGuestAgentClient(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "qga.sock")
	serveMockAgent(t, sock)

	client, err := dialQGA(sock, 2*time.Second)
	if err != nil {
		t.Fatalf("dial/sync: %v", err)
	}
	defer func() { _ = client.close() }()

	var base struct {
		Version string `json:"version"`
	}
	if err := client.execute("guest-info", nil, &base); err != nil || base.Version != "8.2.2" {
		t.Fatalf("guest-info: %v version=%q", err, base.Version)
	}

	if host := client.hostname(); host != "testhost" {
		t.Fatalf("hostname %q", host)
	}

	os := client.osInfo()
	if os == nil || os.PrettyName != "Ubuntu 24.04 LTS" || os.KernelRelease != "6.8.0-40-generic" || os.Machine != "aarch64" {
		t.Fatalf("osinfo %+v", os)
	}

	ifaces := client.interfaces()
	if len(ifaces) != 2 || ifaces[1].Name != "enp0s1" || len(ifaces[1].IPAddresses) != 1 ||
		ifaces[1].IPAddresses[0].Address != "10.0.0.5" || ifaces[1].IPAddresses[0].Prefix != 24 {
		t.Fatalf("interfaces %+v", ifaces)
	}

	fs := client.filesystems()
	if len(fs) != 1 || fs[0].Mountpoint != "/" || fs[0].Type != "ext4" || fs[0].UsedBytes != 400 || fs[0].TotalBytes != 1000 {
		t.Fatalf("filesystems %+v", fs)
	}

	if err := client.execute("guest-fsfreeze-freeze", nil, nil); err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if err := client.execute("guest-fsfreeze-thaw", nil, nil); err != nil {
		t.Fatalf("thaw: %v", err)
	}
}
