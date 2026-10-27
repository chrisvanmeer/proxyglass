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
