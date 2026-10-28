package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRefreshDefaultsToFiveAndAllowsZero(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want int
	}{{"implicit", "server:\n  listen: ':0'\n", 5}, {"disabled", "server:\n  listen: ':0'\n  default_refresh: 0\n", 0}, {"configured", "server:\n  listen: ':0'\n  default_refresh: 17\n", 17}} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Server.DefaultRefresh != tc.want {
				t.Fatalf("got refresh %d, want %d", cfg.Server.DefaultRefresh, tc.want)
			}
		})
	}
}

func TestLoadTopologyFrontendBackendMappings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	contents := `targets:
  - name: edge-a
    url: http://edge-a/stats;csv
topology:
  - target: edge-a
    frontend: public
    backends: [web, api]
`
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Topology) != 1 || cfg.Topology[0].Target != "edge-a" || cfg.Topology[0].Frontend != "public" {
		t.Fatalf("topology mapping not loaded: %#v", cfg.Topology)
	}
	if got := cfg.Topology[0].Backends; len(got) != 2 || got[0] != "web" || got[1] != "api" {
		t.Fatalf("topology backends = %#v, want [web api]", got)
	}
}
