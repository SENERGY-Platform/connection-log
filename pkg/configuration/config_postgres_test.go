package configuration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLoad_PostgresDefaultsAndEnvNames(t *testing.T) {
	for _, name := range []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PW", "POSTGRES_DB"} {
		t.Setenv(name, "")
	}
	cfg, err := Load("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostgresDb != "connectionlog" || cfg.PostgresUser != "connectionlog" || cfg.PostgresPort != 5432 {
		t.Errorf("defaults: db=%q user=%q port=%d", cfg.PostgresDb, cfg.PostgresUser, cfg.PostgresPort)
	}

	t.Setenv("POSTGRES_HOST", "timescale-rw.timescale.svc.cluster.local")
	t.Setenv("POSTGRES_PORT", "6543")
	t.Setenv("POSTGRES_USER", "u")
	t.Setenv("POSTGRES_PW", "s3cr3t-pg-pw")
	t.Setenv("POSTGRES_DB", "other")
	var out string
	out = captureStdout(t, func() {
		cfg, err = Load("../../config.json")
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostgresHost != "timescale-rw.timescale.svc.cluster.local" || cfg.PostgresPort != 6543 || cfg.PostgresUser != "u" || cfg.PostgresPw != "s3cr3t-pg-pw" || cfg.PostgresDb != "other" {
		t.Errorf("env not applied: %+v", cfg.masked())
	}
	// secrets are neither echoed on load nor printed or marshalled afterwards
	b, _ := json.Marshal(cfg)
	for name, s := range map[string]string{"stdout": out, "json": string(b), "string": cfg.String()} {
		if strings.Contains(s, "s3cr3t-pg-pw") {
			t.Errorf("%s leaks the postgres password: %s", name, s)
		}
	}
	if strings.Contains(string(b), `"PostgresUser":"u"`) {
		t.Errorf("postgres user is not masked: %s", b)
	}
}
