package framework

import (
	"os"
	"testing"
)

func TestResolveKubeconfigUsesEnvironmentFile(t *testing.T) {
	kubeconfig := []byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://backplane.example.test
  name: hosted-cluster
contexts:
- context:
    cluster: hosted-cluster
    user: backplane
  name: hosted-cluster
current-context: hosted-cluster
users:
- name: backplane
  user:
    token: test-token
`)

	path := t.TempDir() + "/kubeconfig"
	if err := os.WriteFile(path, kubeconfig, 0o600); err != nil {
		t.Fatalf("writing kubeconfig: %v", err)
	}
	t.Setenv("KUBECONFIG", path)

	cfg, err := resolveKubeconfig("KUBECONFIG", nil, "unused-cluster-id")
	if err != nil {
		t.Fatalf("resolveKubeconfig() returned an error: %v", err)
	}
	if cfg.Host != "https://backplane.example.test" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "https://backplane.example.test")
	}
	if cfg.BearerToken != "test-token" {
		t.Fatalf("BearerToken = %q, want %q", cfg.BearerToken, "test-token")
	}
}

// writeTestKubeconfig writes a minimal valid kubeconfig to a temp file and returns its path.
func writeTestKubeconfig(t *testing.T) string {
	t.Helper()
	kubeconfig := []byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://backplane.example.test
  name: hosted-cluster
contexts:
- context:
    cluster: hosted-cluster
    user: backplane
  name: hosted-cluster
current-context: hosted-cluster
users:
- name: backplane
  user:
    token: test-token
`)
	path := t.TempDir() + "/bp-kubeconfig"
	if err := os.WriteFile(path, kubeconfig, 0o600); err != nil {
		t.Fatalf("writing kubeconfig: %v", err)
	}
	return path
}

func TestInitBPClientsLoadsBackplaneKubeconfig(t *testing.T) {
	t.Setenv("BP_KUBECONFIG", writeTestKubeconfig(t))

	tc := &TestContext{}
	if err := tc.InitBPClients(); err != nil {
		t.Fatalf("InitBPClients() returned an error: %v", err)
	}
	if tc.BPKubeClient() == nil {
		t.Fatal("BPKubeClient() = nil, want a client")
	}
	if tc.BPDynamicClient() == nil {
		t.Fatal("BPDynamicClient() = nil, want a client")
	}
	if tc.bpRestConfig == nil {
		t.Fatal("bpRestConfig = nil, want a rest config")
	}
	if tc.bpRestConfig.Host != "https://backplane.example.test" {
		t.Fatalf("bpRestConfig.Host = %q, want %q", tc.bpRestConfig.Host, "https://backplane.example.test")
	}
}

func TestInitBPClientsErrorsWhenEnvUnset(t *testing.T) {
	// An empty value is indistinguishable from unset for os.Getenv and exercises the same guard.
	t.Setenv("BP_KUBECONFIG", "")

	tc := &TestContext{}
	if err := tc.InitBPClients(); err == nil {
		t.Fatal("InitBPClients() = nil error, want an error when BP_KUBECONFIG is unset")
	}
	if tc.BPKubeClient() != nil || tc.BPDynamicClient() != nil {
		t.Fatal("clients should remain nil when InitBPClients fails")
	}
}

func TestInitBPClientsErrorsOnMissingFile(t *testing.T) {
	t.Setenv("BP_KUBECONFIG", t.TempDir()+"/does-not-exist")

	tc := &TestContext{}
	if err := tc.InitBPClients(); err == nil {
		t.Fatal("InitBPClients() = nil error, want an error for a missing kubeconfig file")
	}
}

func TestHasBPAccess(t *testing.T) {
	tc := &TestContext{}

	t.Setenv("BP_KUBECONFIG", "/tmp/bp-kubeconfig")
	if !tc.HasBPAccess() {
		t.Error("HasBPAccess() = false, want true when BP_KUBECONFIG is set")
	}

	t.Setenv("BP_KUBECONFIG", "")
	if tc.HasBPAccess() {
		t.Error("HasBPAccess() = true, want false when BP_KUBECONFIG is empty")
	}
}
