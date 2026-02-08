package pki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPKIManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pki-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mgr := NewManager(tempDir, "192.168.1.10:6443", "cluster.local", "10.96.0.0/12")

	// 1. Ensure CAs
	if err := mgr.EnsureCAs(); err != nil {
		t.Fatalf("EnsureCAs failed: %v", err)
	}

	// Check CA files
	files := []string{
		"ca.crt", "ca.key",
		"etcd/ca.crt", "etcd/ca.key",
		"front-proxy-ca.crt", "front-proxy-ca.key",
		"sa.key", "sa.pub",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(tempDir, f)); err != nil {
			t.Errorf("File %s missing", f)
		}
	}

	// 2. Create Master Certs
	if err := mgr.CreateMasterCerts("master-1", "192.168.1.10"); err != nil {
		t.Fatalf("CreateMasterCerts failed: %v", err)
	}

	// Check Master Certs
	masterFiles := []string{
		"apiserver.crt", "apiserver.key",
		"apiserver-kubelet-client.crt", "apiserver-kubelet-client.key",
		"front-proxy-client.crt", "front-proxy-client.key",
		"etcd/server.crt", "etcd/server.key",
		"etcd/peer.crt", "etcd/peer.key",
		"etcd/healthcheck-client.crt", "etcd/healthcheck-client.key",
		"apiserver-etcd-client.crt", "apiserver-etcd-client.key",
		"admin.conf", "controller-manager.conf", "scheduler.conf",
	}
	for _, f := range masterFiles {
		if _, err := os.Stat(filepath.Join(tempDir, f)); err != nil {
			t.Errorf("File %s missing", f)
		}
	}

	// 3. Create Worker Certs
	if err := mgr.CreateWorkerCerts("worker-1"); err != nil {
		t.Fatalf("CreateWorkerCerts failed: %v", err)
	}

	// Check Worker Files
	workerFiles := []string{
		"kubelet-worker-1.conf",
	}
	for _, f := range workerFiles {
		if _, err := os.Stat(filepath.Join(tempDir, f)); err != nil {
			t.Errorf("File %s missing", f)
		}
	}
}
