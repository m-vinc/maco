package tlscert

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSelfSigned(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "tls.crt")
	key := filepath.Join(dir, "tls.key")

	if err := EnsureSelfSigned(crt, key); err != nil {
		t.Fatalf("generate: %v", err)
	}

	if _, err := tls.LoadX509KeyPair(crt, key); err != nil {
		t.Fatalf("load keypair: %v", err)
	}

	info, err := os.Stat(key)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("key mode = %v, want 0o600", info.Mode().Perm())
	}

	before, err := os.ReadFile(crt)
	if err != nil {
		t.Fatal(err)
	}

	if err := EnsureSelfSigned(crt, key); err != nil {
		t.Fatalf("regenerate: %v", err)
	}

	after, err := os.ReadFile(crt)
	if err != nil {
		t.Fatal(err)
	}

	if string(before) != string(after) {
		t.Fatal("EnsureSelfSigned overwrote an existing certificate")
	}
}
